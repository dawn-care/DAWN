package prompts

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	models "server/internal/models/prompts"
)

// --- Mock service implementing CaregiverLocationServiceInterface ---

type mockCaregiverLocationService struct {
	createErr error
	getErr    error
	updateErr error
	deleteErr error
	locResult *models.CaregiverLocationPrompt
}

func (m *mockCaregiverLocationService) CreateCaregiverLocation(loc *models.CaregiverLocationPrompt) error {
	if loc == nil || loc.Location == "" || loc.Message == "" {
		return errors.New("invalid input")
	}
	loc.ID = 10
	return m.createErr
}

func (m *mockCaregiverLocationService) GetCaregiverLocationByID(id int) (*models.CaregiverLocationPrompt, error) {
	if id <= 0 {
		return nil, errors.New("invalid id")
	}
	return m.locResult, m.getErr
}

func (m *mockCaregiverLocationService) UpdateCaregiverLocation(loc *models.CaregiverLocationPrompt) error {
	if loc == nil || loc.ID <= 0 {
		return errors.New("invalid input")
	}
	return m.updateErr
}

func (m *mockCaregiverLocationService) DeleteCaregiverLocation(id int) error {
	if id <= 0 {
		return errors.New("invalid id")
	}
	return m.deleteErr
}

// --- Tests ---

func TestCreateCaregiverLocationHandler_Success(t *testing.T) {
	mockSvc := &mockCaregiverLocationService{}
	controller := NewCaregiverLocationController(mockSvc)

	reqBody := `{"location":"Home","message":"OK"}`
	req := httptest.NewRequest(http.MethodPost, "/caregiver-location", strings.NewReader(reqBody))
	w := httptest.NewRecorder()

	controller.CreateCaregiverLocationHandler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestGetCaregiverLocationByIDHandler_Success(t *testing.T) {
	mockSvc := &mockCaregiverLocationService{
		locResult: &models.CaregiverLocationPrompt{
			ID:       1,
			Location: "Home",
			Message:  "OK",
		},
	}
	controller := NewCaregiverLocationController(mockSvc)

	req := httptest.NewRequest(http.MethodGet, "/caregiver-location?id=1", nil)
	w := httptest.NewRecorder()

	controller.GetCaregiverLocationByIDHandler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestGetCaregiverLocationByIDHandler_InvalidID(t *testing.T) {
	controller := NewCaregiverLocationController(&mockCaregiverLocationService{})

	req := httptest.NewRequest(http.MethodGet, "/caregiver-location?id=abc", nil)
	w := httptest.NewRecorder()

	controller.GetCaregiverLocationByIDHandler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestUpdateCaregiverLocationHandler_Success(t *testing.T) {
	mockSvc := &mockCaregiverLocationService{}
	controller := NewCaregiverLocationController(mockSvc)

	reqBody := `{"id":5,"location":"Work","message":"Updated"}`
	req := httptest.NewRequest(http.MethodPut, "/caregiver-location", strings.NewReader(reqBody))
	w := httptest.NewRecorder()

	controller.UpdateCaregiverLocationHandler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestDeleteCaregiverLocationHandler_Success(t *testing.T) {
	mockSvc := &mockCaregiverLocationService{}
	controller := NewCaregiverLocationController(mockSvc)

	req := httptest.NewRequest(http.MethodDelete, "/caregiver-location?id=10", nil)
	w := httptest.NewRecorder()

	controller.DeleteCaregiverLocationHandler(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", w.Code)
	}
}

func TestDeleteCaregiverLocationHandler_InvalidID(t *testing.T) {
	controller := NewCaregiverLocationController(&mockCaregiverLocationService{})

	req := httptest.NewRequest(http.MethodDelete, "/caregiver-location?id=0", nil)
	w := httptest.NewRecorder()

	controller.DeleteCaregiverLocationHandler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}
