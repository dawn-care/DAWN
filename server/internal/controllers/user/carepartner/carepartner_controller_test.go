package carepartner

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"encoding/json"
	"testing"

	model "server/internal/models/user/carepartner"

	"github.com/go-chi/chi/v5"
)

/* -----------------------------------------------------
   Mock Service
----------------------------------------------------- */

type mockCarepartnerService struct {
	CreateCarepartnerFunc  func(c *model.Carepartner) error
	UpdateCarepartnerFunc  func(c *model.Carepartner) error
	GetCarepartnerByIDFunc func(id int) (*model.Carepartner, error)
	GetAllCarepartnersFunc func() ([]model.Carepartner, error)
	DeleteCarepartnerFunc  func(id int) error
}

func (m *mockCarepartnerService) CreateCarepartner(c *model.Carepartner) error {
	return m.CreateCarepartnerFunc(c)
}
func (m *mockCarepartnerService) UpdateCarepartner(c *model.Carepartner) error {
	return m.UpdateCarepartnerFunc(c)
}
func (m *mockCarepartnerService) GetCarepartnerByID(id int) (*model.Carepartner, error) {
	return m.GetCarepartnerByIDFunc(id)
}
func (m *mockCarepartnerService) GetAllCarepartners() ([]model.Carepartner, error) {
	return m.GetAllCarepartnersFunc()
}
func (m *mockCarepartnerService) DeleteCarepartner(id int) error {
	return m.DeleteCarepartnerFunc(id)
}

/* -----------------------------------------------------
   CREATE
----------------------------------------------------- */

func TestCreateCarepartner_Success(t *testing.T) {
	var created *model.Carepartner
	mockService := &mockCarepartnerService{
		CreateCarepartnerFunc: func(c *model.Carepartner) error {
			c.ID = 7
			created = c
			return nil
		},
	}

	controller := NewCarepartnerController(mockService)

	body := []byte(`{"name":"Ray"}`)
	req := httptest.NewRequest("POST", "/carepartners", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	controller.CreateCarepartner(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}
	if created.Name != "Ray" || created.CaregiverID != 0 {
		t.Errorf("unexpected carepartner %+v", created)
	}
	var resp map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["carepartner_id"] != float64(7) {
		t.Errorf("expected carepartner_id 7, got %v", resp["carepartner_id"])
	}
	if _, ok := resp["auto_start"]; ok {
		t.Error("auto_start should no longer be returned")
	}
}

// The tablet's pairing screen still sends AutoStart; it must be ignored, not rejected
func TestCreateCarepartner_IgnoresLegacyAutoStart(t *testing.T) {
	mockService := &mockCarepartnerService{
		CreateCarepartnerFunc: func(c *model.Carepartner) error { return nil },
	}

	controller := NewCarepartnerController(mockService)

	body := []byte(`{"Name":"PlaceHolder Name","AutoStart":"08:00:00"}`)
	req := httptest.NewRequest("POST", "/carepartners", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	controller.CreateCarepartner(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}
}

func TestCreateCarepartner_InvalidBody(t *testing.T) {
	controller := NewCarepartnerController(&mockCarepartnerService{})

	req := httptest.NewRequest("POST", "/carepartners", bytes.NewBuffer([]byte("{bad json")))
	w := httptest.NewRecorder()

	controller.CreateCarepartner(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestCreateCarepartner_ServiceError(t *testing.T) {
	mockService := &mockCarepartnerService{
		CreateCarepartnerFunc: func(c *model.Carepartner) error {
			return errors.New("service error")
		},
	}

	controller := NewCarepartnerController(mockService)

	body := []byte(`{"name":"Ray"}`)
	req := httptest.NewRequest("POST", "/carepartners", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	controller.CreateCarepartner(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}

/* -----------------------------------------------------
   UPDATE
----------------------------------------------------- */

func TestUpdateCarepartner_Success(t *testing.T) {
	var updated *model.Carepartner
	mockService := &mockCarepartnerService{
		GetCarepartnerByIDFunc: func(id int) (*model.Carepartner, error) {
			return &model.Carepartner{ID: id, Name: "Old", CaregiverID: 2}, nil
		},
		UpdateCarepartnerFunc: func(c *model.Carepartner) error {
			updated = c
			return nil
		},
	}

	controller := NewCarepartnerController(mockService)

	body := []byte(`{"id":1,"name":"Updated"}`)
	req := httptest.NewRequest("PUT", "/carepartners", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	controller.UpdateCarepartner(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if updated.Name != "Updated" || updated.CaregiverID != 2 {
		t.Errorf("expected new name and kept caregiver, got %+v", updated)
	}
}

func TestUpdateCarepartner_InvalidBody(t *testing.T) {
	controller := NewCarepartnerController(&mockCarepartnerService{})

	req := httptest.NewRequest("PUT", "/carepartners", bytes.NewBuffer([]byte("{bad json")))
	w := httptest.NewRecorder()

	controller.UpdateCarepartner(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestUpdateCarepartner_MissingID(t *testing.T) {
	controller := NewCarepartnerController(&mockCarepartnerService{})

	body := []byte(`{"name":"NoID"}`)
	req := httptest.NewRequest("PUT", "/carepartners", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	controller.UpdateCarepartner(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestUpdateCarepartner_ServiceError(t *testing.T) {
	mockService := &mockCarepartnerService{
		GetCarepartnerByIDFunc: func(id int) (*model.Carepartner, error) {
			return &model.Carepartner{ID: id, Name: "Old", CaregiverID: 2}, nil
		},
		UpdateCarepartnerFunc: func(c *model.Carepartner) error {
			return errors.New("update failed")
		},
	}

	controller := NewCarepartnerController(mockService)

	body := []byte(`{"id":1,"name":"Fail"}`)
	req := httptest.NewRequest("PUT", "/carepartners", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	controller.UpdateCarepartner(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

/* -----------------------------------------------------
   GET BY ID
----------------------------------------------------- */

func TestGetCarepartnerByID_Success(t *testing.T) {
	mockService := &mockCarepartnerService{
		GetCarepartnerByIDFunc: func(id int) (*model.Carepartner, error) {
			return &model.Carepartner{
				ID:       id,
				Name:     "Ray",
				Timezone: "America/New_York",
			}, nil
		},
	}

	controller := NewCarepartnerController(mockService)

	r := chi.NewRouter()
	r.Get("/carepartners/{id}", controller.GetCarepartnerByID)

	req := httptest.NewRequest("GET", "/carepartners/1", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestGetCarepartnerByID_InvalidID(t *testing.T) {
	controller := NewCarepartnerController(&mockCarepartnerService{})

	r := chi.NewRouter()
	r.Get("/carepartners/{id}", controller.GetCarepartnerByID)

	req := httptest.NewRequest("GET", "/carepartners/abc", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestGetCarepartnerByID_NotFound(t *testing.T) {
	mockService := &mockCarepartnerService{
		GetCarepartnerByIDFunc: func(id int) (*model.Carepartner, error) {
			return nil, nil
		},
	}

	controller := NewCarepartnerController(mockService)

	r := chi.NewRouter()
	r.Get("/carepartners/{id}", controller.GetCarepartnerByID)

	req := httptest.NewRequest("GET", "/carepartners/1", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestGetCarepartnerByID_ServiceError(t *testing.T) {
	mockService := &mockCarepartnerService{
		GetCarepartnerByIDFunc: func(id int) (*model.Carepartner, error) {
			return nil, errors.New("db error")
		},
	}

	controller := NewCarepartnerController(mockService)

	r := chi.NewRouter()
	r.Get("/carepartners/{id}", controller.GetCarepartnerByID)

	req := httptest.NewRequest("GET", "/carepartners/1", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}

/* -----------------------------------------------------
   GET ALL
----------------------------------------------------- */

func TestGetAllCarepartners_Success(t *testing.T) {
	mockService := &mockCarepartnerService{
		GetAllCarepartnersFunc: func() ([]model.Carepartner, error) {
			return []model.Carepartner{
				{ID: 1, Name: "A"},
			}, nil
		},
	}

	controller := NewCarepartnerController(mockService)

	req := httptest.NewRequest("GET", "/carepartners", nil)
	w := httptest.NewRecorder()

	controller.GetAllCarepartners(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestGetAllCarepartners_ServiceError(t *testing.T) {
	mockService := &mockCarepartnerService{
		GetAllCarepartnersFunc: func() ([]model.Carepartner, error) {
			return nil, errors.New("db error")
		},
	}

	controller := NewCarepartnerController(mockService)

	req := httptest.NewRequest("GET", "/carepartners", nil)
	w := httptest.NewRecorder()

	controller.GetAllCarepartners(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

/* -----------------------------------------------------
   DELETE
----------------------------------------------------- */

func TestDeleteCarepartner_Success(t *testing.T) {
	mockService := &mockCarepartnerService{
		DeleteCarepartnerFunc: func(id int) error { return nil },
	}

	controller := NewCarepartnerController(mockService)

	r := chi.NewRouter()
	r.Delete("/carepartners/{id}", controller.DeleteCarepartner)

	req := httptest.NewRequest("DELETE", "/carepartners/1", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestDeleteCarepartner_InvalidID(t *testing.T) {
	controller := NewCarepartnerController(&mockCarepartnerService{})

	r := chi.NewRouter()
	r.Delete("/carepartners/{id}", controller.DeleteCarepartner)

	req := httptest.NewRequest("DELETE", "/carepartners/abc", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestDeleteCarepartner_ServiceError(t *testing.T) {
	mockService := &mockCarepartnerService{
		DeleteCarepartnerFunc: func(id int) error {
			return errors.New("delete failed")
		},
	}

	controller := NewCarepartnerController(mockService)

	r := chi.NewRouter()
	r.Delete("/carepartners/{id}", controller.DeleteCarepartner)

	req := httptest.NewRequest("DELETE", "/carepartners/1", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}
