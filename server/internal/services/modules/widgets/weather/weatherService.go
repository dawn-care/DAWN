package weather

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"

	"server/internal/models/modules/widgets"
)

type WeatherService struct {
	db     *sql.DB
	apiKey string
	client *http.Client
}

type CoordinateResponse struct {
	Lat  float64 `json:"lat"`
	Lon  float64 `json:"lon"`
	Name string  `json:"name"`
	Zip  string  `json:"zip"`
}

type OneCallOverviewResponse struct {
	Lat             float64 `json:"lat"`
	Lon             float64 `json:"lon"`
	TZ              string  `json:"tz"`
	Date            string  `json:"date"`
	Units           string  `json:"units"`
	WeatherOverview string  `json:"weather_overview"`
}

func NewWeatherService(db *sql.DB, apiKey string) *WeatherService {
	return &WeatherService{
		db:     db,
		apiKey: apiKey,
		client: &http.Client{},
	}
}

// Create a new weather widget module
func (s *WeatherService) CreateWeatherWidget(weather_widget *widgets.WeatherWidget) error {
	if weather_widget == nil {
		return errors.New("WeatherWidget cannot be nil")
	}
	regex := `^\d{5}(?:[-\s]\d{4})?$`
	match, _ := regexp.MatchString(regex, weather_widget.ZIPCODE)
	if !match {
		return errors.New("ZIPCODE must be in valid zip code format")
	}
	return widgets.CreateWeather(s.db, weather_widget)
}

// Create multiple weather widgets in bulk
func (s *WeatherService) CreateWeatherWidgetsBulk(weather_widgets []widgets.WeatherWidget) error {
	if len(weather_widgets) == 0 {
		return errors.New("no weather widgets to create")
	}
	for _, weather_widget := range weather_widgets {
		regex := `^\d{5}(?:[-\s]\d{4})?$`
		match, _ := regexp.MatchString(regex, weather_widget.ZIPCODE)
		if !match {
			return errors.New("ZIPCODE must be in valid zip code format")
		}
	}
	return widgets.CreateWeathersBulk(s.db, weather_widgets)
}

// Retrieve a weather widget by its ID
func (s *WeatherService) GetWeatherWidgetByID(id int) (*widgets.WeatherWidget, error) {
	if id <= 0 {
		return nil, errors.New("invalid WeatherWidget ID")
	}
	weather_widget, err := widgets.GetWeatherByID(s.db, id)
	if err != nil {
		return nil, fmt.Errorf("Failed to get WeatherWidget by ID: %v", err)
	}
	if weather_widget == nil {
		return nil, fmt.Errorf("WeatherWidget with ID %d not found", id)
	}
	return weather_widget, nil
}

// Retrieve all weather widgets for given carepartner and caregiver
func (s *WeatherService) GetWeatherWidgetsByCarepartnerAndCaregiver(carepartner_id int, caregiver_id int) ([]widgets.WeatherWidget, error) {
	if carepartner_id <= 0 || caregiver_id <= 0 {
		return nil, errors.New("invalid Carepartner ID or Caregiver ID")
	}
	return widgets.GetWeatherByCarepartnerAndCaregiver(s.db, carepartner_id, caregiver_id)
}

// Retrieve all weather widgets for a carepartner
func (s *WeatherService) GetWeatherWidgetsByCarepartnerIDAndDay(carepartner_id int, day, month, year int) ([]widgets.WeatherWidget, error) {
	if carepartner_id <= 0 {
		return nil, errors.New("invalid Carepartner ID")
	}
	if day < 0 || month < 0 || year < 0 || day > 31 || month > 11 {
		return nil, errors.New("invalid date parameters")
	}
	return widgets.GetWeatherByDayAndCarepartner(s.db, carepartner_id, day, month, year)
}

// Update an existing weather widget
func (s *WeatherService) UpdateWeatherWidget(weather_widget *widgets.WeatherWidget) error {
	if weather_widget == nil {
		return errors.New("WeatherWidget cannot be nil")
	}
	if weather_widget.ID <= 0 {
		return errors.New("invalid WeatherWidget ID")
	}
	return widgets.UpdateWeather(s.db, weather_widget)
}

// Delete a weather widget by its ID
func (s *WeatherService) DeleteWeatherWidget(id int) error {
	if id <= 0 {
		return errors.New("invalid WeatherWidget ID")
	}
	return widgets.DeleteWeather(s.db, id)
}

func (s *WeatherService) GetWeatherOverview(weather_widget *widgets.WeatherWidget) (string, error) {
	url := fmt.Sprintf(
		"https://api.openweathermap.org/geo/1.0/zip?zip=%s,US&appid=%s",
		weather_widget.ZIPCODE,
		s.apiKey,
	)

	resp, err := s.client.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed geocode request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("geocode API returned status %d", resp.StatusCode)
	}

	var coordinate_response CoordinateResponse
	if err := json.NewDecoder(resp.Body).Decode(&coordinate_response); err != nil {
		return "", fmt.Errorf("failed parsing geocode response: %v", err)
	}

	unitsParam := "imperial"

	url = fmt.Sprintf(
		"https://api.openweathermap.org/data/3.0/onecall/overview?lon=%f&lat=%f&units=%s&appid=%s",
		coordinate_response.Lon, coordinate_response.Lat, unitsParam, s.apiKey,
	)

	resp, err = s.client.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed overview request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("overview API returned status %d", resp.StatusCode)
	}

	var oneCallResponse OneCallOverviewResponse
	if err := json.NewDecoder(resp.Body).Decode(&oneCallResponse); err != nil {
		return "", fmt.Errorf("failed parsing overview response: %v", err)
	}

	if oneCallResponse.WeatherOverview == "" {
		return "", fmt.Errorf("overview response missing weather_overview")
	}

	return oneCallResponse.WeatherOverview, nil
}

// Get total number of weather widgets
func (s *WeatherService) GetWeatherWidgetStats() (int, error) {
	return widgets.GetWeatherStats(s.db)
}
