package media

import (
	"bytes"
	"encoding/json"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	controllers "server/internal/controllers/media"
	models "server/internal/models/media"
)

// Mock service implementing VideoServiceInterface
type mockVideoService struct {
	CreateVideoFn                        func(*models.VideoMedia) error
	CreateVideoFileFn                    func(multipart.File, *multipart.FileHeader, string) error
	GetVideoByIDFn                       func(int) (*models.VideoMedia, error)
	GetVideoFileFn                       func(string) (multipart.File, error)
	GetVideosByCarepartnerAndCaregiverFn func(int, int) ([]models.VideoMedia, error)
	UpdateVideoFn                        func(*models.VideoMedia) error
	UpdateVideoFileFn                    func(multipart.File, *multipart.FileHeader, string) error
	DeleteVideoFn                        func(int) error
	DeleteVideoFileFn                    func(string) error
	GetVideoStatsFn                      func() (int, error)
}

func (m *mockVideoService) CreateVideo(v *models.VideoMedia) error {
	return m.CreateVideoFn(v)
}

func (m *mockVideoService) CreateVideoFile(f multipart.File, h *multipart.FileHeader, u string) error {
	return m.CreateVideoFileFn(f, h, u)
}

func (m *mockVideoService) GetVideoByID(id int) (*models.VideoMedia, error) {
	return m.GetVideoByIDFn(id)
}

func (m *mockVideoService) GetVideoFile(u string) (multipart.File, error) {
	return m.GetVideoFileFn(u)
}

func (m *mockVideoService) GetVideosByCarepartnerAndCaregiver(carepartnerID, caregiverID int) ([]models.VideoMedia, error) {
	return m.GetVideosByCarepartnerAndCaregiverFn(carepartnerID, caregiverID)
}

func (m *mockVideoService) UpdateVideo(v *models.VideoMedia) error {
	return m.UpdateVideoFn(v)
}

func (m *mockVideoService) UpdateVideoFile(f multipart.File, h *multipart.FileHeader, u string) error {
	return m.UpdateVideoFileFn(f, h, u)
}

func (m *mockVideoService) DeleteVideo(id int) error {
	return m.DeleteVideoFn(id)
}

func (m *mockVideoService) DeleteVideoFile(u string) error {
	return m.DeleteVideoFileFn(u)
}

func (m *mockVideoService) GetVideoStats() (int, error) {
	return m.GetVideoStatsFn()
}

// ---- Tests ----

func TestCreateVideoHandler(t *testing.T) {
	mockSvc := &mockVideoService{
		CreateVideoFn: func(v *models.VideoMedia) error { return nil },
	}

	controller := controllers.NewVideoController(mockSvc)

	video := models.VideoMedia{URI: "test.com/video.mp4", Filename: "video.mp4", Extension: "mp4"}
	body, _ := json.Marshal(video)
	req := httptest.NewRequest(http.MethodPost, "/videos", bytes.NewBuffer(body))
	rec := httptest.NewRecorder()

	controller.CreateVideoHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestGetVideoByIDHandler(t *testing.T) {
	mockSvc := &mockVideoService{
		GetVideoByIDFn: func(id int) (*models.VideoMedia, error) {
			return &models.VideoMedia{ID: id, Filename: "sample.mp4"}, nil
		},
	}

	controller := controllers.NewVideoController(mockSvc)
	req := httptest.NewRequest(http.MethodGet, "/videos?id=1", nil)
	rec := httptest.NewRecorder()

	controller.GetVideoByIDHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestGetVideosByCarepartnerAndCaregiverHandler(t *testing.T) {
	mockSvc := &mockVideoService{
		GetVideosByCarepartnerAndCaregiverFn: func(carepartnerID, caregiverID int) ([]models.VideoMedia, error) {
			return []models.VideoMedia{{ID: 1, Filename: "vid1.mp4"}}, nil
		},
	}

	controller := controllers.NewVideoController(mockSvc)
	req := httptest.NewRequest(http.MethodGet, "/videos?carepartner_id=1&caregiver_id=2", nil)
	rec := httptest.NewRecorder()

	controller.GetVideosByCarepartnerAndCaregiverHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestUpdateVideoHandler(t *testing.T) {
	mockSvc := &mockVideoService{
		UpdateVideoFn: func(v *models.VideoMedia) error { return nil },
	}

	controller := controllers.NewVideoController(mockSvc)

	video := models.VideoMedia{ID: 1, URI: "updated.mp4"}
	body, _ := json.Marshal(video)
	req := httptest.NewRequest(http.MethodPut, "/videos", bytes.NewBuffer(body))
	rec := httptest.NewRecorder()

	controller.UpdateVideoHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestDeleteVideoHandler(t *testing.T) {
	mockSvc := &mockVideoService{
		DeleteVideoFn: func(id int) error { return nil },
	}

	controller := controllers.NewVideoController(mockSvc)
	req := httptest.NewRequest(http.MethodDelete, "/videos?id=1", nil)
	rec := httptest.NewRecorder()

	controller.DeleteVideoHandler(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("expected status 204, got %d", rec.Code)
	}
}

func TestGetVideoStatsHandler(t *testing.T) {
	mockSvc := &mockVideoService{
		GetVideoStatsFn: func() (int, error) { return 42, nil },
	}

	controller := controllers.NewVideoController(mockSvc)
	req := httptest.NewRequest(http.MethodGet, "/videos/stats", nil)
	rec := httptest.NewRecorder()

	controller.GetVideoStatsHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestGetVideoByIDHandler_Error(t *testing.T) {
	mockSvc := &mockVideoService{
		GetVideoByIDFn: func(id int) (*models.VideoMedia, error) {
			return nil, errors.New("video not found")
		},
	}

	controller := controllers.NewVideoController(mockSvc)
	req := httptest.NewRequest(http.MethodGet, "/videos?id=1", nil)
	rec := httptest.NewRecorder()

	controller.GetVideoByIDHandler(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", rec.Code)
	}
}
