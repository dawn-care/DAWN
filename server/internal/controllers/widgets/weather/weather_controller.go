package widgets

import (
	"encoding/json"
	"net/http"
	widgets "server/internal/models/widgets/weather"
	"strconv"

	"github.com/go-chi/chi/v5"
)

// Service Interface
type WeatherWidgetServiceInterface interface {
	CreateWeatherWidget(weather_widget *widgets.WeatherWidget) error
	CreateWeatherWidgetsBulk(weather_widgets []widgets.WeatherWidget) error
	GetWeatherWidgetByID(id int) (*widgets.WeatherWidget, error)
	GetWeatherWidgetsByCarepartnerAndCaregiver(carepartner_id int, caregiver_id int) ([]widgets.WeatherWidget, error)
	GetWeatherWidgetsByCarepartnerIDAndDay(carepartner_id int, day, month, year int) ([]widgets.WeatherWidget, error)
	UpdateWeatherWidget(weather_widget *widgets.WeatherWidget) error
	DeleteWeatherWidget(id int) error
	GetWeatherOverview(weather_widget *widgets.WeatherWidget) (string, error)
	GetWeatherWidgetStats() (int, error)
}

// Controller
type WeatherWidgetController struct {
	service WeatherWidgetServiceInterface
}

func NewWeatherWidgetController(service WeatherWidgetServiceInterface) *WeatherWidgetController {
	return &WeatherWidgetController{service: service}
}

// CreateWeatherWidgetHandler godoc
// @Summary Create a weather widget
// @Description Creates a new weather widget
// @Tags weather-widget
// @Accept json
// @Produce json
// @Param weather_widget body widgets.WeatherWidget true "Weather Widget to create"
// @Success 200 {object} widgets.WeatherWidget
// @Failure 400 {object} map[string]string "invalid request body"
// @Failure 500 {object} map[string]string "internal server error"
// @Router /weather-widgets [post]
func (c *WeatherWidgetController) CreateWeatherWidgetHandler(w http.ResponseWriter, r *http.Request) {
	var weather_widget widgets.WeatherWidget
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

// GetWeatherWidgetByIDHandler godoc
// @Summary Get a weather widget by ID
// @Description Retrieves a weather widget by its ID
// @Tags weather-widget
// @Accept json
// @Produce json
// @Param id query int true "Weather Widget ID"
// @Success 200 {object} widgets.WeatherWidget
// @Failure 400 {object} map[string]string "invalid WeatherWidget ID"
// @Failure 500 {object} map[string]string "internal server error"
// @Router /weather-widgets/{id} [get]
func (c *WeatherWidgetController) GetWeatherWidgetByIDHandler(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if id <= 0 {
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

// GetWeatherWidgetsByCarepartnerAndCaregiverHandler godoc
// @Summary Get weather widgets by carepartner and caregiver
// @Description Retrieves all weather widgets for a given carepartner and caregiver
// @Tags weather-widget
// @Accept json
// @Produce json
// @Param carepartner_id query int true "Carepartner ID"
// @Param caregiver_id query int true "Caregiver ID"
// @Success 200 {array} widgets.WeatherWidget
// @Failure 400 {object} map[string]string "invalid Carepartner or Caregiver ID"
// @Failure 500 {object} map[string]string "internal server error"
// @Router /weather-widgets/by-carepartner-and-caregiver [get]
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

// GetWeatherByDayAndCarepartnerHandler godoc
// @Summary Get weather widgets by carepartner and date
// @Description Retrieves weather widgets for a carepartner filtered by day, month, and year
// @Tags weather-widget
// @Accept json
// @Produce json
// @Param carepartner_id query int true "Carepartner ID"
// @Param day query int true "Day"
// @Param month query int true "Month"
// @Param year query int true "Year"
// @Success 200 {array} widgets.WeatherWidget
// @Failure 400 {object} map[string]string "invalid Carepartner ID or date"
// @Failure 500 {object} map[string]string "internal server error"
// @Router /weather-widgets/by-day [get]
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

// UpdateWeatherWidgetHandler godoc
// @Summary Update a weather widget
// @Description Updates an existing weather widget
// @Tags weather-widget
// @Accept json
// @Produce json
// @Param weather_widget body widgets.WeatherWidget true "Weather Widget to update"
// @Success 200 {object} widgets.WeatherWidget
// @Failure 400 {object} map[string]string "invalid request body"
// @Failure 500 {object} map[string]string "internal server error"
// @Router /weather-widgets [put]
func (c *WeatherWidgetController) UpdateWeatherWidgetHandler(w http.ResponseWriter, r *http.Request) {
	var weather_widget widgets.WeatherWidget
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

// DeleteWeatherWidgetHandler godoc
// @Summary Delete a weather widget
// @Description Deletes a weather widget by ID
// @Tags weather-widget
// @Accept json
// @Produce json
// @Param id query int true "Weather Widget ID"
// @Success 204 "No Content"
// @Failure 400 {object} map[string]string "invalid WeatherWidget ID"
// @Failure 500 {object} map[string]string "internal server error"
// @Router /weather-widgets/{id} [delete]
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

// GetWeatherOverviewHandler godoc
// @Summary Get daily weather overview
// @Description Generates a daily weather overview for a given weather widget
// @Tags weather-widget
// @Accept json
// @Produce json
// @Param weather_widget body widgets.WeatherWidget true "Weather Widget"
// @Success 200 {string} string "Weather overview"
// @Failure 400 {object} map[string]string "invalid request body"
// @Failure 500 {object} map[string]string "internal server error"
// @Router /weather-widgets/overview [post]
func (c *WeatherWidgetController) GetWeatherOverviewHandler(w http.ResponseWriter, r *http.Request) {
	var weather_widget widgets.WeatherWidget
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

// GetWeatherWidgetStatsHandler godoc
// @Summary Get total number of weather widgets
// @Description Retrieves the total count of weather widgets
// @Tags weather-widget
// @Accept json
// @Produce json
// @Success 200 {integer} int "Total count"
// @Failure 500 {object} map[string]string "internal server error"
// @Router /weather-widgets/stats [get]
func (c *WeatherWidgetController) GetWeatherWidgetStatsHandler(w http.ResponseWriter, r *http.Request) {
	count, err := c.service.GetWeatherWidgetStats()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(count)
}
