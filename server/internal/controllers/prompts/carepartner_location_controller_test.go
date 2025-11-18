package prompts

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	models "server/internal/models/prompts"
)

// Mock service implementing CarepartnerLocationServiceInterface
type mockCarepartnerLocationService struct {
	createErr error
	getErr    error
	updateErr error
	deleteErr error

	locationResult *models.CarepartnerLocationPrompt
}

func (m *mockCarepartnerLocationService) CreateCarepartnerLocation(loc *models.CarepartnerLocationPrompt) error {
	return m.createErr
}

func (m *mockCarepartnerLocationService) GetCarepartnerLocationByID(id int) (*models.CarepartnerLocationPrompt, error) {
	return m.locationResult, m.getErr
}

func (m *mockCarepartnerLocationService) UpdateCarepartnerLocation(loc *models.CarepartnerLocationPrompt) error {
	return m.updateErr
}

func (m *mockCarepartnerLocationService) DeleteCarepartnerLocation(id int) error {
	return m.deleteErr
}

// --- Tests ---

func TestCreateCarepartnerLocationHandler_Success(t *testing.T) {
	mockSvc := &mockCarepartnerLocationService{}
	controller := NewCarepartnerLocationController(mockSvc)

	body := `{"location": "Home", "message": "Arrived"}`
	req := httptest.NewRequest(http.MethodPost, "/location", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	controller.CreateCarepartnerLocationHandler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestCreateCarepartnerLocationHandler_BadBody(t *testing.T) {
	controller := NewCarepartnerLocationController(&mockCarepartnerLocationService{})

	req := httptest.NewRequest(http.MethodPost, "/location", bytes.NewBufferString("bad json"))
	w := httptest.NewRecorder()

	controller.CreateCarepartnerLocationHandler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestGetCarepartnerLocationByIDHandler_Success(t *testing.T) {
	mockSvc := &mockCarepartnerLocationService{
		locationResult: &models.CarepartnerLocationPrompt{ID: 1, Location: "Home"},
	}

	controller := NewCarepartnerLocationController(mockSvc)
	req := httptest.NewRequest(http.MethodGet, "/location?id=1", nil)
	w := httptest.NewRecorder()

	controller.GetCarepartnerLocationByIDHandler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestGetCarepartnerLocationByIDHandler_InvalidID(t *testing.T) {
	controller := NewCarepartnerLocationController(&mockCarepartnerLocationService{})
	req := httptest.NewRequest(http.MethodGet, "/location?id=abc", nil)
	w := httptest.NewRecorder()

	controller.GetCarepartnerLocationByIDHandler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestUpdateCarepartnerLocationHandler_Success(t *testing.T) {
	mockSvc := &mockCarepartnerLocationService{}
	controller := NewCarepartnerLocationController(mockSvc)

	loc := models.CarepartnerLocationPrompt{ID: 5, Location: "Work", Message: "Updated"}
	body, _ := json.Marshal(loc)
	req := httptest.NewRequest(http.MethodPut, "/location", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	controller.UpdateCarepartnerLocationHandler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestUpdateCarepartnerLocationHandler_InvalidID(t *testing.T) {
	controller := NewCarepartnerLocationController(&mockCarepartnerLocationService{})

	body := `{"location": "Work", "message": "Updated"}`
	req := httptest.NewRequest(http.MethodPut, "/location", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	controller.UpdateCarepartnerLocationHandler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestDeleteCarepartnerLocationHandler_Success(t *testing.T) {
	controller := NewCarepartnerLocationController(&mockCarepartnerLocationService{})

	req := httptest.NewRequest(http.MethodDelete, "/location?id=1", nil)
	w := httptest.NewRecorder()

	controller.DeleteCarepartnerLocationHandler(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", w.Code)
	}
}

func TestDeleteCarepartnerLocationHandler_InvalidID(t *testing.T) {
	controller := NewCarepartnerLocationController(&mockCarepartnerLocationService{})
	req := httptest.NewRequest(http.MethodDelete, "/location?id=abc", nil)
	w := httptest.NewRecorder()

	controller.DeleteCarepartnerLocationHandler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}
