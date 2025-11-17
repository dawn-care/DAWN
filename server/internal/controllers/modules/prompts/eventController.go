package prompts

import (
	"encoding/json"
	"net/http"
	"strconv"

	models "server/internal/models/modules/prompts"
)

type EventServiceInterface interface {
	CreateEventList(events []*models.EventPrompt) error
	GetEventListByListID(listID int) ([]models.EventPrompt, error)
	UpdateEvent(event *models.EventPrompt) error
	DeleteEvent(id int) error
	DeleteEventList(listID int) error
}

type EventController struct {
	service EventServiceInterface
}

func NewEventController(service EventServiceInterface) *EventController {
	return &EventController{service: service}
}

// Create a new list of events
func (c *EventController) CreateEventListHandler(w http.ResponseWriter, r *http.Request) {
	var events []*models.EventPrompt
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

// Get events by list ID
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

// Update an event
func (c *EventController) UpdateEventHandler(w http.ResponseWriter, r *http.Request) {
	var event models.EventPrompt
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

// Delete an event by ID
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

// Delete all events by list ID
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
