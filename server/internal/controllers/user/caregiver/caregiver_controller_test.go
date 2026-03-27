package caregiver

import (
	"bytes"
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	caregiverModel "server/internal/models/user/caregiver"

	"github.com/go-chi/chi/v5"
)

/* -----------------------------------------------------
   Mock Service
----------------------------------------------------- */

type mockCaregiverService struct {
	CreateCaregiverFunc             func(cg *caregiverModel.Caregiver) error
	GetCaregiverByIDFunc            func(id int) (*caregiverModel.Caregiver, error)
	GetAllCaregiversFunc            func() ([]caregiverModel.Caregiver, error)
	// GetCaregiverByCarepartnerIDFunc func(id int) (*caregiverModel.Caregiver, error)
	SearchCaregiversFunc            func(query string) ([]caregiverModel.Caregiver, error)
	UpdateCaregiverFunc             func(cg *caregiverModel.Caregiver) error
	DeleteCaregiverFunc             func(id int) error
}

func (m *mockCaregiverService) CreateCaregiver(cg *caregiverModel.Caregiver) error {
	return m.CreateCaregiverFunc(cg)
}
func (m *mockCaregiverService) GetCaregiverByID(id int) (*caregiverModel.Caregiver, error) {
	return m.GetCaregiverByIDFunc(id)
}
func (m *mockCaregiverService) GetAllCaregivers() ([]caregiverModel.Caregiver, error) {
	return m.GetAllCaregiversFunc()
}
// func (m *mockCaregiverService) GetCaregiverByCarepartnerID(id int) (*caregiverModel.Caregiver, error) {
// 	return m.GetCaregiverByCarepartnerIDFunc(id)
// }
func (m *mockCaregiverService) SearchCaregivers(query string) ([]caregiverModel.Caregiver, error) {
	return m.SearchCaregiversFunc(query)
}
func (m *mockCaregiverService) UpdateCaregiver(cg *caregiverModel.Caregiver) error {
	return m.UpdateCaregiverFunc(cg)
}
func (m *mockCaregiverService) DeleteCaregiver(id int) error {
	return m.DeleteCaregiverFunc(id)
}

/* -----------------------------------------------------
   CREATE
----------------------------------------------------- */

func TestCreateCaregiver_Success(t *testing.T) {
	mockService := &mockCaregiverService{
		CreateCaregiverFunc: func(cg *caregiverModel.Caregiver) error { return nil },
	}
	controller := NewCaregiverController(mockService)

	body := []byte(`{"id":1,"first_name":"John"}`)
	req := httptest.NewRequest("POST", "/caregivers", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	controller.CreateCaregiver(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}
}

func TestCreateCaregiver_InvalidBody(t *testing.T) {
	controller := NewCaregiverController(&mockCaregiverService{})
	req := httptest.NewRequest("POST", "/caregivers", bytes.NewBuffer([]byte("{bad json")))
	w := httptest.NewRecorder()

	controller.CreateCaregiver(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestCreateCaregiver_ServiceError(t *testing.T) {
	mockService := &mockCaregiverService{
		CreateCaregiverFunc: func(cg *caregiverModel.Caregiver) error { return errors.New("validation error") },
	}
	controller := NewCaregiverController(mockService)

	body := []byte(`{"id":1}`)
	req := httptest.NewRequest("POST", "/caregivers", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	controller.CreateCaregiver(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

/* -----------------------------------------------------
   GET BY ID
----------------------------------------------------- */

func TestGetCaregiverByID_Success(t *testing.T) {
	mockService := &mockCaregiverService{
		GetCaregiverByIDFunc: func(id int) (*caregiverModel.Caregiver, error) {
			return &caregiverModel.Caregiver{ID: id}, nil
		},
	}

	controller := NewCaregiverController(mockService)

	r := chi.NewRouter()
	r.Get("/caregivers/{id}", controller.GetCaregiverByID)

	req := httptest.NewRequest("GET", "/caregivers/1", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestGetCaregiverByID_InvalidID(t *testing.T) {
	controller := NewCaregiverController(&mockCaregiverService{})

	r := chi.NewRouter()
	r.Get("/caregivers/{id}", controller.GetCaregiverByID)

	req := httptest.NewRequest("GET", "/caregivers/abc", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestGetCaregiverByID_NotFound(t *testing.T) {
	mockService := &mockCaregiverService{
		GetCaregiverByIDFunc: func(id int) (*caregiverModel.Caregiver, error) { return nil, nil },
	}
	controller := NewCaregiverController(mockService)

	r := chi.NewRouter()
	r.Get("/caregivers/{id}", controller.GetCaregiverByID)

	req := httptest.NewRequest("GET", "/caregivers/1", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestGetCaregiverByID_ServiceError(t *testing.T) {
	mockService := &mockCaregiverService{
		GetCaregiverByIDFunc: func(id int) (*caregiverModel.Caregiver, error) {
			return nil, errors.New("db error")
		},
	}
	controller := NewCaregiverController(mockService)

	r := chi.NewRouter()
	r.Get("/caregivers/{id}", controller.GetCaregiverByID)

	req := httptest.NewRequest("GET", "/caregivers/1", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}

/* -----------------------------------------------------
   GET ALL
----------------------------------------------------- */

func TestGetAllCaregivers_Success(t *testing.T) {
	mockService := &mockCaregiverService{
		GetAllCaregiversFunc: func() ([]caregiverModel.Caregiver, error) {
			return []caregiverModel.Caregiver{{ID: 1}}, nil
		},
	}

	controller := NewCaregiverController(mockService)

	req := httptest.NewRequest("GET", "/caregivers", nil)
	w := httptest.NewRecorder()

	controller.GetAllCaregivers(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestGetAllCaregivers_ServiceError(t *testing.T) {
	mockService := &mockCaregiverService{
		GetAllCaregiversFunc: func() ([]caregiverModel.Caregiver, error) {
			return nil, errors.New("db error")
		},
	}

	controller := NewCaregiverController(mockService)

	req := httptest.NewRequest("GET", "/caregivers", nil)
	w := httptest.NewRecorder()

	controller.GetAllCaregivers(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

/* -----------------------------------------------------
   GET BY Carepartner ID
----------------------------------------------------- */

// func TestGetCaregiverByCarepartnerID_Success(t *testing.T) {
// 	mockService := &mockCaregiverService{
// 		GetCaregiverByCarepartnerIDFunc: func(id int) (*caregiverModel.Caregiver, error) {
// 			return &caregiverModel.Caregiver{ID: id}, nil
// 		},
// 	}
// 	controller := NewCaregiverController(mockService)

// 	r := chi.NewRouter()
// 	r.Get("/caregivers/carepartner/{CarepartnerId}", controller.GetCaregiverByCarepartnerID)

// 	req := httptest.NewRequest("GET", "/caregivers/carepartner/1", nil)
// 	w := httptest.NewRecorder()

// 	r.ServeHTTP(w, req)

// 	if w.Code != http.StatusOK {
// 		t.Fatalf("expected 200, got %d", w.Code)
// 	}
// }

// func TestGetCaregiverByCarepartnerID_InvalidID(t *testing.T) {
// 	controller := NewCaregiverController(&mockCaregiverService{})

// 	r := chi.NewRouter()
// 	r.Get("/caregivers/carepartner/{CarepartnerId}", controller.GetCaregiverByCarepartnerID)

// 	req := httptest.NewRequest("GET", "/caregivers/carepartner/abc", nil)
// 	w := httptest.NewRecorder()

// 	r.ServeHTTP(w, req)

// 	if w.Code != http.StatusBadRequest {
// 		t.Errorf("expected 400, got %d", w.Code)
// 	}
// }

// func TestGetCaregiverByCarepartnerID_NotFound(t *testing.T) {
// 	mockService := &mockCaregiverService{
// 		GetCaregiverByCarepartnerIDFunc: func(id int) (*caregiverModel.Caregiver, error) {
// 			return nil, nil
// 		},
// 	}

// 	controller := NewCaregiverController(mockService)

// 	r := chi.NewRouter()
// 	r.Get("/caregivers/carepartner/{CarepartnerId}", controller.GetCaregiverByCarepartnerID)

// 	req := httptest.NewRequest("GET", "/caregivers/carepartner/1", nil)
// 	w := httptest.NewRecorder()

// 	r.ServeHTTP(w, req)

// 	if w.Code != http.StatusNotFound {
// 		t.Errorf("expected 404, got %d", w.Code)
// 	}
// }

// func TestGetCaregiverByCarepartnerID_ServiceError(t *testing.T) {
// 	mockService := &mockCaregiverService{
// 		GetCaregiverByCarepartnerIDFunc: func(id int) (*caregiverModel.Caregiver, error) {
// 			return nil, errors.New("db error")
// 		},
// 	}

// 	controller := NewCaregiverController(mockService)

// 	r := chi.NewRouter()
// 	r.Get("/caregivers/carepartner/{CarepartnerId}", controller.GetCaregiverByCarepartnerID)

// 	req := httptest.NewRequest("GET", "/caregivers/carepartner/1", nil)
// 	w := httptest.NewRecorder()

// 	r.ServeHTTP(w, req)

// 	if w.Code != http.StatusInternalServerError {
// 		t.Errorf("expected 500, got %d", w.Code)
// 	}
// }

/* -----------------------------------------------------
   SEARCH
----------------------------------------------------- */

func TestSearchCaregivers_Success(t *testing.T) {
	mockService := &mockCaregiverService{
		SearchCaregiversFunc: func(q string) ([]caregiverModel.Caregiver, error) {
			return []caregiverModel.Caregiver{{ID: 1}}, nil
		},
	}

	controller := NewCaregiverController(mockService)

	req := httptest.NewRequest("GET", "/caregivers/search?q=jo", nil)
	w := httptest.NewRecorder()

	controller.SearchCaregivers(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestSearchCaregivers_QueryTooShort(t *testing.T) {
	controller := NewCaregiverController(&mockCaregiverService{})

	req := httptest.NewRequest("GET", "/caregivers/search?q=j", nil)
	w := httptest.NewRecorder()

	controller.SearchCaregivers(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestSearchCaregivers_ServiceError(t *testing.T) {
	mockService := &mockCaregiverService{
		SearchCaregiversFunc: func(q string) ([]caregiverModel.Caregiver, error) {
			return nil, errors.New("db error")
		},
	}

	controller := NewCaregiverController(mockService)

	req := httptest.NewRequest("GET", "/caregivers/search?q=john", nil)
	w := httptest.NewRecorder()

	controller.SearchCaregivers(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}

/* -----------------------------------------------------
   UPDATE
----------------------------------------------------- */

func TestUpdateCaregiver_Success(t *testing.T) {
	mockService := &mockCaregiverService{
		UpdateCaregiverFunc: func(cg *caregiverModel.Caregiver) error { return nil },
	}

	controller := NewCaregiverController(mockService)

	body := []byte(`{"id":1,"first_name":"Updated"}`)
	req := httptest.NewRequest("PUT", "/caregivers", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	controller.UpdateCaregiver(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestUpdateCaregiver_InvalidBody(t *testing.T) {
	controller := NewCaregiverController(&mockCaregiverService{})

	req := httptest.NewRequest("PUT", "/caregivers", bytes.NewBuffer([]byte("{bad json")))
	w := httptest.NewRecorder()

	controller.UpdateCaregiver(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestUpdateCaregiver_MissingID(t *testing.T) {
	controller := NewCaregiverController(&mockCaregiverService{})

	body := []byte(`{"first_name":"Name"}`)
	req := httptest.NewRequest("PUT", "/caregivers", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	controller.UpdateCaregiver(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestUpdateCaregiver_ServiceError(t *testing.T) {
	mockService := &mockCaregiverService{
		UpdateCaregiverFunc: func(cg *caregiverModel.Caregiver) error {
			return errors.New("update failed")
		},
	}

	controller := NewCaregiverController(mockService)

	body := []byte(`{"id":1}`)
	req := httptest.NewRequest("PUT", "/caregivers", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	controller.UpdateCaregiver(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

/* -----------------------------------------------------
   DELETE
----------------------------------------------------- */

func TestDeleteCaregiver_Success(t *testing.T) {
	mockService := &mockCaregiverService{
		DeleteCaregiverFunc: func(id int) error { return nil },
	}

	controller := NewCaregiverController(mockService)

	r := chi.NewRouter()
	r.Delete("/caregivers/{id}", controller.DeleteCaregiver)

	req := httptest.NewRequest("DELETE", "/caregivers/1", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestDeleteCaregiver_InvalidID(t *testing.T) {
	controller := NewCaregiverController(&mockCaregiverService{})

	r := chi.NewRouter()
	r.Delete("/caregivers/{id}", controller.DeleteCaregiver)

	req := httptest.NewRequest("DELETE", "/caregivers/abc", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestDeleteCaregiver_NotFound(t *testing.T) {
	mockService := &mockCaregiverService{
		DeleteCaregiverFunc: func(id int) error { return sql.ErrNoRows },
	}

	controller := NewCaregiverController(mockService)

	r := chi.NewRouter()
	r.Delete("/caregivers/{id}", controller.DeleteCaregiver)

	req := httptest.NewRequest("DELETE", "/caregivers/1", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestDeleteCaregiver_ServiceError(t *testing.T) {
	mockService := &mockCaregiverService{
		DeleteCaregiverFunc: func(id int) error { return errors.New("unknown error") },
	}

	controller := NewCaregiverController(mockService)

	r := chi.NewRouter()
	r.Delete("/caregivers/{id}", controller.DeleteCaregiver)

	req := httptest.NewRequest("DELETE", "/caregivers/1", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}
