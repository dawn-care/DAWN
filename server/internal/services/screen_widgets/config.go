package screen_widgets

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	sports "server/internal/services/widgets/news"
)

// Widget types, matching the CHECK constraint on "ScreenWidget".type
const (
	TypePhoto    = "photo"
	TypeMessage  = "message"
	TypeReminder = "reminder"
	TypeWeather  = "weather"
	TypeSports   = "sports"
	TypeNews     = "news"
)

var WidgetTypes = []string{TypePhoto, TypeMessage, TypeReminder, TypeWeather, TypeSports, TypeNews}

// DefaultWeatherZip is St. Ann's, used when a weather widget has no zip
const DefaultWeatherZip = "14617"

// ReminderIcons are the icons a reminder item can show; "other" is used when none is given
var ReminderIcons = []string{"meal", "medication", "walk", "visit", "appointment", "call", "rest", "other"}

// MaxReminderItems caps how many reminders one widget holds
const MaxReminderItems = 10

// Settings for each widget type. Unknown fields are rejected so typos surface.

type PhotoConfig struct {
	Caption string `json:"caption,omitempty"` // e.g. the daughter's name under the frame
}

type MessageConfig struct {
	Text     string `json:"text"`
	Location string `json:"location,omitempty"`
	About    string `json:"about,omitempty"` // "caregiver" or "carepartner" (location reassurance)
	Label    string `json:"label,omitempty"` // a role under the name, e.g. "Nurse"
}

type ReminderItem struct {
	Time string `json:"time"` // "HH:MM"; the tablet highlights a reminder that's coming up
	Text string `json:"text"`
	Icon string `json:"icon"`
}

type ReminderConfig struct {
	Items []ReminderItem `json:"items"`
}

type WeatherConfig struct {
	Zip string `json:"zip"`
}

type SportsConfig struct {
	League string `json:"league"` // nfl, nba, mlb or nhl
	Team   string `json:"team"`   // abbreviation from GET /api/widgets/sports/teams
}

type NewsConfig struct{}

var zipPattern = regexp.MustCompile(`^\d{5}$`)

func validType(t string) bool {
	return contains(WidgetTypes, t)
}

func contains(list []string, v string) bool {
	for _, x := range list {
		if x == v {
			return true
		}
	}
	return false
}

func configErr(format string, args ...any) error {
	return fmt.Errorf("%w: %s", ErrInvalidConfig, fmt.Sprintf(format, args...))
}

// decodeStrict reads a config object, treating a missing config as {}
func decodeStrict(raw json.RawMessage, v any) error {
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 || bytes.Equal(trimmed, []byte("null")) {
		trimmed = []byte("{}")
	}
	if trimmed[0] != '{' {
		return configErr("config must be a JSON object")
	}
	dec := json.NewDecoder(bytes.NewReader(trimmed))
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		return configErr("%v", err)
	}
	return nil
}

// cleanText trims s and checks it is at most max characters (and present if required)
func cleanText(field, s string, max int, required bool) (string, error) {
	s = strings.TrimSpace(s)
	if required && s == "" {
		return "", configErr("%s is required", field)
	}
	if utf8.RuneCountInString(s) > max {
		return "", configErr("%s must be at most %d characters", field, max)
	}
	return s, nil
}

/*
NormalizeConfig validates a widget's settings for its type and returns them in
canonical form: trimmed text, zero-padded times, sorted reminders, an uppercase team,
and the default zip for weather.
*/
func NormalizeConfig(widgetType string, raw json.RawMessage) (json.RawMessage, error) {
	var out any
	var err error

	switch widgetType {
	case TypePhoto:
		var c PhotoConfig
		if err = decodeStrict(raw, &c); err != nil {
			return nil, err
		}
		if c.Caption, err = cleanText("caption", c.Caption, 60, false); err != nil {
			return nil, err
		}
		out = c

	case TypeMessage:
		var c MessageConfig
		if err = decodeStrict(raw, &c); err != nil {
			return nil, err
		}
		if c.Text, err = cleanText("text", c.Text, 200, true); err != nil {
			return nil, err
		}
		if c.Location, err = cleanText("location", c.Location, 100, false); err != nil {
			return nil, err
		}
		if c.Label, err = cleanText("label", c.Label, 40, false); err != nil {
			return nil, err
		}
		if c.About != "" && c.About != "caregiver" && c.About != "carepartner" {
			return nil, configErr("about must be caregiver or carepartner")
		}
		out = c

	case TypeReminder:
		var c ReminderConfig
		if err = decodeStrict(raw, &c); err != nil {
			return nil, err
		}
		if len(c.Items) == 0 || len(c.Items) > MaxReminderItems {
			return nil, configErr("items must have 1 to %d reminders", MaxReminderItems)
		}
		for i := range c.Items {
			item := &c.Items[i]
			t, perr := time.Parse("15:04", strings.TrimSpace(item.Time))
			if perr != nil {
				return nil, configErr("items[%d].time must be HH:MM", i)
			}
			item.Time = t.Format("15:04")
			if item.Text, err = cleanText(fmt.Sprintf("items[%d].text", i), item.Text, 60, true); err != nil {
				return nil, err
			}
			if item.Icon == "" {
				item.Icon = "other"
			}
			if !contains(ReminderIcons, item.Icon) {
				return nil, configErr("items[%d].icon must be one of %v", i, ReminderIcons)
			}
		}
		sort.SliceStable(c.Items, func(a, b int) bool { return c.Items[a].Time < c.Items[b].Time })
		out = c

	case TypeWeather:
		var c WeatherConfig
		if err = decodeStrict(raw, &c); err != nil {
			return nil, err
		}
		c.Zip = strings.TrimSpace(c.Zip)
		if c.Zip == "" {
			c.Zip = DefaultWeatherZip
		}
		if !zipPattern.MatchString(c.Zip) {
			return nil, configErr("zip must be a 5 digit US zip code")
		}
		out = c

	case TypeSports:
		var c SportsConfig
		if err = decodeStrict(raw, &c); err != nil {
			return nil, err
		}
		c.League = strings.ToLower(strings.TrimSpace(c.League))
		c.Team = strings.ToUpper(strings.TrimSpace(c.Team))
		if c.League == "" || c.Team == "" {
			return nil, configErr("league and team are required")
		}
		if _, ok := sports.FindTeam(c.League, c.Team); !ok {
			return nil, configErr("unknown team %q in league %q; see GET /api/widgets/sports/teams", c.Team, c.League)
		}
		out = c

	case TypeNews:
		var c NewsConfig
		if err = decodeStrict(raw, &c); err != nil {
			return nil, err
		}
		out = c

	default:
		return nil, ErrInvalidType
	}

	return json.Marshal(out)
}
