package modules

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	modules "server/internal/models/modules"

	"github.com/go-chi/chi/v5"
)

// --- Mock service implementing ModuleServiceInterface ---
type mockModuleService struct {
	CreateModuleFunc                             func(moduleType string, caregiverID, carepartnerID, metadataID int) (*modules.Module, error)
	GetModuleByIDFunc                            func(id int) (*modules.Module, error)
	GetMetadataByCaregiverCarepartnerAndTypeFunc func(caregiverID, carepartnerID int, moduleType string) ([]int, error)
	GetAllMetadataByCaregiverCarepartnerFunc     func(caregiverID, carepartnerID int) ([]int, error)
	DeleteModuleFunc							 func(id int) error
}

func (m *mockModuleService) CreateModule(moduleType string, caregiverID, carepartnerID, metadataID int) (*modules.Module, error) {
	if m.CreateModuleFunc != nil {
		return m.CreateModuleFunc(moduleType, caregiverID, carepartnerID, metadataID)
	}
	return nil, nil
}

func (m *mockModuleService) CreateModuleWithMedia(moduleType string, caregiverID, carepartnerID int, uri, filename, extension string) (*modules.Module, error) {
	// no longer called by controller
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

func (m *mockModuleService) GetAllMetadataByCaregiverCarepartner(caregiverID, carepartnerID int) ([]int, error) {
	if m.GetAllMetadataByCaregiverCarepartnerFunc != nil {
		return m.GetAllMetadataByCaregiverCarepartnerFunc(caregiverID, carepartnerID)
	}
	return nil, nil
}

func (m *mockModuleService) DeleteModule(id int) error {
    return nil
}

func (m *mockModuleService) GetModulesByCaregiverAndCarepartner(caregiverID, carepartnerID int) ([]*modules.Module, error) {
	return nil, nil
}

// --- Tests ---

func TestCreateModuleHandler_Success(t *testing.T) {
	mockService := &mockModuleService{
		CreateModuleFunc: func(moduleType string, caregiverID, carepartnerID, metadataID int) (*modules.Module, error) {
			return &modules.Module{
				ID:         1,
				ModuleType: moduleType,
				MetadataID: metadataID,
			}, nil
		},
	}

	controller := NewModuleController(mockService)

	reqBody := map[string]interface{}{
		"module_type":    "video",
		"caregiver_id":   2,
		"carepartner_id": 3,
		"metadata_id":    42,
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/module", bytes.NewBuffer(body))
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
	if mod.MetadataID != 42 {
		t.Errorf("expected MetadataID 42, got %d", mod.MetadataID)
	}
}

func TestCreateModuleHandler_InvalidBody(t *testing.T) {
	controller := NewModuleController(&mockModuleService{})

	req := httptest.NewRequest("POST", "/module", bytes.NewBuffer([]byte("{invalid-json")))
	w := httptest.NewRecorder()

	controller.CreateModuleHandler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestCreateModuleHandler_ServiceError(t *testing.T) {
	mockService := &mockModuleService{
		CreateModuleFunc: func(moduleType string, caregiverID, carepartnerID, metadataID int) (*modules.Module, error) {
			return nil, errors.New("database failure")
		},
	}

	controller := NewModuleController(mockService)

	reqBody := map[string]interface{}{
		"module_type":    "photo",
		"caregiver_id":   1,
		"carepartner_id": 1,
		"metadata_id":    99,
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/module", bytes.NewBuffer(body))
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

	controller := NewModuleController(mockService)

	r := chi.NewRouter()
	r.Get("/module/{id}", controller.GetModuleByIDHandler)

	req := httptest.NewRequest("GET", "/module/1", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

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
	controller := NewModuleController(&mockModuleService{})

	r := chi.NewRouter()
	r.Get("/module/{id}", controller.GetModuleByIDHandler)

	req := httptest.NewRequest("GET", "/module/abc", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

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

	controller := NewModuleController(mockService)

	r := chi.NewRouter()
	r.Get("/module/{id}", controller.GetModuleByIDHandler)

	req := httptest.NewRequest("GET", "/module/1", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

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

	controller := NewModuleController(mockService)

	req := httptest.NewRequest("GET", "/module/metadata?caregiver_id=1&carepartner_id=2&module_type=photo", nil)
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

	controller := NewModuleController(mockService)

	req := httptest.NewRequest("GET", "/module/metadata?caregiver_id=1&carepartner_id=2&module_type=video", nil)
	w := httptest.NewRecorder()

	controller.GetMetadataHandler(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}
