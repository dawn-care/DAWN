package prompts_test

import (
	"database/sql"
	"testing"
	"time"

	models "server/internal/models/modules/prompts"
	service "server/internal/services/modules/prompts"
)

type mockEventServiceDB struct {
	createErr error
	getResult []models.EventPrompt
	getErr    error
	updateErr error
	deleteErr error
}

func (m *mockEventServiceDB) CreateEventList(db *sql.DB, events []*models.EventPrompt) error {
	return m.createErr
}
func (m *mockEventServiceDB) GetEventListByListID(db *sql.DB, listID *int) ([]models.EventPrompt, error) {
	return m.getResult, m.getErr
}
func (m *mockEventServiceDB) UpdateEvent(db *sql.DB, event *models.EventPrompt) error {
	return m.updateErr
}
func (m *mockEventServiceDB) DeleteEvent(db *sql.DB, id *int) error {
	return m.deleteErr
}
func (m *mockEventServiceDB) DeleteEventList(db *sql.DB, listID *int) error {
	return m.deleteErr
}

// ---- Tests ----
func TestCreateEventList_Success(t *testing.T) {
	svc := service.NewEventService(&sql.DB{})
	event := &models.EventPrompt{List_ID: 1, Text: "Test Event", Time: time.Now()}
	err := svc.CreateEventList([]*models.EventPrompt{event})
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
}

func TestGetEventListByListID_Success(t *testing.T) {
	svc := service.NewEventService(&sql.DB{})
	_, err := svc.GetEventListByListID(1)
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
}

func TestUpdateEvent_Success(t *testing.T) {
	svc := service.NewEventService(&sql.DB{})
	err := svc.UpdateEvent(&models.EventPrompt{ID: 1})
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
}

func TestDeleteEvent_Success(t *testing.T) {
	svc := service.NewEventService(&sql.DB{})
	err := svc.DeleteEvent(1)
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
}

func TestDeleteEventList_Success(t *testing.T) {
	svc := service.NewEventService(&sql.DB{})
	err := svc.DeleteEventList(1)
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
}
