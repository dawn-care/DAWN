package widgets

import (
	"encoding/json"
	"net/http"
	models "server/internal/models/modules/widgets"
	service "server/internal/services/modules/widgets/news"
	"strconv"
)

// Service Interface
type NewsWidgetServiceInterface interface {
	CreateNewsWidget(weather_widget *models.NewsWidget) error
	CreateNewsWidgetsBulk(weather_widgets []models.NewsWidget) error
	GetNewsWidgetByID(id int) (*models.NewsWidget, error)
	GetNewsWidgetsByCarepartnerAndCaregiver(carepartner_id int, caregiver_id int) ([]models.NewsWidget, error)
	GetNewsWidgetsByCarepartnerIDAndDay(carepartner_id int, day, month, year int) ([]models.NewsWidget, error)
	UpdateNewsWidget(weather_widget *models.NewsWidget) error
	DeleteNewsWidget(id int) error
	GetNewsWidgetStats() (int, error)
	GetGoodNews() (*service.NewsAPIResponse, error)
	GetJoke() (string, error)
}

// Controller
type NewsWidgetController struct {
	service NewsWidgetServiceInterface
}

func NewNewsWidgetController(service NewsWidgetServiceInterface) *NewsWidgetController {
	return &NewsWidgetController{service: service}
}

// Handler for creating a new news widget
func (c *NewsWidgetController) CreateNewsWidgetHandler(w http.ResponseWriter, r *http.Request) {
	var weather_widget models.NewsWidget
	if err := json.NewDecoder(r.Body).Decode(&weather_widget); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := c.service.CreateNewsWidget(&weather_widget); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(weather_widget)
}

// Handler for getting a news widget by ID
func (c *NewsWidgetController) GetNewsWidgetByIDHandler(w http.ResponseWriter, r *http.Request) {
	idSTR := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idSTR)
	if err != nil || id <= 0 {
		http.Error(w, "Invalid NewsWidget ID", http.StatusBadRequest)
		return
	}

	weather_widget, err := c.service.GetNewsWidgetByID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(weather_widget)
}

// Handler for getting news widgets by carepartner and caregiver IDs
func (c *NewsWidgetController) GetNewsWidgetsByCarepartnerAndCaregiverHandler(w http.ResponseWriter, r *http.Request) {
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
	weather_widgets, err := c.service.GetNewsWidgetsByCarepartnerAndCaregiver(carepartnerID, caregiverID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(weather_widgets)
}

// Handler for getting news widgets by carepartner ID and date
func (c *NewsWidgetController) GetNewsByDayAndCarepartnerHandler(w http.ResponseWriter, r *http.Request) {
	carepartnerID, err := strconv.Atoi(r.URL.Query().Get("carepartner_id"))
	if err != nil || carepartnerID <= 0 {
		http.Error(w, "Invalid Carepartner ID", http.StatusBadRequest)
		return
	}
	day, err := strconv.Atoi(r.URL.Query().Get("day"))
	month, err := strconv.Atoi(r.URL.Query().Get("month"))
	year, err := strconv.Atoi(r.URL.Query().Get("year"))

	weather_widgets, err := c.service.GetNewsWidgetsByCarepartnerIDAndDay(carepartnerID, day, month, year)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(weather_widgets)
}

// Handler for updating an existing news widget
func (c *NewsWidgetController) UpdateNewsWidgetHandler(w http.ResponseWriter, r *http.Request) {
	var weather_widget models.NewsWidget
	if err := json.NewDecoder(r.Body).Decode(&weather_widget); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := c.service.UpdateNewsWidget(&weather_widget); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(weather_widget)
}

// Handler for deleting a news widget by ID
func (c *NewsWidgetController) DeleteNewsWidgetHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil || id <= 0 {
		http.Error(w, "Invalid NewsWidget ID", http.StatusBadRequest)
		return
	}

	if err := c.service.DeleteNewsWidget(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Handler for getting total number of news widgets
func (c *NewsWidgetController) GetNewsWidgetStatsHandler(w http.ResponseWriter, r *http.Request) {
	count, err := c.service.GetNewsWidgetStats()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(count)
}

// Handler for getting good news articles
func (c *NewsWidgetController) GetGoodNewsHandler(w http.ResponseWriter, r *http.Request) {
	newsResponse, err := c.service.GetGoodNews()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(newsResponse)
}

// Handler for getting a joke
func (c *NewsWidgetController) GetJokeHandler(w http.ResponseWriter, r *http.Request) {
	jokeResponse, err := c.service.GetJoke()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(jokeResponse)
}
