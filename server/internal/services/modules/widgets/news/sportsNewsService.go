package news

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"time"

	"server/internal/models/modules/widgets"
)

type SportsNewsWidgetService struct {
	db     *sql.DB
	client *http.Client
}

type SportAPIResponse struct {
	Events []struct {
		Competitions []struct {
			Date        string `json:"date"`
			Competitors []struct {
				HomeAway bool `json:"homeAway"`
				Score    struct {
					Value float64 `json:"value"`
				} `json:"score"`
				Team struct {
					DisplayName  string `json:"displayName"`
					Abbreviation string `json:"abbreviation"`
				} `json:"team"`
			} `json:"competitors"`
		} `json:"competitions"`
	} `json:"events"`
}

type Team struct {
	Name     string
	HomeAway bool
	Score    float64
}

var VALID_SPORTS_LEAGUES = []string{"nfl", "nba", "mlb", "nhl"}
var SPORTS_LEAGUES_MAP = map[string]string{"nfl": "football", "nba": "basketball", "mlb": "baseball", "nhl": "hockey"}

func NewSportsNewsWidgetService(db *sql.DB) *SportsNewsWidgetService {
	return &SportsNewsWidgetService{
		db:     db,
		client: &http.Client{},
	}
}

// Create a new sports sports widget module
func (s *SportsNewsWidgetService) CreateSportsNewsWidget(sports_news_widget *widgets.NewsWidget) error {
	if sports_news_widget == nil {
		return errors.New("NewsWidget cannot be nil")
	}
	if sports_news_widget.COUNTRY == "" || sports_news_widget.LANGUAGE == "" || sports_news_widget.CATEGORY == "" {
		return errors.New("COUNTRY, LANGUAGE, and CATEGORY cannot be empty")
	}
	if slices.Contains(VALID_SPORTS_LEAGUES, sports_news_widget.COUNTRY) {
		return errors.New("COUNTRY must be one of the valid sports leagues: nfl, nba, mlb, nhl")
	}
	return widgets.CreateNews(s.db, sports_news_widget)
}

// Create multiple sports widgets in bulk
func (s *SportsNewsWidgetService) CreateSportsNewsWidgetsBulk(sports_news_widgets []widgets.NewsWidget) error {
	if len(sports_news_widgets) == 0 {
		return errors.New("no news widgets to create")
	}
	for _, sports_news_widget := range sports_news_widgets {
		if sports_news_widget.COUNTRY == "" || sports_news_widget.LANGUAGE == "" || sports_news_widget.CATEGORY == "" {
			return errors.New("COUNTRY, LANGUAGE, and CATEGORY cannot be empty")
		}
		if slices.Contains(VALID_SPORTS_LEAGUES, sports_news_widget.COUNTRY) {
			return errors.New("COUNTRY must be one of the valid sports leagues: nfl, nba, mlb, nhl")
		}
	}
	return widgets.CreateNewsBulk(s.db, sports_news_widgets)
}

// Retrieve a sports widget by its ID
func (s *SportsNewsWidgetService) GetSportsNewsWidgetByID(id int) (*widgets.NewsWidget, error) {
	if id <= 0 {
		return nil, errors.New("invalid NewsWidget ID")
	}
	sports_news_widget, err := widgets.GetNewsByID(s.db, id)
	if err != nil {
		return nil, fmt.Errorf("Failed to get NewsWidget by ID: %v", err)
	}
	if sports_news_widget == nil {
		return nil, fmt.Errorf("NewsWidget with ID %d not found", id)
	}
	return sports_news_widget, nil
}

// Retrieve all sports widgets for given carepartner and caregiver
func (s *SportsNewsWidgetService) GetSportsNewsWidgetsByCarepartnerAndCaregiver(carepartner_id int, caregiver_id int) ([]widgets.NewsWidget, error) {
	if carepartner_id <= 0 || caregiver_id <= 0 {
		return nil, errors.New("invalid Carepartner ID or Caregiver ID")
	}
	return widgets.GetNewsByCarepartnerAndCaregiver(s.db, carepartner_id, caregiver_id)
}

// Retrieve all sports widgets for a carepartner
func (s *SportsNewsWidgetService) GetSportsNewsWidgetsByCarepartnerIDAndDay(carepartner_id int, day, month, year int) ([]widgets.NewsWidget, error) {
	if carepartner_id <= 0 {
		return nil, errors.New("invalid Carepartner ID")
	}
	if day < 0 || month < 0 || year < 0 || day > 31 || month > 11 {
		return nil, errors.New("invalid date parameters")
	}
	return widgets.GetNewsByDayAndCarepartner(s.db, carepartner_id, day, month, year)
}

// Update an existing sports widget
func (s *SportsNewsWidgetService) UpdateSportsNewsWidget(sports_news_widget *widgets.NewsWidget) error {
	if sports_news_widget == nil {
		return errors.New("NewsWidget cannot be nil")
	}
	if sports_news_widget.ID <= 0 {
		return errors.New("invalid NewsWidget ID")
	}
	if sports_news_widget.COUNTRY == "" || sports_news_widget.LANGUAGE == "" || sports_news_widget.CATEGORY == "" {
		return errors.New("COUNTRY, LANGUAGE, and CATEGORY cannot be empty")
	}
	if slices.Contains(VALID_SPORTS_LEAGUES, sports_news_widget.COUNTRY) {
		return errors.New("COUNTRY must be one of the valid sports leagues: nfl, nba, mlb, nhl")
	}
	return widgets.UpdateNews(s.db, sports_news_widget)
}

// Delete a sports widget by its ID
func (s *SportsNewsWidgetService) DeleteSportsNewsWidget(id int) error {
	if id <= 0 {
		return errors.New("invalid NewsWidget ID")
	}
	return widgets.DeleteNews(s.db, id)
}

// Get total number of sports widgets
func (s *SportsNewsWidgetService) GetSportsNewsWidgetStats() (int, error) {
	return widgets.GetNewsStats(s.db)
}

func (s *SportsNewsWidgetService) GetTeamGame(league, team string) (string, error) {
	if league == "" || team == "" {
		return "", errors.New("league and team cannot be empty")
	}
	sport := SPORTS_LEAGUES_MAP[league]

	url := fmt.Sprintf(
		"http://site.api.espn.com/apis/site/v2/sports/%s/%s/teams/%s/schedule",
		sport, league, team,
	)
	resp, err := s.client.Get(url)
	if err != nil {
		return "", fmt.Errorf("Failed to fetch team schedule: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("ESPN APi returend status %d", resp.StatusCode)
	}
	var apiResponse SportAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		return "", fmt.Errorf("Failed to parse team schedule response: %v", err)
	}

	game_summary, err := CreateGameSummary(&apiResponse, team)
	if err != nil {
		return "", fmt.Errorf("Failed to create game summary: %v", err)
	}

	return game_summary, nil
}

func CreateGameSummary(apiResponse *SportAPIResponse, team string) (string, error) {
	if apiResponse == nil || len(apiResponse.Events) == 0 {
		return "No upcoming games found.", nil
	}

	game_index := 0
	now := time.Now().UTC()
	for i := len(apiResponse.Events) - 1; i >= 0; i-- {
		date, err := time.Parse(time.RFC3339, apiResponse.Events[i].Competitions[0].Date)
		if err != nil {
			return "", err
		}
		if date.Before(now) {
			game_index = i
			break
		}
	}

	var user_team, opponent_team Team
	competition := apiResponse.Events[game_index].Competitions[0]
	for _, competitor := range competition.Competitors {
		if competitor.Team.Abbreviation == team {
			user_team = Team{
				Name:     competitor.Team.DisplayName,
				HomeAway: competitor.HomeAway,
				Score:    competitor.Score.Value,
			}
		} else {
			opponent_team = Team{
				Name:     competitor.Team.DisplayName,
				HomeAway: competitor.HomeAway,
				Score:    competitor.Score.Value,
			}
		}
	}

	if user_team.Score > opponent_team.Score {
		if user_team.HomeAway {
			return fmt.Sprintf("In the last game, your team, the %s won against the %s at home, with a score of %d to %d.", user_team.Name, opponent_team.Name, int(user_team.Score), int(opponent_team.Score)), nil
		} else {
			return fmt.Sprintf("In the last game, your team, the %s won against the %s away, with a score of %d to %d.", user_team.Name, opponent_team.Name, int(user_team.Score), int(opponent_team.Score)), nil
		}
	} else {
		if user_team.HomeAway {
			return fmt.Sprintf("In the last game, your team, the %s lost against the %s at home, with a score of %d to %d.", user_team.Name, opponent_team.Name, int(user_team.Score), int(opponent_team.Score)), nil
		} else {
			return fmt.Sprintf("In the last game, your team, the %s lost against the %s away, with a score of %d to %d.", user_team.Name, opponent_team.Name, int(user_team.Score), int(opponent_team.Score)), nil
		}
	}
}
