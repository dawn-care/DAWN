package prompts

import (
	"database/sql"
	"testing"
	"time"

	"server/internal/models/prompts"
)

func TestCreateEventList_Success(t *testing.T) {
	svc := NewEventService(&sql.DB{})
	event := &prompts.EventPrompt{List_ID: 1, Text: "Test Event", Time: time.Now()}
	err := svc.CreateEventList([]*prompts.EventPrompt{event})
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
}

func TestGetEventListByListID_Success(t *testing.T) {
	svc := NewEventService(&sql.DB{})
	_, err := svc.GetEventListByListID(1)
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
}

func TestDeleteEvent_Success(t *testing.T) {
	svc := NewEventService(&sql.DB{})
	err := svc.DeleteEvent(1)
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
}

func TestDeleteEventList_Success(t *testing.T) {
	svc := NewEventService(&sql.DB{})
	err := svc.DeleteEventList(1)
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
}
