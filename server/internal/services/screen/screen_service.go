package screen

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"
	_ "time/tzdata" // the runtime image may not have /usr/share/zoneinfo

	mediaModel "server/internal/models/media"
	scheduledModel "server/internal/models/scheduled_items"
	settingsModel "server/internal/models/screen_settings"
	widgetModel "server/internal/models/screen_widgets"
	carepartnerModel "server/internal/models/user/carepartner"
	scheduled "server/internal/services/scheduled_items"
	settings "server/internal/services/screen_settings"
	widgets "server/internal/services/screen_widgets"
	newsSvc "server/internal/services/widgets/news"
	"server/internal/services/widgets/weather"
)

var (
	ErrInvalidCarepartner = errors.New("invalid carepartner_id")
	ErrCarepartnerMissing = errors.New("carepartner not found")
)

// Outside data the screen needs; the widget services implement these

type WeatherLookup interface {
	GetWeatherForZip(zip string) (*weather.WeatherOverviewResponse, error)
}

type SportsLookup interface {
	GetTeamGameFor(league, team string) (*newsSvc.GameSummaryResponse, error)
}

type NewsLookup interface {
	GetGoodNews() (*newsSvc.NewsAPIResponse, error)
}

type CalendarSource interface {
	Calendar(carepartnerID int, from, to string) ([]scheduledModel.Occurrence, error)
}

type ScreenService struct {
	DB       *sql.DB
	Weather  WeatherLookup
	Sports   SportsLookup
	News     NewsLookup
	Calendar CalendarSource
	Now      func() time.Time
	cache    *lookupCache
}

func NewScreenService(db *sql.DB, weather WeatherLookup, sports SportsLookup, news NewsLookup, calendar CalendarSource) *ScreenService {
	return &ScreenService{
		DB: db, Weather: weather, Sports: sports, News: news, Calendar: calendar,
		Now:   time.Now,
		cache: newLookupCache(time.Now),
	}
}

func (s *ScreenService) lookups() *lookupCache {
	if s.cache == nil {
		s.cache = newLookupCache(s.Now)
	}
	return s.cache
}

/*
GetScreen assembles the carepartner's screen: the current preset's regions with each
slot's widget and its data, awake hours, and today's and tomorrow's schedule.
*/
func (s *ScreenService) GetScreen(carepartnerID int) (*Screen, error) {
	if carepartnerID <= 0 {
		return nil, ErrInvalidCarepartner
	}

	cp, err := carepartnerModel.GetCarepartnerByID(s.DB, carepartnerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get carepartner: %w", err)
	}
	if cp == nil {
		return nil, ErrCarepartnerMissing
	}
	loc, err := time.LoadLocation(cp.Timezone)
	if err != nil {
		return nil, fmt.Errorf("carepartner %d has unknown timezone %q: %w", carepartnerID, cp.Timezone, err)
	}
	now := s.Now().In(loc)

	ss, err := settingsModel.GetScreenSettings(s.DB, carepartnerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get screen settings: %w", err)
	}
	if ss == nil {
		return nil, ErrCarepartnerMissing
	}
	preset, ok := settings.PresetByName(ss.LayoutPreset)
	if !ok {
		return nil, fmt.Errorf("carepartner %d has unknown layout preset %q", carepartnerID, ss.LayoutPreset)
	}

	stored, err := widgetModel.ListScreenWidgets(s.DB, carepartnerID)
	if err != nil {
		return nil, fmt.Errorf("failed to list widgets: %w", err)
	}
	regions, err := s.fillRegions(preset, stored)
	if err != nil {
		return nil, err
	}

	today := now.Format(scheduled.DateLayout)
	tomorrow := now.AddDate(0, 0, 1).Format(scheduled.DateLayout)
	schedule, err := s.Calendar.Calendar(carepartnerID, today, tomorrow)
	if err != nil {
		return nil, fmt.Errorf("failed to get schedule: %w", err)
	}

	clock := now.Format("15:04")
	return &Screen{
		CarepartnerID: carepartnerID,
		GeneratedAt:   now.Format(time.RFC3339),
		Timezone:      cp.Timezone,
		Today:         today,
		LayoutPreset:  preset.Name,
		AwakeHours: AwakeHours{
			Wake:       ss.WakeTime,
			Bed:        ss.BedTime,
			IsAwakeNow: clock >= ss.WakeTime && clock < ss.BedTime,
		},
		Regions:  regions,
		Schedule: schedule,
	}, nil
}

/*
fillRegions puts each stored widget in its region and slot of the preset (leaving out
widgets whose slot the preset doesn't have) and fills in their photos and data. Outside
lookups run in parallel.
*/
func (s *ScreenService) fillRegions(preset settings.Preset, stored []widgetModel.ScreenWidget) ([]Region, error) {
	at := map[string]widgetModel.ScreenWidget{}
	for _, w := range stored {
		at[fmt.Sprintf("%s/%d", w.Region, w.Slot)] = w
	}

	regions := make([]Region, len(preset.Regions))
	var wg sync.WaitGroup
	for i, r := range preset.Regions {
		regions[i] = Region{Name: r.Name, Slots: make([]*Widget, r.Slots)}
		for slot := 0; slot < r.Slots; slot++ {
			w, ok := at[fmt.Sprintf("%s/%d", r.Name, slot)]
			if !ok {
				continue
			}
			out := &Widget{ID: w.ID, Type: w.Type, Config: w.Config}
			if w.PhotoMediaID != nil {
				photo, err := mediaModel.GetMediaByID(s.DB, *w.PhotoMediaID)
				if err != nil {
					return nil, fmt.Errorf("failed to get widget photo: %w", err)
				}
				if photo != nil {
					out.Photo = &Photo{ID: photo.ID, DisplayName: photo.DisplayName, URL: photo.URL}
				}
			}
			regions[i].Slots[slot] = out

			wg.Add(1)
			go func() {
				defer wg.Done()
				s.fillData(out)
			}()
		}
	}
	wg.Wait()
	return regions, nil
}

// fillData sets a widget's Data, or its Error if an outside lookup fails
func (s *ScreenService) fillData(w *Widget) {
	var data any
	var err error

	switch w.Type {
	case widgets.TypePhoto:
		var c widgets.PhotoConfig
		_ = json.Unmarshal(w.Config, &c)
		caption := c.Caption
		if caption == "" && w.Photo != nil {
			caption = w.Photo.DisplayName
		}
		data = PhotoData{Caption: caption}

	case widgets.TypeWeather:
		var c widgets.WeatherConfig
		_ = json.Unmarshal(w.Config, &c)
		data, err = s.lookups().get("weather:"+c.Zip, func() (any, error) {
			return s.Weather.GetWeatherForZip(c.Zip)
		})

	case widgets.TypeSports:
		var c widgets.SportsConfig
		_ = json.Unmarshal(w.Config, &c)
		data, err = s.lookups().get("sports:"+c.League+":"+c.Team, func() (any, error) {
			game, err := s.Sports.GetTeamGameFor(c.League, c.Team)
			if err != nil {
				return nil, err
			}
			team, _ := newsSvc.FindTeam(c.League, c.Team)
			return SportsData{
				League:           c.League,
				TeamName:         team.Name,
				Result:           game.Result,
				OpponentTeamName: game.OpponentTeamName,
				TeamScore:        game.UserScore,
				OpponentScore:    game.OpponentScore,
				Message:          game.Message,
			}, nil
		})

	case widgets.TypeNews:
		data, err = s.lookups().get("news", func() (any, error) {
			news, err := s.News.GetGoodNews()
			if err != nil {
				return nil, err
			}
			out := NewsData{Articles: []NewsArticle{}}
			for _, a := range news.Articles {
				if len(out.Articles) == MaxNewsArticles {
					break
				}
				out.Articles = append(out.Articles, NewsArticle{Title: a.Title, Description: a.Description, ImageURL: a.URLToImage})
			}
			return out, nil
		})

	default:
		// message and reminder widgets: everything is in Config
		return
	}

	if err != nil {
		log.Printf("screen: %s widget %d lookup failed: %v", w.Type, w.ID, err)
		code := ErrorUpstreamUnavailable
		w.Error = &code
		return
	}
	w.Data = data
}
