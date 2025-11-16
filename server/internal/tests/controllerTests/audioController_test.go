package media_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	controllers "server/internal/controllers/modules/media"
	models "server/internal/models/modules/media"
)

// Mock service implementing AudioServiceInterface
type mockAudioService struct {
	CreateAudioFunc                        func(audio *models.AudioMedia) error
	CreateAudioFileFunc                    func(file multipart.File, handler *multipart.FileHeader, uri string) error
	GetAudioByIDFunc                       func(id int) (*models.AudioMedia, error)
	GetAudioFileFunc                       func(uri string) (multipart.File, error)
	GetAudiosByCarepartnerAndCaregiverFunc func(carepartnerID, caregiverID int) ([]models.AudioMedia, error)
	UpdateAudioFunc                        func(audio *models.AudioMedia) error
	UpdateAudioFileFunc                    func(file multipart.File, handler *multipart.FileHeader, uri string) error
	DeleteAudioFunc                        func(id int) error
	DeleteAudioFileFunc                    func(uri string) error
	SearchAudiosFunc                       func(keyword string) ([]models.AudioMedia, error)
	GetMediaStatsFunc                      func() (*models.MediaStats, error)
}

func (m *mockAudioService) CreateAudio(a *models.AudioMedia) error {
	return m.CreateAudioFunc(a)
}
func (m *mockAudioService) CreateAudioFile(f multipart.File, h *multipart.FileHeader, u string) error {
	return m.CreateAudioFileFunc(f, h, u)
}
func (m *mockAudioService) GetAudioByID(id int) (*models.AudioMedia, error) {
	return m.GetAudioByIDFunc(id)
}
func (m *mockAudioService) GetAudioFile(uri string) (multipart.File, error) {
	return m.GetAudioFileFunc(uri)
}
func (m *mockAudioService) GetAudiosByCarepartnerAndCaregiver(carepartnerID, caregiverID int) ([]models.AudioMedia, error) {
	return m.GetAudiosByCarepartnerAndCaregiverFunc(carepartnerID, caregiverID)
}
func (m *mockAudioService) UpdateAudio(a *models.AudioMedia) error {
	return m.UpdateAudioFunc(a)
}
func (m *mockAudioService) UpdateAudioFile(f multipart.File, h *multipart.FileHeader, u string) error {
	return m.UpdateAudioFileFunc(f, h, u)
}
func (m *mockAudioService) DeleteAudio(id int) error {
	return m.DeleteAudioFunc(id)
}
func (m *mockAudioService) DeleteAudioFile(uri string) error {
	return m.DeleteAudioFileFunc(uri)
}
func (m *mockAudioService) SearchAudios(keyword string) ([]models.AudioMedia, error) {
	return m.SearchAudiosFunc(keyword)
}
func (m *mockAudioService) GetMediaStats() (*models.MediaStats, error) {
	return m.GetMediaStatsFunc()
}

// ---- Tests ----

func TestCreateAudioHandler_Success(t *testing.T) {
	mockSvc := &mockAudioService{
		CreateAudioFunc: func(a *models.AudioMedia) error { return nil },
	}

	ctrl := controllers.NewAudioController(mockSvc)

	body, _ := json.Marshal(models.AudioMedia{URI: "test", Filename: "a.mp3", Extension: "mp3"})
	req := httptest.NewRequest(http.MethodPost, "/audio", bytes.NewReader(body))
	w := httptest.NewRecorder()

	ctrl.CreateAudioHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestCreateAudioHandler_InvalidBody(t *testing.T) {
	ctrl := controllers.NewAudioController(&mockAudioService{})
	req := httptest.NewRequest(http.MethodPost, "/audio", bytes.NewReader([]byte("invalid")))
	w := httptest.NewRecorder()

	ctrl.CreateAudioHandler(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestGetAudioByIDHandler_Success(t *testing.T) {
	mockSvc := &mockAudioService{
		GetAudioByIDFunc: func(id int) (*models.AudioMedia, error) {
			return &models.AudioMedia{ID: 1, Filename: "a.mp3"}, nil
		},
	}

	ctrl := controllers.NewAudioController(mockSvc)
	req := httptest.NewRequest(http.MethodGet, "/audio?id=1", nil)
	w := httptest.NewRecorder()

	ctrl.GetAudioByIDHandler(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestGetAudioByIDHandler_Error(t *testing.T) {
	mockSvc := &mockAudioService{
		GetAudioByIDFunc: func(id int) (*models.AudioMedia, error) { return nil, errors.New("not found") },
	}
	ctrl := controllers.NewAudioController(mockSvc)

	req := httptest.NewRequest(http.MethodGet, "/audio?id=99", nil)
	w := httptest.NewRecorder()
	ctrl.GetAudioByIDHandler(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}
