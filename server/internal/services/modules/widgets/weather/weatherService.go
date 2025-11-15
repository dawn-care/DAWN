package weather

import (
	"database/sql"
	"errors"
	"fmt"
	"regexp"

	"server/internal/models/modules/widgets"
)

type WeatherService struct {
	db *sql.DB
}

func NewWeatherService(db *sql.DB) *WeatherService {
	return &WeatherService{db: db}
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

// Get total number of weather widgets
func (s *WeatherService) GetWeatherWidgetStats() (int, error) {
	return widgets.GetWeatherStats(s.db)
}
