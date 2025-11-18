package prompts

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	models "server/internal/models/prompts"
)

type mockEventService struct{}

func (m *mockEventService) CreateEventList(events []*models.EventPrompt) error {
	return nil
}
func (m *mockEventService) GetEventListByListID(listID int) ([]models.EventPrompt, error) {
	return []models.EventPrompt{{ID: 1, List_ID: listID, Text: "Test", Time: time.Now()}}, nil
}
func (m *mockEventService) UpdateEvent(event *models.EventPrompt) error { return nil }
func (m *mockEventService) DeleteEvent(id int) error                    { return nil }
func (m *mockEventService) DeleteEventList(listID int) error            { return nil }

// --- Tests ---
func TestCreateEventHandler_Success(t *testing.T) {
	controller := NewEventController(&mockEventService{})
	body := `[{"list_id":1,"text":"Test","time":"2025-01-01T12:00:00Z"}]`
	req := httptest.NewRequest(http.MethodPost, "/events", bytes.NewBufferString(body))
	w := httptest.NewRecorder()
	controller.CreateEventListHandler(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestGetEventListHandler_Success(t *testing.T) {
	controller := NewEventController(&mockEventService{})
	req := httptest.NewRequest(http.MethodGet, "/events?list_id=1", nil)
	w := httptest.NewRecorder()
	controller.GetEventListByListIDHandler(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestUpdateEventHandler_Success(t *testing.T) {
	controller := NewEventController(&mockEventService{})
	event := models.EventPrompt{ID: 1, List_ID: 1, Text: "Updated", Time: time.Now()}
	body, _ := json.Marshal(event)
	req := httptest.NewRequest(http.MethodPut, "/events", bytes.NewBuffer(body))
	w := httptest.NewRecorder()
	controller.UpdateEventHandler(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestDeleteEventHandler_Success(t *testing.T) {
	controller := NewEventController(&mockEventService{})
	req := httptest.NewRequest(http.MethodDelete, "/events?id=1", nil)
	w := httptest.NewRecorder()
	controller.DeleteEventHandler(w, req)
	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", w.Code)
	}
}
