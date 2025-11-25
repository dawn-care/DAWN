package widgets

import (
	"encoding/json"
	"net/http"
	widgets "server/internal/models/widgets/news"
	"strconv"
)

// Service Interface
type SportsNewsWidgetServiceInterface interface {
	CreateSportsNewsWidget(weather_widget *widgets.NewsWidget) error
	CreateSportsNewsWidgetsBulk(weather_widgets []widgets.NewsWidget) error
	GetSportsNewsWidgetByID(id int) (*widgets.NewsWidget, error)
	GetSportsNewsWidgetsByCarepartnerAndCaregiver(carepartner_id int, caregiver_id int) ([]widgets.NewsWidget, error)
	GetSportsNewsWidgetsByCarepartnerIDAndDay(carepartner_id int, day, month, year int) ([]widgets.NewsWidget, error)
	UpdateSportsNewsWidget(weather_widget *widgets.NewsWidget) error
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

// CreateSportsNewsWidgetHandler godoc
// @Summary Create a sports news widget
// @Description Creates a new sports news widget
// @Tags sports-news-widget
// @Accept json
// @Produce json
// @Param sports_news_widget body widgets.NewsWidget true "Sports News Widget to create"
// @Success 200 {object} widgets.NewsWidget
// @Failure 400 {object} map[string]string "invalid request body"
// @Failure 500 {object} map[string]string "internal server error"
// @Router /sports-news-widgets [post]
func (c *SportsNewsWidgetController) CreateSportsNewsWidgetHandler(w http.ResponseWriter, r *http.Request) {
	var weather_widget widgets.NewsWidget
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

// GetSportsNewsWidgetByIDHandler godoc
// @Summary Get a sports news widget by ID
// @Description Retrieves a sports news widget by its ID
// @Tags sports-news-widget
// @Accept json
// @Produce json
// @Param id query int true "Sports News Widget ID"
// @Success 200 {object} widgets.NewsWidget
// @Failure 400 {object} map[string]string "invalid ID"
// @Failure 500 {object} map[string]string "internal server error"
// @Router /sports-news-widgets/{id} [get]
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

// GetSportsNewsWidgetsByCarepartnerAndCaregiverHandler godoc
// @Summary Get sports news widgets by carepartner and caregiver
// @Description Retrieves all sports news widgets for a given carepartner and caregiver
// @Tags sports-news-widget
// @Accept json
// @Produce json
// @Param carepartner_id query int true "Carepartner ID"
// @Param caregiver_id query int true "Caregiver ID"
// @Success 200 {array} widgets.NewsWidget
// @Failure 400 {object} map[string]string "invalid IDs"
// @Failure 500 {object} map[string]string "internal server error"
// @Router /sports-news-widgets/by-carepartner-and-caregiver [get]
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

// GetSportsNewsByDayAndCarepartnerHandler godoc
// @Summary Get sports news widgets by carepartner and date
// @Description Retrieves sports news widgets for a carepartner filtered by day, month, and year
// @Tags sports-news-widget
// @Accept json
// @Produce json
// @Param carepartner_id query int true "Carepartner ID"
// @Param day query int true "Day"
// @Param month query int true "Month"
// @Param year query int true "Year"
// @Success 200 {array} widgets.NewsWidget
// @Failure 400 {object} map[string]string "invalid IDs or date"
// @Failure 500 {object} map[string]string "internal server error"
// @Router /sports-news-widgets/by-day [get]
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

// UpdateSportsNewsWidgetHandler godoc
// @Summary Update a sports news widget
// @Description Updates an existing sports news widget
// @Tags sports-news-widget
// @Accept json
// @Produce json
// @Param sports_news_widget body widgets.NewsWidget true "Sports News Widget to update"
// @Success 200 {object} widgets.NewsWidget
// @Failure 400 {object} map[string]string "invalid request body"
// @Failure 500 {object} map[string]string "internal server error"
// @Router /sports-news-widgets [put]
func (c *SportsNewsWidgetController) UpdateSportsNewsWidgetHandler(w http.ResponseWriter, r *http.Request) {
	var weather_widget widgets.NewsWidget
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

// DeleteSportsNewsWidgetHandler godoc
// @Summary Delete a sports news widget
// @Description Deletes a sports news widget by ID
// @Tags sports-news-widget
// @Accept json
// @Produce json
// @Param id query int true "Sports News Widget ID"
// @Success 204 "No Content"
// @Failure 400 {object} map[string]string "invalid ID"
// @Failure 500 {object} map[string]string "internal server error"
// @Router /sports-news-widgets/{id} [delete]
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

// GetSportsNewsWidgetStatsHandler godoc
// @Summary Get total number of sports news widgets
// @Description Retrieves the total count of sports news widgets
// @Tags sports-news-widget
// @Accept json
// @Produce json
// @Success 200 {integer} int "Total count"
// @Failure 500 {object} map[string]string "internal server error"
// @Router /sports-news-widgets/stats [get]
func (c *SportsNewsWidgetController) GetSportsNewsWidgetStatsHandler(w http.ResponseWriter, r *http.Request) {
	count, err := c.service.GetSportsNewsWidgetStats()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(count)
}

// GetTeamGameHandler godoc
// @Summary Get a team's game summary
// @Description Retrieves the latest game summary for a specific team in a league
// @Tags sports-news-widget
// @Accept json
// @Produce json
// @Param league query string true "League"
// @Param team query string true "Team"
// @Success 200 {string} string "Game summary"
// @Failure 400 {object} map[string]string "league and team cannot be empty"
// @Failure 500 {object} map[string]string "internal server error"
// @Router /sports-news-widgets/team-game [get]
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
