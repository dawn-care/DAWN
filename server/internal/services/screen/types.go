package screen

import (
	"encoding/json"

	"server/internal/models/scheduled_items"
	"server/internal/services/widgets/weather"
)

/*
Screen is everything the carepartner's tablet shows, as returned by GET /api/screen.
It is the contract between the backend and the tablet (and the caregiver's preview).

Times are in the carepartner's timezone. Regions are in the preset's order, and each
region's slots array always has exactly as many entries as the preset has slots there;
an empty slot is null. Widgets whose slot the current preset doesn't have are left out.
*/
type Screen struct {
	CarepartnerID int                          `json:"carepartner_id"`
	GeneratedAt   string                       `json:"generated_at"` // RFC 3339 with the carepartner's UTC offset
	Timezone      string                       `json:"timezone"`
	Today         string                       `json:"today"` // "YYYY-MM-DD" in the carepartner's timezone
	LayoutPreset  string                       `json:"layout_preset"`
	AwakeHours    AwakeHours                   `json:"awake_hours"`
	Regions       []Region                     `json:"regions"`
	Schedule      []scheduled_items.Occurrence `json:"schedule"` // today and tomorrow, so playback works across midnight
}

// AwakeHours are when the screen is on; outside them the tablet shows the screensaver
type AwakeHours struct {
	Wake       string `json:"wake"` // "HH:MM"
	Bed        string `json:"bed"`  // "HH:MM"
	IsAwakeNow bool   `json:"is_awake_now"`
}

// Region is one area of the layout, e.g. the three tiles
type Region struct {
	Name  string    `json:"name"`
	Slots []*Widget `json:"slots"` // null for an empty slot
}

/*
Widget is a widget placed in a slot, with its data filled in.
Config is the widget's settings as stored (see the screen widget endpoints).
Data depends on the type and is null for photo, message and reminder widgets,
whose content is all in Config and Photo:
  - photo: PhotoData (the caption to show)
  - weather: WeatherData
  - sports: SportsData
  - news: NewsData

If an outside service fails, Data is null and Error is set, so one bad lookup
never breaks the whole screen.
*/
type Widget struct {
	ID     int             `json:"id"`
	Type   string          `json:"type"`
	Config json.RawMessage `json:"config" swaggertype:"object"`
	Photo  *Photo          `json:"photo"`
	Data   any             `json:"data"`
	Error  *string         `json:"error"`
}

// Photo is a widget's photo, fetched from URL
type Photo struct {
	ID          int    `json:"id"`
	DisplayName string `json:"display_name"`
	URL         string `json:"url"`
}

// Widget error codes
const ErrorUpstreamUnavailable = "UPSTREAM_UNAVAILABLE"

// PhotoData is what to write under the photo frame: the caption, or the photo's name
type PhotoData struct {
	Caption string `json:"caption"`
}

// WeatherData is the current weather for the widget's zip code
type WeatherData = weather.WeatherOverviewResponse

// SportsData is the followed team's last game
type SportsData struct {
	League           string `json:"league"`
	TeamName         string `json:"team_name"`
	Result           string `json:"result"` // won, lost, tied or upcoming
	OpponentTeamName string `json:"opponent_team_name"`
	TeamScore        int    `json:"team_score"`
	OpponentScore    int    `json:"opponent_score"`
	Message          string `json:"message"`
}

// NewsData is a few good-news headlines
type NewsData struct {
	Articles []NewsArticle `json:"articles"`
}

type NewsArticle struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	ImageURL    string `json:"image_url"`
}

// MaxNewsArticles caps how many headlines a news widget gets
const MaxNewsArticles = 3
