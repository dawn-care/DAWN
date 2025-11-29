package widgets

import (
	"encoding/json"
	"net/http"
	widgets "server/internal/models/widgets/news"
	"server/internal/services/widgets/news"
	"strconv"
)

// Service Interface
type NewsWidgetServiceInterface interface {
	CreateNewsWidget(weather_widget *widgets.NewsWidget) error
	CreateNewsWidgetsBulk(weather_widgets []widgets.NewsWidget) error
	GetNewsWidgetByID(id int) (*widgets.NewsWidget, error)
	GetNewsWidgetsByCarepartnerAndCaregiver(carepartner_id int, caregiver_id int) ([]widgets.NewsWidget, error)
	GetNewsWidgetsByCarepartnerIDAndDay(carepartner_id int, day, month, year int) ([]widgets.NewsWidget, error)
	UpdateNewsWidget(weather_widget *widgets.NewsWidget) error
	DeleteNewsWidget(id int) error
	GetNewsWidgetStats() (int, error)
	GetGoodNews() (*news.NewsAPIResponse, error)
	GetJoke() (string, error)
}

// Controller
type NewsWidgetController struct {
	service NewsWidgetServiceInterface
}

func NewNewsWidgetController(service NewsWidgetServiceInterface) *NewsWidgetController {
	return &NewsWidgetController{service: service}
}

// CreateNewsWidgetHandler godoc
// @Summary Create a news widget
// @Description Creates a new news widget
// @Tags news-widget
// @Accept json
// @Produce json
// @Param news_widget body widgets.NewsWidget true "News Widget to create"
// @Success 200 {object} widgets.NewsWidget
// @Failure 400 {object} map[string]string "invalid request body"
// @Failure 500 {object} map[string]string "internal server error"
// @Router /news-widgets [post]
func (c *NewsWidgetController) CreateNewsWidgetHandler(w http.ResponseWriter, r *http.Request) {
	var weather_widget widgets.NewsWidget
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

// GetNewsWidgetByIDHandler godoc
// @Summary Get a news widget by ID
// @Description Retrieves a news widget by its ID
// @Tags news-widget
// @Accept json
// @Produce json
// @Param id query int true "News Widget ID"
// @Success 200 {object} widgets.NewsWidget
// @Failure 400 {object} map[string]string "invalid ID"
// @Failure 500 {object} map[string]string "internal server error"
// @Router /news-widgets/{id} [get]
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

// GetNewsWidgetsByCarepartnerAndCaregiverHandler godoc
// @Summary Get news widgets by carepartner and caregiver
// @Description Retrieves all news widgets for a given carepartner and caregiver
// @Tags news-widget
// @Accept json
// @Produce json
// @Param carepartner_id query int true "Carepartner ID"
// @Param caregiver_id query int true "Caregiver ID"
// @Success 200 {array} widgets.NewsWidget
// @Failure 400 {object} map[string]string "invalid IDs"
// @Failure 500 {object} map[string]string "internal server error"
// @Router /news-widgets/by-carepartner-and-caregiver [get]
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

// GetNewsByDayAndCarepartnerHandler godoc
// @Summary Get news widgets by carepartner and date
// @Description Retrieves news widgets for a carepartner filtered by day, month, and year
// @Tags news-widget
// @Accept json
// @Produce json
// @Param carepartner_id query int true "Carepartner ID"
// @Param day query int true "Day"
// @Param month query int true "Month"
// @Param year query int true "Year"
// @Success 200 {array} widgets.NewsWidget
// @Failure 400 {object} map[string]string "invalid IDs or date"
// @Failure 500 {object} map[string]string "internal server error"
// @Router /news-widgets/by-day [get]
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

// UpdateNewsWidgetHandler godoc
// @Summary Update a news widget
// @Description Updates an existing news widget
// @Tags news-widget
// @Accept json
// @Produce json
// @Param news_widget body widgets.NewsWidget true "News Widget to update"
// @Success 200 {object} widgets.NewsWidget
// @Failure 400 {object} map[string]string "invalid request body"
// @Failure 500 {object} map[string]string "internal server error"
// @Router /news-widgets [put]
func (c *NewsWidgetController) UpdateNewsWidgetHandler(w http.ResponseWriter, r *http.Request) {
	var weather_widget widgets.NewsWidget
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

// DeleteNewsWidgetHandler godoc
// @Summary Delete a news widget
// @Description Deletes a news widget by ID
// @Tags news-widget
// @Accept json
// @Produce json
// @Param id query int true "News Widget ID"
// @Success 204 "No Content"
// @Failure 400 {object} map[string]string "invalid ID"
// @Failure 500 {object} map[string]string "internal server error"
// @Router /news-widgets/{id} [delete]
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

// GetNewsWidgetStatsHandler godoc
// @Summary Get total number of news widgets
// @Description Retrieves the total count of news widgets
// @Tags news-widget
// @Accept json
// @Produce json
// @Success 200 {integer} int "Total count"
// @Failure 500 {object} map[string]string "internal server error"
// @Router /news-widgets/stats [get]
func (c *NewsWidgetController) GetNewsWidgetStatsHandler(w http.ResponseWriter, r *http.Request) {
	count, err := c.service.GetNewsWidgetStats()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(count)
}

// GetGoodNewsHandler godoc
// @Summary Get good news articles
// @Description Retrieves curated good news articles
// @Tags news-widget
// @Accept json
// @Produce json
// @Success 200 {object} news.NewsAPIResponse
// @Failure 500 {object} map[string]string "internal server error"
// @Router /news-widgets/good-news [get]
func (c *NewsWidgetController) GetGoodNewsHandler(w http.ResponseWriter, r *http.Request) {
	newsResponse, err := c.service.GetGoodNews()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(newsResponse)
}

// GetJokeHandler godoc
// @Summary Get a joke
// @Description Retrieves a random joke
// @Tags news-widget
// @Accept json
// @Produce json
// @Success 200 {string} string "Joke text"
// @Failure 500 {object} map[string]string "internal server error"
// @Router /news-widgets/joke [get]
func (c *NewsWidgetController) GetJokeHandler(w http.ResponseWriter, r *http.Request) {
	jokeResponse, err := c.service.GetJoke()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(jokeResponse)
}
