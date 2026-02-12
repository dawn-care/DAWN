package media

import (
	"bytes"
	"encoding/json"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	models "server/internal/models/media"
)

// ---- Mock service implementing AudioServiceInterface ----

type mockAudioService struct {
	CreateAudioFunc                        func(*models.AudioMedia) error
	CreateAudioFileFunc                    func(multipart.File, *multipart.FileHeader, string) error
	GetAudioByIDFunc                       func(int) (*models.AudioMedia, error)
	GetAudioFileFunc                       func(string) (multipart.File, error)
	GetAudiosByCarepartnerAndCaregiverFunc func(int, int) ([]models.AudioMedia, error)
	UpdateAudioFunc                        func(*models.AudioMedia) error
	UpdateAudioFileFunc                    func(multipart.File, *multipart.FileHeader, string) error
	DeleteAudioFunc                        func(int) error
	DeleteAudioFileFunc                    func(string) error
}

func (m *mockAudioService) CreateAudio(a *models.AudioMedia) error {
	if m.CreateAudioFunc != nil {
		return m.CreateAudioFunc(a)
	}
	return nil
}

func (m *mockAudioService) CreateAudioFile(f multipart.File, h *multipart.FileHeader, u string) error {
	if m.CreateAudioFileFunc != nil {
		return m.CreateAudioFileFunc(f, h, u)
	}
	return nil
}

func (m *mockAudioService) GetAudioByID(id int) (*models.AudioMedia, error) {
	if m.GetAudioByIDFunc != nil {
		return m.GetAudioByIDFunc(id)
	}
	return nil, nil
}

func (m *mockAudioService) GetAudioFile(uri string) (multipart.File, error) {
	if m.GetAudioFileFunc != nil {
		return m.GetAudioFileFunc(uri)
	}
	return nil, nil
}

func (m *mockAudioService) GetAudiosByCarepartnerAndCaregiver(carepartnerID, caregiverID int) ([]models.AudioMedia, error) {
	if m.GetAudiosByCarepartnerAndCaregiverFunc != nil {
		return m.GetAudiosByCarepartnerAndCaregiverFunc(carepartnerID, caregiverID)
	}
	return nil, nil
}

func (m *mockAudioService) UpdateAudio(a *models.AudioMedia) error {
	if m.UpdateAudioFunc != nil {
		return m.UpdateAudioFunc(a)
	}
	return nil
}

func (m *mockAudioService) UpdateAudioFile(f multipart.File, h *multipart.FileHeader, u string) error {
	if m.UpdateAudioFileFunc != nil {
		return m.UpdateAudioFileFunc(f, h, u)
	}
	return nil
}

func (m *mockAudioService) DeleteAudio(id int) error {
	if m.DeleteAudioFunc != nil {
		return m.DeleteAudioFunc(id)
	}
	return nil
}

func (m *mockAudioService) DeleteAudioFile(uri string) error {
	if m.DeleteAudioFileFunc != nil {
		return m.DeleteAudioFileFunc(uri)
	}
	return nil
}

// ---- Tests ----

func TestCreateAudioHandler_Success(t *testing.T) {
	mockSvc := &mockAudioService{
		CreateAudioFunc: func(a *models.AudioMedia) error {
			return nil
		},
	}

	ctrl := NewAudioController(mockSvc)

	body, _ := json.Marshal(models.AudioMedia{
		URI:       "test/audio.mp3",
		Filename:  "audio.mp3",
		Extension: ".mp3",
	})

	req := httptest.NewRequest(http.MethodPost, "/audio", bytes.NewReader(body))
	w := httptest.NewRecorder()

	ctrl.CreateAudioHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestCreateAudioHandler_InvalidBody(t *testing.T) {
	ctrl := NewAudioController(&mockAudioService{
		CreateAudioFunc: func(a *models.AudioMedia) error { return nil },
	})

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

	ctrl := NewAudioController(mockSvc)
	req := httptest.NewRequest(http.MethodGet, "/audio?id=1", nil)
	w := httptest.NewRecorder()

	ctrl.GetAudioByIDHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestGetAudioByIDHandler_Error(t *testing.T) {
	mockSvc := &mockAudioService{
		GetAudioByIDFunc: func(id int) (*models.AudioMedia, error) {
			return nil, errors.New("not found")
		},
	}

	ctrl := NewAudioController(mockSvc)

	req := httptest.NewRequest(http.MethodGet, "/audio?id=99", nil)
	w := httptest.NewRecorder()

	ctrl.GetAudioByIDHandler(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}
