package prompts_test

import (
	"database/sql"
	"testing"

	db "server/internal/models/modules/prompts"
	models "server/internal/models/modules/prompts"
	service "server/internal/services/modules/prompts"
)

// ---- Mock DB function overrides ----

func TestCreateCaregiverLocation_Success(t *testing.T) {
	original := db.CreateCaregiverLocation
	db.CreateCaregiverLocation = func(_ *sql.DB, loc *models.CaregiverLocationPrompt) error {
		loc.ID = 10
		return nil
	}
	defer func() { db.CreateCaregiverLocation = original }()

	svc := service.NewCaregiverLocationService(&sql.DB{})

	loc := &models.CaregiverLocationPrompt{
		Location: "Home",
		Message:  "OK",
	}

	err := svc.CreateCaregiverLocation(loc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if loc.ID != 10 {
		t.Fatalf("expected ID 10, got %d", loc.ID)
	}
}

func TestCreateCaregiverLocation_Invalid(t *testing.T) {
	svc := service.NewCaregiverLocationService(&sql.DB{})

	err := svc.CreateCaregiverLocation(nil)
	if err == nil {
		t.Fatalf("expected error for nil input")
	}

	err = svc.CreateCaregiverLocation(&models.CaregiverLocationPrompt{})
	if err == nil {
		t.Fatalf("expected error for missing fields")
	}
}

func TestGetCaregiverLocationByID_Success(t *testing.T) {
	original := db.GetCaregiverLocationByID
	db.GetCaregiverLocationByID = func(_ *sql.DB, id int) (*models.CaregiverLocationPrompt, error) {
		return &models.CaregiverLocationPrompt{
			ID:       1,
			Location: "Home",
			Message:  "OK",
		}, nil
	}
	defer func() { db.GetCaregiverLocationByID = original }()

	svc := service.NewCaregiverLocationService(&sql.DB{})

	loc, err := svc.GetCaregiverLocationByID(1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if loc.ID != 1 {
		t.Fatalf("expected ID 1, got %d", loc.ID)
	}
}

func TestGetCaregiverLocationByID_NotFound(t *testing.T) {
	original := db.GetCaregiverLocationByID
	db.GetCaregiverLocationByID = func(_ *sql.DB, id int) (*models.CaregiverLocationPrompt, error) {
		return nil, nil
	}
	defer func() { db.GetCaregiverLocationByID = original }()

	svc := service.NewCaregiverLocationService(&sql.DB{})

	_, err := svc.GetCaregiverLocationByID(99)
	if err == nil {
		t.Fatalf("expected not found error")
	}
}

func TestUpdateCaregiverLocation_Success(t *testing.T) {
	original := db.UpdateCaregiverLocation
	db.UpdateCaregiverLocation = func(_ *sql.DB, _ *models.CaregiverLocationPrompt) error {
		return nil
	}
	defer func() { db.UpdateCaregiverLocation = original }()

	svc := service.NewCaregiverLocationService(&sql.DB{})

	err := svc.UpdateCaregiverLocation(&models.CaregiverLocationPrompt{ID: 5})
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
}

func TestUpdateCaregiverLocation_Invalid(t *testing.T) {
	svc := service.NewCaregiverLocationService(&sql.DB{})

	err := svc.UpdateCaregiverLocation(nil)
	if err == nil {
		t.Fatalf("expected error for nil input")
	}

	err = svc.UpdateCaregiverLocation(&models.CaregiverLocationPrompt{ID: 0})
	if err == nil {
		t.Fatalf("expected error for invalid ID")
	}
}

func TestDeleteCaregiverLocation_Success(t *testing.T) {
	original := db.DeleteCaregiverLocation
	db.DeleteCaregiverLocation = func(_ *sql.DB, _ int) error {
		return nil
	}
	defer func() { db.DeleteCaregiverLocation = original }()

	svc := service.NewCaregiverLocationService(&sql.DB{})

	err := svc.DeleteCaregiverLocation(10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteCaregiverLocation_Invalid(t *testing.T) {
	svc := service.NewCaregiverLocationService(&sql.DB{})

	err := svc.DeleteCaregiverLocation(0)
	if err == nil {
		t.Fatalf("expected error for invalid ID")
	}
}
