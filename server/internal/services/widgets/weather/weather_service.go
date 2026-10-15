package weather

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"net/http"
	neturl "net/url"
	"regexp"
	"time"

	widgets "server/internal/models/widgets/weather"
)

// OpenWeatherBaseURL is where the OpenWeather API lives; tests point a service elsewhere
const OpenWeatherBaseURL = "https://api.openweathermap.org"

// Service struct
type WeatherService struct {
	db      *sql.DB
	apiKey  string
	client  *http.Client
	baseURL string
}

// Struct to represent the geocoding response from OpenWeather API
type CoordinateResponse struct {
	Lat  float64 `json:"lat"`
	Lon  float64 `json:"lon"`
	Name string  `json:"name"`
	Zip  string  `json:"zip"`
}

// Struct to represent the overview response from OpenWeather API
type OneCallOverviewResponse struct {
	Lat             float64 `json:"lat"`
	Lon             float64 `json:"lon"`
	TZ              string  `json:"tz"`
	Date            string  `json:"date"`
	Units           string  `json:"units"`
	WeatherOverview string  `json:"weather_overview"`
}

// Struct to represent the current weather response from OpenWeather API
type CurrentWeatherResponse struct {
	Main struct {
		Temp      float64 `json:"temp"`
		FeelsLike float64 `json:"feels_like"`
		Humidity  int     `json:"humidity"`
		TempMin   float64 `json:"temp_min"`
		TempMax   float64 `json:"temp_max"`
	} `json:"main"`
	Weather []struct {
		Main        string `json:"main"`
		Description string `json:"description"`
	} `json:"weather"`
	Wind struct {
		Speed float64 `json:"speed"`
	} `json:"wind"`
	Name string `json:"name"`
}

// Struct to represent the weather overview information to be sent to the frontend for display in the weather widget
type WeatherOverviewResponse struct {
	Location    string `json:"location"`
	Temp        int    `json:"temp"`
	FeelsLike   int    `json:"feels_like"`
	TempMin     int    `json:"temp_min"`
	TempMax     int    `json:"temp_max"`
	Humidity    int    `json:"humidity"`
	WindSpeed   int    `json:"wind_speed"`
	Condition   string `json:"condition"`
	Description string `json:"description"`
	Message     string `json:"message"`
}

func NewWeatherService(db *sql.DB, apiKey string) *WeatherService {
	return &WeatherService{
		db:      db,
		apiKey:  apiKey,
		client:  &http.Client{Timeout: 10 * time.Second},
		baseURL: OpenWeatherBaseURL,
	}
}

// CREATE OPERATIONS

/*
Create a new weather widget module
weather_widget - information inputted formatted as a WeatherWidget struct
*/
func (s *WeatherService) CreateWeatherWidget(weather_widget *widgets.WeatherWidget) error {
	if weather_widget == nil {
		return errors.New("WeatherWidget cannot be nil")
	}
	regex := `^\d{5}(?:[-\s]\d{4})?$`
	match, _ := regexp.MatchString(regex, weather_widget.ZIPCODE)
	if !match {
		return errors.New("ZIPCODE must be in valid zip code format")
	}
	_, err := widgets.CreateWeather(s.db, weather_widget)
	if err != nil {
		return fmt.Errorf("failed to create weather widget: %v", err)
	}
	return nil
}

/*
Create multiple weather widgets in bulk
weather_widget - information inputted formatted as a WeatherWidget struct
*/
func (s *WeatherService) CreateWeatherWidgetsBulk(weather_widgets []widgets.WeatherWidget) error {
	if len(weather_widgets) == 0 {
		return errors.New("no weather widgets to create")
	}

	zipRegex := regexp.MustCompile(`^\d{5}(?:[-\s]\d{4})?$`)

	for _, weather_widget := range weather_widgets {
		if !zipRegex.MatchString(weather_widget.ZIPCODE) {
			return errors.New("ZIPCODE must be in valid zip code format")
		}
	}

	return widgets.CreateWeathersBulk(s.db, weather_widgets)
}

// RETRIEVE OPERATIONS

/*
Retrieve a weather widget by its ID
id - The ID of the weather widget to retrieve
*/
func (s *WeatherService) GetWeatherWidgetByID(id int) (*widgets.WeatherWidget, error) {
	if id <= 0 {
		return nil, errors.New("invalid WeatherWidget ID")
	}
	weather_widget, err := widgets.GetWeatherByID(s.db, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get weatherwidget by id: %v", err)
	}
	if weather_widget == nil {
		return nil, fmt.Errorf("WeatherWidget with ID %d not found", id)
	}
	return weather_widget, nil
}

/*
Retrieve all weather widgets for given carepartner and caregiver
carepartner_id - The carepartner ID associated with the weather widgets to retrieve
caregiver_id - The caregiver ID associated with the weather widgets to retrieve
*/
func (s *WeatherService) GetWeatherWidgetsByCarepartnerAndCaregiver(carepartner_id int, caregiver_id int) ([]widgets.WeatherWidget, error) {
	if carepartner_id <= 0 || caregiver_id <= 0 {
		return nil, errors.New("invalid Carepartner ID or Caregiver ID")
	}
	return widgets.GetWeatherByCarepartnerAndCaregiver(s.db, carepartner_id, caregiver_id)
}

/*
Retrieve all weather widgets for a carepartner on a given day
carepartner_id - The carepartner ID associated with the weather widgets to retrieve
day, month, year - The date to retrieve weather widgets for
*/
func (s *WeatherService) GetWeatherWidgetsByCarepartnerIDAndDay(carepartner_id int, day, month, year int) ([]widgets.WeatherWidget, error) {
	if carepartner_id <= 0 {
		return nil, errors.New("invalid Carepartner ID")
	}
	if day < 0 || month < 0 || year < 0 || day > 31 || month > 11 {
		return nil, errors.New("invalid date parameters")
	}
	return widgets.GetWeatherByDayAndCarepartner(s.db, carepartner_id, day, month, year)
}

// UPDATE OPERATION

/*
Update an existing weather widget
weather_widget - information inputted formatted as a WeatherWidget struct
*/
func (s *WeatherService) UpdateWeatherWidget(weather_widget *widgets.WeatherWidget) error {
	if weather_widget == nil {
		return errors.New("WeatherWidget cannot be nil")
	}
	if weather_widget.ID <= 0 {
		return errors.New("invalid WeatherWidget ID")
	}
	return widgets.UpdateWeather(s.db, weather_widget)
}

// DELETE OPERATION

/*
Delete a weather widget by its ID
id - The ID of the weather widget to delete
*/
func (s *WeatherService) DeleteWeatherWidget(id int) error {
	if id <= 0 {
		return errors.New("invalid WeatherWidget ID")
	}
	return widgets.DeleteWeather(s.db, id)
}

// ANALYTICS & STATS

/*
Get total number of weather widgets in the database for analytics purposes
*/
func (s *WeatherService) GetWeatherWidgetStats() (int, error) {
	return widgets.GetWeatherStats(s.db)
}

// EXTERNAL API CALLS

/*
Gets the weather overview for an old WeatherWidget row (used by /api/weather/overview)
weather_widget - weather widget containing the zipcode information
*/
func (s *WeatherService) GetWeatherOverview(weather_widget *widgets.WeatherWidget) (*WeatherOverviewResponse, error) {
	return s.GetWeatherForZip(weather_widget.ZIPCODE)
}

/*
Uses OpenWeather API to get coordinates from a US zip code, then fetches current weather
and the overview for the day. Used with a weather screen widget's config.zip.
The day's overview text is optional: if that request fails (it needs a One Call 3.0
subscription), the current weather is still returned with an empty Message.
zip - 5 digit US zip code
*/
func (s *WeatherService) GetWeatherForZip(zip string) (*WeatherOverviewResponse, error) {
	if zip == "" {
		return nil, errors.New("zip cannot be empty")
	}

	// 1. Geocode zipcode to coordinates
	url := fmt.Sprintf(
		"%s/geo/1.0/zip?zip=%s,US&appid=%s",
		s.baseURL,
		neturl.QueryEscape(zip),
		s.apiKey,
	)

	resp, err := s.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed geocode request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("geocode API returned status %d", resp.StatusCode)
	}

	var coordinate_response CoordinateResponse
	if err := json.NewDecoder(resp.Body).Decode(&coordinate_response); err != nil {
		return nil, fmt.Errorf("failed parsing geocode response: %v", err)
	}

	unitsParam := "imperial"

	// 2. Fetch current weather (temp, conditions, etc.)
	currentURL := fmt.Sprintf(
		"%s/data/2.5/weather?lat=%f&lon=%f&units=%s&appid=%s",
		s.baseURL, coordinate_response.Lat, coordinate_response.Lon, unitsParam, s.apiKey,
	)

	currentResp, err := s.client.Get(currentURL)
	if err != nil {
		return nil, fmt.Errorf("failed current weather request: %v", err)
	}
	defer currentResp.Body.Close()

	if currentResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("current weather API returned status %d", currentResp.StatusCode)
	}

	var currentWeather CurrentWeatherResponse
	if err := json.NewDecoder(currentResp.Body).Decode(&currentWeather); err != nil {
		return nil, fmt.Errorf("failed parsing current weather response: %v", err)
	}

	// 3. Fetch overview text (optional; the widget still works without it)
	overviewURL := fmt.Sprintf(
		"%s/data/3.0/onecall/overview?lon=%f&lat=%f&units=%s&appid=%s",
		s.baseURL, coordinate_response.Lon, coordinate_response.Lat, unitsParam, s.apiKey,
	)
	oneCallResponse, err := s.getOverview(overviewURL)
	if err != nil {
		log.Printf("weather: no daily overview for %s: %v", zip, err)
		oneCallResponse = &OneCallOverviewResponse{}
	}

	// 4. Build structured response
	condition := ""
	description := ""
	if len(currentWeather.Weather) > 0 {
		condition = currentWeather.Weather[0].Main
		description = currentWeather.Weather[0].Description
	}

	location := currentWeather.Name
	if location == "" {
		location = coordinate_response.Name
	}

	return &WeatherOverviewResponse{
		Location:    location,
		Temp:        int(math.Round(currentWeather.Main.Temp)),
		FeelsLike:   int(math.Round(currentWeather.Main.FeelsLike)),
		TempMin:     int(math.Round(currentWeather.Main.TempMin)),
		TempMax:     int(math.Round(currentWeather.Main.TempMax)),
		Humidity:    currentWeather.Main.Humidity,
		WindSpeed:   int(math.Round(currentWeather.Wind.Speed)),
		Condition:   condition,
		Description: description,
		Message:     oneCallResponse.WeatherOverview,
	}, nil
}

// getOverview fetches the day's weather overview text
func (s *WeatherService) getOverview(url string) (*OneCallOverviewResponse, error) {
	resp, err := s.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed overview request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("overview API returned status %d", resp.StatusCode)
	}

	var oneCallResponse OneCallOverviewResponse
	if err := json.NewDecoder(resp.Body).Decode(&oneCallResponse); err != nil {
		return nil, fmt.Errorf("failed parsing overview response: %v", err)
	}
	return &oneCallResponse, nil
}
