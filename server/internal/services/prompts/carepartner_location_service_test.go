package prompts

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	models "server/internal/models/prompts"
)

// --- Mock service implementing CarepartnerLocationServiceInterface ---

type mockCarepartnerLocationService struct {
	createErr error
	getErr    error
	updateErr error
	deleteErr error
	locResult *models.CarepartnerLocationPrompt
}

func (m *mockCarepartnerLocationService) CreateCarepartnerLocation(loc *models.CarepartnerLocationPrompt) error {
	if loc == nil || loc.Location == "" || loc.Message == "" {
		return errors.New("invalid input")
	}
	loc.ID = 10
	return m.createErr
}

func (m *mockCarepartnerLocationService) GetCarepartnerLocationByID(id int) (*models.CarepartnerLocationPrompt, error) {
	if id <= 0 {
		return nil, errors.New("invalid id")
	}
	return m.locResult, m.getErr
}

func (m *mockCarepartnerLocationService) UpdateCarepartnerLocation(loc *models.CarepartnerLocationPrompt) error {
	if loc == nil || loc.ID <= 0 {
		return errors.New("invalid input")
	}
	return m.updateErr
}

func (m *mockCarepartnerLocationService) DeleteCarepartnerLocation(id int) error {
	if id <= 0 {
		return errors.New("invalid id")
	}
	return m.deleteErr
}

// --- Tests ---

func TestCreateCarepartnerLocationHandler_Success(t *testing.T) {
	mockSvc := &mockCarepartnerLocationService{}
	controller := NewCarepartnerLocationController(mockSvc)

	reqBody := `{"location":"Home","message":"OK"}`
	req := httptest.NewRequest(http.MethodPost, "/carepartner-location", strings.NewReader(reqBody))
	w := httptest.NewRecorder()

	controller.CreateCarepartnerLocationHandler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestGetCarepartnerLocationByIDHandler_Success(t *testing.T) {
	mockSvc := &mockCarepartnerLocationService{
		locResult: &models.CarepartnerLocationPrompt{
			ID:       1,
			Location: "Home",
			Message:  "OK",
		},
	}
	controller := NewCarepartnerLocationController(mockSvc)

	req := httptest.NewRequest(http.MethodGet, "/carepartner-location?id=1", nil)
	w := httptest.NewRecorder()

	controller.GetCarepartnerLocationByIDHandler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestGetCarepartnerLocationByIDHandler_InvalidID(t *testing.T) {
	controller := NewCarepartnerLocationController(&mockCarepartnerLocationService{})

	req := httptest.NewRequest(http.MethodGet, "/carepartner-location?id=abc", nil)
	w := httptest.NewRecorder()

	controller.GetCarepartnerLocationByIDHandler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestUpdateCarepartnerLocationHandler_Success(t *testing.T) {
	mockSvc := &mockCarepartnerLocationService{}
	controller := NewCarepartnerLocationController(mockSvc)

	reqBody := `{"id":5,"location":"Work","message":"Updated"}`
	req := httptest.NewRequest(http.MethodPut, "/carepartner-location", strings.NewReader(reqBody))
	w := httptest.NewRecorder()

	controller.UpdateCarepartnerLocationHandler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestDeleteCarepartnerLocationHandler_Success(t *testing.T) {
	mockSvc := &mockCarepartnerLocationService{}
	controller := NewCarepartnerLocationController(mockSvc)

	req := httptest.NewRequest(http.MethodDelete, "/carepartner-location?id=10", nil)
	w := httptest.NewRecorder()

	controller.DeleteCarepartnerLocationHandler(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", w.Code)
	}
}

func TestDeleteCarepartnerLocationHandler_InvalidID(t *testing.T) {
	controller := NewCarepartnerLocationController(&mockCarepartnerLocationService{})

	req := httptest.NewRequest(http.MethodDelete, "/carepartner-location?id=0", nil)
	w := httptest.NewRecorder()

	controller.DeleteCarepartnerLocationHandler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}
