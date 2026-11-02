package prompts

import (
	"database/sql"
	"testing"

	"server/internal/models/prompts"
)

func TestCreateCarepartnerLocation_Invalid_NoLocation(t *testing.T) {
	svc := NewCarepartnerLocationService(&sql.DB{})

	loc := &prompts.CarepartnerLocationPrompt{
		Location: "",
		Message:  "Missing location",
	}

	err := svc.CreateCarepartnerLocation(loc)
	if err == nil {
		t.Fatalf("expected error due to missing location, got nil")
	}
}

func TestGetCarepartnerLocationByID_InvalidID(t *testing.T) {
	svc := NewCarepartnerLocationService(&sql.DB{})

	_, err := svc.GetCarepartnerLocationByID(0)
	if err == nil {
		t.Fatalf("expected error for invalid ID, got nil")
	}
}

func TestUpdateCarepartnerLocation_InvalidID(t *testing.T) {
	svc := NewCarepartnerLocationService(&sql.DB{})

	loc := &prompts.CarepartnerLocationPrompt{
		ID:       0,
		Location: "Anywhere",
	}

	err := svc.UpdateCarepartnerLocation(loc)
	if err == nil {
		t.Fatalf("expected error for invalid ID, got nil")
	}
}

func TestDeleteCarepartnerLocation_InvalidID(t *testing.T) {
	svc := NewCarepartnerLocationService(&sql.DB{})

	err := svc.DeleteCarepartnerLocation(0)
	if err == nil {
		t.Fatalf("expected error for invalid ID, got nil")
	}
}
