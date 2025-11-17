package prompts_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	controllers "server/internal/controllers/modules/prompts"
	models "server/internal/models/modules/prompts"
)

// Mock service implementing CaregiverLocationServiceInterface
type mockCaregiverLocationService struct {
	createErr error
	getErr    error
	updateErr error
	deleteErr error

	locationResult *models.CaregiverLocationPrompt
}

func (m *mockCaregiverLocationService) CreateCaregiverLocation(loc *models.CaregiverLocationPrompt) error {
	return m.createErr
}

func (m *mockCaregiverLocationService) GetCaregiverLocationByID(id int) (*models.CaregiverLocationPrompt, error) {
	return m.locationResult, m.getErr
}

func (m *mockCaregiverLocationService) UpdateCaregiverLocation(loc *models.CaregiverLocationPrompt) error {
	return m.updateErr
}

func (m *mockCaregiverLocationService) DeleteCaregiverLocation(id int) error {
	return m.deleteErr
}

// --- Tests ---

func TestCreateCaregiverLocationHandler_Success(t *testing.T) {
	mockSvc := &mockCaregiverLocationService{}
	controller := controllers.NewCaregiverLocationController(mockSvc)

	body := `{"location": "Home", "message": "Arrived"}`
	req := httptest.NewRequest(http.MethodPost, "/location", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	controller.CreateCaregiverLocationHandler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestCreateCaregiverLocationHandler_BadBody(t *testing.T) {
	controller := controllers.NewCaregiverLocationController(&mockCaregiverLocationService{})

	req := httptest.NewRequest(http.MethodPost, "/location", bytes.NewBufferString("bad json"))
	w := httptest.NewRecorder()

	controller.CreateCaregiverLocationHandler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestGetCaregiverLocationByIDHandler_Success(t *testing.T) {
	mockSvc := &mockCaregiverLocationService{
		locationResult: &models.CaregiverLocationPrompt{ID: 1, Location: "Home"},
	}

	controller := controllers.NewCaregiverLocationController(mockSvc)
	req := httptest.NewRequest(http.MethodGet, "/location?id=1", nil)
	w := httptest.NewRecorder()

	controller.GetCaregiverLocationByIDHandler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestGetCaregiverLocationByIDHandler_InvalidID(t *testing.T) {
	controller := controllers.NewCaregiverLocationController(&mockCaregiverLocationService{})
	req := httptest.NewRequest(http.MethodGet, "/location?id=abc", nil)
	w := httptest.NewRecorder()

	controller.GetCaregiverLocationByIDHandler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestUpdateCaregiverLocationHandler_Success(t *testing.T) {
	mockSvc := &mockCaregiverLocationService{}
	controller := controllers.NewCaregiverLocationController(mockSvc)

	loc := models.CaregiverLocationPrompt{ID: 5, Location: "Work", Message: "Updated"}
	body, _ := json.Marshal(loc)
	req := httptest.NewRequest(http.MethodPut, "/location", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	controller.UpdateCaregiverLocationHandler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestUpdateCaregiverLocationHandler_InvalidID(t *testing.T) {
	controller := controllers.NewCaregiverLocationController(&mockCaregiverLocationService{})

	body := `{"location": "Work", "message": "Updated"}`
	req := httptest.NewRequest(http.MethodPut, "/location", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	controller.UpdateCaregiverLocationHandler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestDeleteCaregiverLocationHandler_Success(t *testing.T) {
	controller := controllers.NewCaregiverLocationController(&mockCaregiverLocationService{})

	req := httptest.NewRequest(http.MethodDelete, "/location?id=1", nil)
	w := httptest.NewRecorder()

	controller.DeleteCaregiverLocationHandler(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", w.Code)
	}
}

func TestDeleteCaregiverLocationHandler_InvalidID(t *testing.T) {
	controller := controllers.NewCaregiverLocationController(&mockCaregiverLocationService{})
	req := httptest.NewRequest(http.MethodDelete, "/location?id=abc", nil)
	w := httptest.NewRecorder()

	controller.DeleteCaregiverLocationHandler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}
