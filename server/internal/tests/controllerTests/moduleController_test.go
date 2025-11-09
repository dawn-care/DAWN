package media_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	controllers "server/internal/controllers/modules"
	modules "server/internal/models/modules"
)

// Mock service implementing ModuleServiceInterface
type mockModuleService struct {
	CreateModuleWithMediaFunc                   func(moduleType string, caregiverID, carepartnerID int, uri, filename, extension string) (*modules.Module, error)
	GetModuleByIDFunc                           func(id int) (*modules.Module, error)
	GetMetadataByCaregiverCarepartnerAndTypeFunc func(caregiverID, carepartnerID int, moduleType string) ([]int, error)
}

func (m *mockModuleService) CreateModuleWithMedia(moduleType string, caregiverID, carepartnerID int, uri, filename, extension string) (*modules.Module, error) {
	if m.CreateModuleWithMediaFunc != nil {
		return m.CreateModuleWithMediaFunc(moduleType, caregiverID, carepartnerID, uri, filename, extension)
	}
	return nil, nil
}

func (m *mockModuleService) GetModuleByID(id int) (*modules.Module, error) {
	if m.GetModuleByIDFunc != nil {
		return m.GetModuleByIDFunc(id)
	}
	return nil, nil
}

func (m *mockModuleService) GetMetadataByCaregiverCarepartnerAndType(caregiverID, carepartnerID int, moduleType string) ([]int, error) {
	if m.GetMetadataByCaregiverCarepartnerAndTypeFunc != nil {
		return m.GetMetadataByCaregiverCarepartnerAndTypeFunc(caregiverID, carepartnerID, moduleType)
	}
	return nil, nil
}

// --- Tests ---

func TestCreateModuleHandler_Success(t *testing.T) {
	mockService := &mockModuleService{
		CreateModuleWithMediaFunc: func(moduleType string, caregiverID, carepartnerID int, uri, filename, extension string) (*modules.Module, error) {
			return &modules.Module{ID: 1, ModuleType: moduleType}, nil
		},
	}
	controller := controllers.NewModuleController(mockService)

	reqBody := map[string]interface{}{
		"module_type":    "video",
		"caregiver_id":   2,
		"carepartner_id": 3,
		"uri":            "test.mp4",
		"filename":       "testfile",
		"extension":      "mp4",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/modules", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	controller.CreateModuleHandler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var mod modules.Module
	if err := json.NewDecoder(w.Body).Decode(&mod); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if mod.ModuleType != "video" {
		t.Errorf("expected ModuleType 'video', got %s", mod.ModuleType)
	}
}

func TestCreateModuleHandler_InvalidBody(t *testing.T) {
	controller := controllers.NewModuleController(&mockModuleService{})
	req := httptest.NewRequest("POST", "/modules", bytes.NewBuffer([]byte("{invalid-json")))
	w := httptest.NewRecorder()

	controller.CreateModuleHandler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestCreateModuleHandler_ServiceError(t *testing.T) {
	mockService := &mockModuleService{
		CreateModuleWithMediaFunc: func(moduleType string, caregiverID, carepartnerID int, uri, filename, extension string) (*modules.Module, error) {
			return nil, errors.New("database failure")
		},
	}
	controller := controllers.NewModuleController(mockService)

	reqBody := map[string]interface{}{
		"module_type":    "photo",
		"caregiver_id":   1,
		"carepartner_id": 1,
		"uri":            "path/to/photo.jpg",
		"filename":       "photo",
		"extension":      "jpg",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/modules", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	controller.CreateModuleHandler(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}

func TestGetModuleByIDHandler_Success(t *testing.T) {
	mockService := &mockModuleService{
		GetModuleByIDFunc: func(id int) (*modules.Module, error) {
			return &modules.Module{ID: id, ModuleType: "video"}, nil
		},
	}
	controller := controllers.NewModuleController(mockService)
	req := httptest.NewRequest("GET", "/modules?id=1", nil)
	w := httptest.NewRecorder()

	controller.GetModuleByIDHandler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var mod modules.Module
	json.NewDecoder(w.Body).Decode(&mod)
	if mod.ID != 1 {
		t.Errorf("expected module ID 1, got %d", mod.ID)
	}
}

func TestGetModuleByIDHandler_InvalidID(t *testing.T) {
	controller := controllers.NewModuleController(&mockModuleService{})
	req := httptest.NewRequest("GET", "/modules?id=abc", nil)
	w := httptest.NewRecorder()

	controller.GetModuleByIDHandler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestGetModuleByIDHandler_ServiceError(t *testing.T) {
	mockService := &mockModuleService{
		GetModuleByIDFunc: func(id int) (*modules.Module, error) {
			return nil, errors.New("not found")
		},
	}
	controller := controllers.NewModuleController(mockService)
	req := httptest.NewRequest("GET", "/modules?id=1", nil)
	w := httptest.NewRecorder()

	controller.GetModuleByIDHandler(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}

func TestGetMetadataHandler_Success(t *testing.T) {
	mockService := &mockModuleService{
		GetMetadataByCaregiverCarepartnerAndTypeFunc: func(caregiverID, carepartnerID int, moduleType string) ([]int, error) {
			return []int{1, 2, 3}, nil
		},
	}
	controller := controllers.NewModuleController(mockService)
	req := httptest.NewRequest("GET", "/metadata?caregiver_id=1&carepartner_id=2&module_type=photo", nil)
	w := httptest.NewRecorder()

	controller.GetMetadataHandler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var ids []int
	json.NewDecoder(w.Body).Decode(&ids)
	if len(ids) != 3 {
		t.Errorf("expected 3 IDs, got %d", len(ids))
	}
}

func TestGetMetadataHandler_ServiceError(t *testing.T) {
	mockService := &mockModuleService{
		GetMetadataByCaregiverCarepartnerAndTypeFunc: func(caregiverID, carepartnerID int, moduleType string) ([]int, error) {
			return nil, errors.New("db error")
		},
	}
	controller := controllers.NewModuleController(mockService)
	req := httptest.NewRequest("GET", "/metadata?caregiver_id=1&carepartner_id=2&module_type=video", nil)
	w := httptest.NewRecorder()

	controller.GetMetadataHandler(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}
