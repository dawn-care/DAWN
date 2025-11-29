package prompts

import (
	"database/sql"
	"testing"

	"server/internal/models/prompts"
)

func TestCreateCaregiverLocation_Success(t *testing.T) {
	svc := NewCaregiverLocationService(&sql.DB{})

	loc := &prompts.CaregiverLocationPrompt{
		Location: "Home",
		Message:  "Arrived safely",
		URI:      "image.jpg",
	}

	err := svc.CreateCaregiverLocation(loc)
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
}

func TestCreateCaregiverLocation_Invalid_NoLocation(t *testing.T) {
	svc := NewCaregiverLocationService(&sql.DB{})

	loc := &prompts.CaregiverLocationPrompt{
		Location: "",
		Message:  "Missing location field",
	}

	err := svc.CreateCaregiverLocation(loc)
	if err == nil {
		t.Fatalf("expected error for missing location, got nil")
	}
}

func TestGetCaregiverLocationByID_Success(t *testing.T) {
	svc := NewCaregiverLocationService(&sql.DB{})

	_, err := svc.GetCaregiverLocationByID(1)
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
}

func TestGetCaregiverLocationByID_InvalidID(t *testing.T) {
	svc := NewCaregiverLocationService(&sql.DB{})

	_, err := svc.GetCaregiverLocationByID(0)
	if err == nil {
		t.Fatalf("expected error for invalid ID, got nil")
	}
}

func TestUpdateCaregiverLocation_Success(t *testing.T) {
	svc := NewCaregiverLocationService(&sql.DB{})

	loc := &prompts.CaregiverLocationPrompt{
		ID:       1,
		Location: "Clinic",
		Message:  "On the way",
		URI:      "",
	}

	err := svc.UpdateCaregiverLocation(loc)
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
}

func TestUpdateCaregiverLocation_Invalid_NoID(t *testing.T) {
	svc := NewCaregiverLocationService(&sql.DB{})

	loc := &prompts.CaregiverLocationPrompt{
		ID:       0,
		Location: "Home",
	}

	err := svc.UpdateCaregiverLocation(loc)
	if err == nil {
		t.Fatalf("expected error for missing ID, got nil")
	}
}

func TestDeleteCaregiverLocation_Success(t *testing.T) {
	svc := NewCaregiverLocationService(&sql.DB{})

	err := svc.DeleteCaregiverLocation(1)
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
}

func TestDeleteCaregiverLocation_InvalidID(t *testing.T) {
	svc := NewCaregiverLocationService(&sql.DB{})

	err := svc.DeleteCaregiverLocation(0)
	if err == nil {
		t.Fatalf("expected error for invalid ID, got nil")
	}
}
