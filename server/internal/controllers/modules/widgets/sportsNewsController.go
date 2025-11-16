package widgets

import (
	"encoding/json"
	"net/http"
	models "server/internal/models/modules/widgets"
	"strconv"
)

// Service Interface
type SportsNewsWidgetServiceInterface interface {
	CreateSportsNewsWidget(weather_widget *models.NewsWidget) error
	CreateSportsNewsWidgetsBulk(weather_widgets []models.NewsWidget) error
	GetSportsNewsWidgetByID(id int) (*models.NewsWidget, error)
	GetSportsNewsWidgetsByCarepartnerAndCaregiver(carepartner_id int, caregiver_id int) ([]models.NewsWidget, error)
	GetSportsNewsWidgetsByCarepartnerIDAndDay(carepartner_id int, day, month, year int) ([]models.NewsWidget, error)
	UpdateSportsNewsWidget(weather_widget *models.NewsWidget) error
	DeleteSportsNewsWidget(id int) error
	GetSportsNewsWidgetStats() (int, error)
	GetTeamGame(league, team string) (string, error)
}

// Controller
type SportsNewsWidgetController struct {
	service SportsNewsWidgetServiceInterface
}

func NewSportsNewsWidgetController(service SportsNewsWidgetServiceInterface) *SportsNewsWidgetController {
	return &SportsNewsWidgetController{service: service}
}

// Handler for creating a new sports news widget
func (c *SportsNewsWidgetController) CreateSportsNewsWidgetHandler(w http.ResponseWriter, r *http.Request) {
	var weather_widget models.NewsWidget
	if err := json.NewDecoder(r.Body).Decode(&weather_widget); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := c.service.CreateSportsNewsWidget(&weather_widget); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(weather_widget)
}

// Handler for getting a sports news widget by ID
func (c *SportsNewsWidgetController) GetSportsNewsWidgetByIDHandler(w http.ResponseWriter, r *http.Request) {
	idSTR := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idSTR)
	if err != nil || id <= 0 {
		http.Error(w, "Invalid SportsNewsWidget ID", http.StatusBadRequest)
		return
	}

	weather_widget, err := c.service.GetSportsNewsWidgetByID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(weather_widget)
}

// Handler for getting a sports news widget by carepartner and caregiver IDs
func (c *SportsNewsWidgetController) GetSportsNewsWidgetsByCarepartnerAndCaregiverHandler(w http.ResponseWriter, r *http.Request) {
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
	weather_widgets, err := c.service.GetSportsNewsWidgetsByCarepartnerAndCaregiver(carepartnerID, caregiverID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(weather_widgets)
}

// Handler for getting a sports news widget by carepartner ID and date
func (c *SportsNewsWidgetController) GetSportsNewsByDayAndCarepartnerHandler(w http.ResponseWriter, r *http.Request) {
	carepartnerID, err := strconv.Atoi(r.URL.Query().Get("carepartner_id"))
	if err != nil || carepartnerID <= 0 {
		http.Error(w, "Invalid Carepartner ID", http.StatusBadRequest)
		return
	}
	day, err := strconv.Atoi(r.URL.Query().Get("day"))
	month, err := strconv.Atoi(r.URL.Query().Get("month"))
	year, err := strconv.Atoi(r.URL.Query().Get("year"))

	weather_widgets, err := c.service.GetSportsNewsWidgetsByCarepartnerIDAndDay(carepartnerID, day, month, year)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(weather_widgets)
}

// Handler for updating a sports news widget
func (c *SportsNewsWidgetController) UpdateSportsNewsWidgetHandler(w http.ResponseWriter, r *http.Request) {
	var weather_widget models.NewsWidget
	if err := json.NewDecoder(r.Body).Decode(&weather_widget); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := c.service.UpdateSportsNewsWidget(&weather_widget); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(weather_widget)
}

// Handler for deleting a sports news widget by ID
func (c *SportsNewsWidgetController) DeleteSportsNewsWidgetHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil || id <= 0 {
		http.Error(w, "Invalid SportsNewsWidget ID", http.StatusBadRequest)
		return
	}

	if err := c.service.DeleteSportsNewsWidget(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Handler for getting sports news widget statistics
func (c *SportsNewsWidgetController) GetSportsNewsWidgetStatsHandler(w http.ResponseWriter, r *http.Request) {
	count, err := c.service.GetSportsNewsWidgetStats()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(count)
}

// Handler for getting a team's game summary
func (c *SportsNewsWidgetController) GetTeamGameHandler(w http.ResponseWriter, r *http.Request) {
	league := r.URL.Query().Get("league")
	team := r.URL.Query().Get("team")
	if league == "" || team == "" {
		http.Error(w, "league and team cannot be empty", http.StatusBadRequest)
		return
	}

	game_summary, err := c.service.GetTeamGame(league, team)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(game_summary)
}
