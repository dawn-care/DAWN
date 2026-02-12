package media

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"server/internal/models/media"
)

// Mock service implementing PhotoServiceInterface
type mockPhotoService struct {
	createErr   error
	getErr      error
	updateErr   error
	deleteErr   error
	photoResult *media.PhotoMedia
	photoFile   multipart.File
	photoList   []media.PhotoMedia
}

// ---- Interface methods ----

func (m *mockPhotoService) CreatePhoto(p *media.PhotoMedia) error {
	return m.createErr
}

func (m *mockPhotoService) CreatePhotoFile(file multipart.File, handler *multipart.FileHeader, uri string) error {
	return m.createErr
}

func (m *mockPhotoService) GetPhotoByID(id int) (*media.PhotoMedia, error) {
	return m.photoResult, m.getErr
}

func (m *mockPhotoService) GetPhotoFile(uri string) (multipart.File, error) {
	return m.photoFile, m.getErr
}

func (m *mockPhotoService) GetPhotosByCarepartnerAndCaregiver(c, g int) ([]media.PhotoMedia, error) {
	return m.photoList, m.getErr
}

func (m *mockPhotoService) UpdatePhoto(p *media.PhotoMedia) error {
	return m.updateErr
}

func (m *mockPhotoService) UpdatePhotoFile(file multipart.File, handler *multipart.FileHeader, uri string) error {
	return m.updateErr
}

func (m *mockPhotoService) DeletePhoto(id int) error {
	return m.deleteErr
}

func (m *mockPhotoService) DeletePhotoFile(uri string) error {
	return m.deleteErr
}

//
// ---------------- TESTS ----------------
//

func TestCreatePhotoHandler_Success(t *testing.T) {
	mockSvc := &mockPhotoService{}
	controller := NewPhotoController(mockSvc)

	payload := map[string]interface{}{
		"uri":            "/photos/pic.jpg",
		"filename":       "pic.jpg",
		"displayed_name": "pic.jpg",
		"extension":      ".jpg",
		"caregiver_id":   1,
		"carepartner_id": 1,
	}

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/photo", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	controller.CreatePhotoHandler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestGetPhotoByIDHandler_Success(t *testing.T) {
	mockSvc := &mockPhotoService{
		photoResult: &media.PhotoMedia{ID: 1, Filename: "image.jpg"},
	}
	controller := NewPhotoController(mockSvc)

	req := httptest.NewRequest(http.MethodGet, "/photo?id=1", nil)
	w := httptest.NewRecorder()

	controller.GetPhotoByIDHandler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestGetPhotoByIDHandler_InvalidID(t *testing.T) {
	controller := NewPhotoController(&mockPhotoService{})
	req := httptest.NewRequest(http.MethodGet, "/photo?id=abc", nil)
	w := httptest.NewRecorder()

	controller.GetPhotoByIDHandler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestDeletePhotoHandler_Success(t *testing.T) {
	controller := NewPhotoController(&mockPhotoService{})
	req := httptest.NewRequest(http.MethodDelete, "/photo?id=1", nil)
	w := httptest.NewRecorder()

	controller.DeletePhotoHandler(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", w.Code)
	}
}

func TestUpdatePhotoHandler_Success(t *testing.T) {
	controller := NewPhotoController(&mockPhotoService{})
	photo := media.PhotoMedia{ID: 2, Filename: "update.jpg"}
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
		photoList: []media.PhotoMedia{
			{ID: 1, Filename: "p1.jpg"},
			{ID: 2, Filename: "p2.jpg"},
		},
	}
	controller := NewPhotoController(mockSvc)

	req := httptest.NewRequest(http.MethodGet, "/photo/list?carepartner_id=1&caregiver_id=1", nil)
	w := httptest.NewRecorder()

	controller.GetPhotosByCarepartnerAndCaregiverHandler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}
