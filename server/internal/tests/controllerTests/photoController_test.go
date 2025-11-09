package media_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	controllers "server/internal/controllers/modules/media"
	models "server/internal/models/modules/media"
)

// Mock service implementing PhotoServiceInterface
type mockPhotoService struct {
	createErr   error
	getErr      error
	updateErr   error
	deleteErr   error
	statsErr    error
	photoResult *models.PhotoMedia
	photoList   []models.PhotoMedia
	statsResult int
}

func (m *mockPhotoService) CreatePhoto(p *models.PhotoMedia) error                        			 { return m.createErr }
func (m *mockPhotoService) GetPhotoByID(id int) (*models.PhotoMedia, error)                			 { return m.photoResult, m.getErr }
func (m *mockPhotoService) GetPhotosByCarepartnerAndCaregiver(c, g int) ([]models.PhotoMedia, error) { return m.photoList, m.getErr }
func (m *mockPhotoService) UpdatePhoto(p *models.PhotoMedia) error 									 { return m.updateErr }
func (m *mockPhotoService) DeletePhoto(id int) error               									 { return m.deleteErr }
func (m *mockPhotoService) GetPhotoStats() (int, error)           									 { return m.statsResult, m.statsErr }

// --- Tests ---

func TestCreatePhotoHandler_Success(t *testing.T) {
	mockSvc := &mockPhotoService{}
	controller := controllers.NewPhotoController(mockSvc)

	photo := models.PhotoMedia{ID: 1, Filename: "pic.jpg", URI: "/photos/pic.jpg"}
	body, _ := json.Marshal(photo)
	req := httptest.NewRequest(http.MethodPost, "/photo", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	controller.CreatePhotoHandler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestGetPhotoByIDHandler_Success(t *testing.T) {
	mockSvc := &mockPhotoService{
		photoResult: &models.PhotoMedia{ID: 1, Filename: "image.jpg"},
	}
	controller := controllers.NewPhotoController(mockSvc)

	req := httptest.NewRequest(http.MethodGet, "/photo?id=1", nil)
	w := httptest.NewRecorder()

	controller.GetPhotoByIDHandler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestGetPhotoByIDHandler_InvalidID(t *testing.T) {
	controller := controllers.NewPhotoController(&mockPhotoService{})
	req := httptest.NewRequest(http.MethodGet, "/photo?id=abc", nil)
	w := httptest.NewRecorder()

	controller.GetPhotoByIDHandler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestDeletePhotoHandler_Success(t *testing.T) {
	controller := controllers.NewPhotoController(&mockPhotoService{})
	req := httptest.NewRequest(http.MethodDelete, "/photo?id=1", nil)
	w := httptest.NewRecorder()

	controller.DeletePhotoHandler(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", w.Code)
	}
}

func TestUpdatePhotoHandler_Success(t *testing.T) {
	controller := controllers.NewPhotoController(&mockPhotoService{})
	photo := models.PhotoMedia{ID: 2, Filename: "update.jpg"}
	body, _ := json.Marshal(photo)
	req := httptest.NewRequest(http.MethodPut, "/photo", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	controller.UpdatePhotoHandler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestGetPhotosByCarepartnerAndCaregiverHandler_Success(t *testing.T) {
	mockSvc := &mockPhotoService{
		photoList: []models.PhotoMedia{
			{ID: 1, Filename: "p1.jpg"},
			{ID: 2, Filename: "p2.jpg"},
		},
	}
	controller := controllers.NewPhotoController(mockSvc)
	req := httptest.NewRequest(http.MethodGet, "/photos?carepartner_id=1&caregiver_id=1", nil)
	w := httptest.NewRecorder()

	controller.GetPhotosByCarepartnerAndCaregiverHandler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestGetPhotoStatsHandler_Error(t *testing.T) {
	mockSvc := &mockPhotoService{statsErr: errors.New("db error")}
	controller := controllers.NewPhotoController(mockSvc)

	req := httptest.NewRequest(http.MethodGet, "/stats", nil)
	w := httptest.NewRecorder()

	controller.GetPhotoStatsHandler(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestGetPhotoStatsHandler_Success(t *testing.T) {
	mockSvc := &mockPhotoService{statsResult: 5}
	controller := controllers.NewPhotoController(mockSvc)

	req := httptest.NewRequest(http.MethodGet, "/stats", nil)
	w := httptest.NewRecorder()

	controller.GetPhotoStatsHandler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}
