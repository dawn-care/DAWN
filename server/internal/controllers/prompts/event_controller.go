package prompts

import (
	"encoding/json"
	"net/http"
	"strconv"

	"server/internal/models/prompts"
)

type EventServiceInterface interface {
	CreateEventList(events []*prompts.EventPrompt) error
	GetEventListByListID(listID int) ([]prompts.EventPrompt, error)
	UpdateEvent(event *prompts.EventPrompt) error
	DeleteEvent(id int) error
	DeleteEventList(listID int) error
}

type EventController struct {
	service EventServiceInterface
}

func NewEventController(service EventServiceInterface) *EventController {
	return &EventController{service: service}
}

// CreateEventListHandler godoc
// @Summary Create a new list of events
// @Description Creates a list of event prompts
// @Tags event
// @Accept json
// @Produce json
// @Param events body []prompts.EventPrompt true "Array of event prompts"
// @Success 200 {array} prompts.EventPrompt
// @Failure 400 {string} string "invalid request body"
// @Failure 500 {string} string "internal server error"
// @Router /events [post]
func (c *EventController) CreateEventListHandler(w http.ResponseWriter, r *http.Request) {
	var events []*prompts.EventPrompt
	if err := json.NewDecoder(r.Body).Decode(&events); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := c.service.CreateEventList(events); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(events)
}

// GetEventListByListIDHandler godoc
// @Summary Get events by list ID
// @Description Retrieves all event prompts for a given list ID
// @Tags event
// @Accept json
// @Produce json
// @Param list_id query int true "Event list ID"
// @Success 200 {array} prompts.EventPrompt
// @Failure 400 {string} string "invalid list ID"
// @Failure 500 {string} string "internal server error"
// @Router /events [get]
func (c *EventController) GetEventListByListIDHandler(w http.ResponseWriter, r *http.Request) {
	listIDStr := r.URL.Query().Get("list_id")
	listID, err := strconv.Atoi(listIDStr)
	if err != nil || listID <= 0 {
		http.Error(w, "invalid list ID", http.StatusBadRequest)
		return
	}

	events, err := c.service.GetEventListByListID(listID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(events)
}

// UpdateEventHandler godoc
// @Summary Update an event
// @Description Updates a single event prompt by ID
// @Tags event
// @Accept json
// @Produce json
// @Param event body prompts.EventPrompt true "Event prompt to update (must include ID)"
// @Success 200 {object} prompts.EventPrompt
// @Failure 400 {string} string "invalid request body or missing event ID"
// @Failure 500 {string} string "internal server error"
// @Router /events [put]
func (c *EventController) UpdateEventHandler(w http.ResponseWriter, r *http.Request) {
	var event prompts.EventPrompt
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if event.ID <= 0 {
		http.Error(w, "event ID is required", http.StatusBadRequest)
		return
	}

	if err := c.service.UpdateEvent(&event); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(event)
}

// DeleteEventHandler godoc
// @Summary Delete an event
// @Description Deletes a single event prompt by ID
// @Tags event
// @Accept json
// @Produce json
// @Param id query int true "Event ID"
// @Success 204 {string} string "No Content"
// @Failure 400 {string} string "invalid event ID"
// @Failure 500 {string} string "internal server error"
// @Router /events [delete]
func (c *EventController) DeleteEventHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil || id <= 0 {
		http.Error(w, "invalid event ID", http.StatusBadRequest)
		return
	}

	if err := c.service.DeleteEvent(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// DeleteEventListHandler godoc
// @Summary Delete all events in a list
// @Description Deletes all event prompts associated with a given list ID
// @Tags event
// @Accept json
// @Produce json
// @Param list_id query int true "Event list ID"
// @Success 204 {string} string "No Content"
// @Failure 400 {string} string "invalid list ID"
// @Failure 500 {string} string "internal server error"
// @Router /events/list [delete]
func (c *EventController) DeleteEventListHandler(w http.ResponseWriter, r *http.Request) {
	listID, err := strconv.Atoi(r.URL.Query().Get("list_id"))
	if err != nil || listID <= 0 {
		http.Error(w, "invalid list ID", http.StatusBadRequest)
		return
	}

	if err := c.service.DeleteEventList(listID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
