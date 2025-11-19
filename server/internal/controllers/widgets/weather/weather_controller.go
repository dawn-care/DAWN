package widgets

import (
	"encoding/json"
	"net/http"
	models "server/internal/models/widgets/weather"
	"strconv"
)

// Service Interface
type WeatherWidgetServiceInterface interface {
	CreateWeatherWidget(weather_widget *models.WeatherWidget) error
	CreateWeatherWidgetsBulk(weather_widgets []models.WeatherWidget) error
	GetWeatherWidgetByID(id int) (*models.WeatherWidget, error)
	GetWeatherWidgetsByCarepartnerAndCaregiver(carepartner_id int, caregiver_id int) ([]models.WeatherWidget, error)
	GetWeatherWidgetsByCarepartnerIDAndDay(carepartner_id int, day, month, year int) ([]models.WeatherWidget, error)
	UpdateWeatherWidget(weather_widget *models.WeatherWidget) error
	DeleteWeatherWidget(id int) error
	GetWeatherOverview(weather_widget *models.WeatherWidget) (string, error)
	GetWeatherWidgetStats() (int, error)
}

// Controller
type WeatherWidgetController struct {
	service WeatherWidgetServiceInterface
}

func NewWeatherWidgetController(service WeatherWidgetServiceInterface) *WeatherWidgetController {
	return &WeatherWidgetController{service: service}
}

// Handler for creating a new weather widget
func (c *WeatherWidgetController) CreateWeatherWidgetHandler(w http.ResponseWriter, r *http.Request) {
	var weather_widget models.WeatherWidget
	if err := json.NewDecoder(r.Body).Decode(&weather_widget); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := c.service.CreateWeatherWidget(&weather_widget); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(weather_widget)
}

// Handler for getting a weather widget by ID
func (c *WeatherWidgetController) GetWeatherWidgetByIDHandler(w http.ResponseWriter, r *http.Request) {
	idSTR := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idSTR)
	if err != nil || id <= 0 {
		http.Error(w, "Invalid WeatherWidget ID", http.StatusBadRequest)
		return
	}

	weather_widget, err := c.service.GetWeatherWidgetByID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(weather_widget)
}

// Handler for getting weather widgets by carepartner and caregiver IDs
func (c *WeatherWidgetController) GetWeatherWidgetsByCarepartnerAndCaregiverHandler(w http.ResponseWriter, r *http.Request) {
	carepartnerID, err := strconv.Atoi(r.URL.Query().Get("carepartner_id"))
	if err != nil || carepartnerID <= 0 {
		http.Error(w, "Invalid Carepartner ID", http.StatusBadRequest)
		return
	}
	caregiverID, err := strconv.Atoi(r.URL.Query().Get("caregiver_id"))
	if err != nil || caregiverID <= 0 {
		http.Error(w, "Invalid Caregiver ID", http.StatusBadRequest)
		return
	}
	weather_widgets, err := c.service.GetWeatherWidgetsByCarepartnerAndCaregiver(carepartnerID, caregiverID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(weather_widgets)
}

// Handler for getting weather widgets by carepartner ID and date
func (c *WeatherWidgetController) GetWeatherByDayAndCarepartnerHandler(w http.ResponseWriter, r *http.Request) {
	carepartnerID, err := strconv.Atoi(r.URL.Query().Get("carepartner_id"))
	if err != nil || carepartnerID <= 0 {
		http.Error(w, "Invalid Carepartner ID", http.StatusBadRequest)
		return
	}
	day, _ := strconv.Atoi(r.URL.Query().Get("day"))
	month, _ := strconv.Atoi(r.URL.Query().Get("month"))
	year, _ := strconv.Atoi(r.URL.Query().Get("year"))

	weather_widgets, err := c.service.GetWeatherWidgetsByCarepartnerIDAndDay(carepartnerID, day, month, year)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(weather_widgets)
}

// Handler for updating an existing weather widget
func (c *WeatherWidgetController) UpdateWeatherWidgetHandler(w http.ResponseWriter, r *http.Request) {
	var weather_widget models.WeatherWidget
	if err := json.NewDecoder(r.Body).Decode(&weather_widget); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := c.service.UpdateWeatherWidget(&weather_widget); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(weather_widget)
}

// Handler for deleting a weather widget by ID
func (c *WeatherWidgetController) DeleteWeatherWidgetHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil || id <= 0 {
		http.Error(w, "Invalid WeatherWidget ID", http.StatusBadRequest)
		return
	}

	if err := c.service.DeleteWeatherWidget(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Handler for getting daily weather overview
func (c *WeatherWidgetController) GetWeatherOverviewHandler(w http.ResponseWriter, r *http.Request) {
	var weather_widget models.WeatherWidget
	if err := json.NewDecoder(r.Body).Decode(&weather_widget); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	overview, err := c.service.GetWeatherOverview(&weather_widget)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(overview)
}

// Handler for getting total number of weather widgets
func (c *WeatherWidgetController) GetWeatherWidgetStatsHandler(w http.ResponseWriter, r *http.Request) {
	count, err := c.service.GetWeatherWidgetStats()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(count)
}
