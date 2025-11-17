package prompts

import (
	"database/sql"
	"errors"
	"fmt"

	models "server/internal/models/modules/prompts"
)

type EventService struct {
	DB *sql.DB
}

func NewEventService(db *sql.DB) *EventService {
	return &EventService{DB: db}
}

// Create a list of events
func (s *EventService) CreateEventList(events []*models.EventPrompt) error {
	if len(events) == 0 {
		return errors.New("events list cannot be empty")
	}
	for _, e := range events {
		if e.List_ID <= 0 || e.Text == "" {
			return errors.New("list_id and text are required for each event")
		}
	}
	return models.CreateEventList(s.DB, events)
}

// Get a list of events by list ID
func (s *EventService) GetEventListByListID(listID int) ([]models.EventPrompt, error) {
	if listID <= 0 {
		return nil, errors.New("invalid list ID")
	}
	events, err := models.GetEventListByListID(s.DB, &listID)
	if err != nil {
		return nil, fmt.Errorf("failed to get events: %w", err)
	}
	return events, nil
}

// Update a specific event
func (s *EventService) UpdateEvent(event *models.EventPrompt) error {
	if event == nil {
		return errors.New("event cannot be nil")
	}
	if event.ID <= 0 {
		return errors.New("invalid event ID")
	}
	return models.UpdateEvent(s.DB, event)
}

// Delete a specific event by ID
func (s *EventService) DeleteEvent(id int) error {
	if id <= 0 {
		return errors.New("invalid event ID")
	}
	return models.DeleteEvent(s.DB, &id)
}

// Delete all events by list ID
func (s *EventService) DeleteEventList(listID int) error {
	if listID <= 0 {
		return errors.New("invalid list ID")
	}
	return models.DeleteEventList(s.DB, &listID)
}
