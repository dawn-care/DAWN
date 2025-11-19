package widgets

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	models "server/internal/models/widgets/news"
	service "server/internal/services/widgets/news"
)

type mockNewsWidgetService struct {
	CreateNewsWidgetFunc                        func(news_widget *models.NewsWidget) error
	CreateNewsWidgetsBulkFunc                   func(news_widgets []models.NewsWidget) error
	GetNewsWidgetByIDFunc                       func(id int) (*models.NewsWidget, error)
	GetNewsWidgetsByCarepartnerAndCaregiverFunc func(carepartner_id int, caregiver_id int) ([]models.NewsWidget, error)
	GetNewsWidgetsByCarepartnerIDAndDayFunc     func(carepartner_id int, day, month, year int) ([]models.NewsWidget, error)
	UpdateNewsWidgetFunc                        func(news_widget *models.NewsWidget) error
	DeleteNewsWidgetFunc                        func(id int) error
	GetNewsWidgetStatsFunc                      func() (int, error)
	GetGoodNewsFunc                             func() (*service.NewsAPIResponse, error)
	GetJokeFunc                                 func() (string, error)
}

func (m *mockNewsWidgetService) CreateNewsWidget(news_widget *models.NewsWidget) error {
	return m.CreateNewsWidgetFunc(news_widget)
}

func (m *mockNewsWidgetService) CreateNewsWidgetsBulk(news_widgets []models.NewsWidget) error {
	return m.CreateNewsWidgetsBulkFunc(news_widgets)
}

func (m *mockNewsWidgetService) GetNewsWidgetByID(id int) (*models.NewsWidget, error) {
	return m.GetNewsWidgetByIDFunc(id)
}

func (m *mockNewsWidgetService) GetNewsWidgetsByCarepartnerAndCaregiver(carepartner_id int, caregiver_id int) ([]models.NewsWidget, error) {
	return m.GetNewsWidgetsByCarepartnerAndCaregiverFunc(carepartner_id, caregiver_id)
}

func (m *mockNewsWidgetService) GetNewsWidgetsByCarepartnerIDAndDay(carepartner_id int, day, month, year int) ([]models.NewsWidget, error) {
	return m.GetNewsWidgetsByCarepartnerIDAndDayFunc(carepartner_id, day, month, year)
}

func (m *mockNewsWidgetService) UpdateNewsWidget(news_widget *models.NewsWidget) error {
	return m.UpdateNewsWidgetFunc(news_widget)
}

func (m *mockNewsWidgetService) DeleteNewsWidget(id int) error {
	return m.DeleteNewsWidgetFunc(id)
}

func (m *mockNewsWidgetService) GetNewsWidgetStats() (int, error) {
	return m.GetNewsWidgetStatsFunc()
}

func (m *mockNewsWidgetService) GetGoodNews() (*service.NewsAPIResponse, error) {
	return m.GetGoodNewsFunc()
}

func (m *mockNewsWidgetService) GetJoke() (string, error) {
	return m.GetJokeFunc()
}

func TestCreateNewsWidgetHandler(t *testing.T) {
	mockService := &mockNewsWidgetService{
		CreateNewsWidgetFunc: func(news_widget *models.NewsWidget) error {
			return nil
		},
	}
	controller := NewNewsWidgetController(mockService)

	news_widget := models.NewsWidget{
		CATEGORY: "news",
		LANGUAGE: "en",
		COUNTRY:  "us",
	}
	body, _ := json.Marshal(news_widget)
	req := httptest.NewRequest("POST", "/news", bytes.NewBuffer(body))
	rec := httptest.NewRecorder()

	controller.CreateNewsWidgetHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", rec.Code)
	}
}

func TestGetNewsWidgetByIDHandler(t *testing.T) {
	mockService := &mockNewsWidgetService{
		GetNewsWidgetByIDFunc: func(id int) (*models.NewsWidget, error) {
			return &models.NewsWidget{
				ID:       id,
				CATEGORY: "news",
				LANGUAGE: "en",
				COUNTRY:  "us",
			}, nil
		},
	}

	controller := NewNewsWidgetController(mockService)

	req := httptest.NewRequest("GET", "/news?id=1", nil)
	rec := httptest.NewRecorder()

	controller.GetNewsWidgetByIDHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", rec.Code)
	}
}

func TestGetNewsWidgetsByCarepartnerAndCaregiverHandler(t *testing.T) {
	mockService := &mockNewsWidgetService{
		GetNewsWidgetsByCarepartnerAndCaregiverFunc: func(carepartner_id int, caregiver_id int) ([]models.NewsWidget, error) {
			return []models.NewsWidget{{
				ID:       1,
				CATEGORY: "news",
				LANGUAGE: "en",
				COUNTRY:  "us",
			}}, nil
		},
	}

	controller := NewNewsWidgetController(mockService)

	req := httptest.NewRequest("GET", "/news?carepartner_id=1&caregiver_id=1", nil)
	rec := httptest.NewRecorder()

	controller.GetNewsWidgetsByCarepartnerAndCaregiverHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", rec.Code)
	}
}

func TestGetNewsWidgetsByCarepartnerIDAndDayHandler(t *testing.T) {
	mockService := &mockNewsWidgetService{
		GetNewsWidgetsByCarepartnerIDAndDayFunc: func(carepartner_id int, day, month, year int) ([]models.NewsWidget, error) {
			return []models.NewsWidget{{
				ID:       1,
				CATEGORY: "news",
				LANGUAGE: "en",
				COUNTRY:  "us",
			}}, nil
		},
	}

	controller := NewNewsWidgetController(mockService)

	req := httptest.NewRequest("GET", "/news?carepartner_id=1&day=1&month=1&year=2024", nil)
	rec := httptest.NewRecorder()

	controller.GetNewsByDayAndCarepartnerHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", rec.Code)
	}
}

func TestUpdateNewsWidgetHandler(t *testing.T) {
	mockService := &mockNewsWidgetService{
		UpdateNewsWidgetFunc: func(news_widget *models.NewsWidget) error {
			return nil
		},
	}

	controller := NewNewsWidgetController(mockService)

	news_widget := models.NewsWidget{
		ID:       1,
		CATEGORY: "news",
		LANGUAGE: "en",
		COUNTRY:  "us",
	}
	body, _ := json.Marshal(news_widget)
	req := httptest.NewRequest("PUT", "/news", bytes.NewBuffer(body))
	rec := httptest.NewRecorder()

	controller.UpdateNewsWidgetHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", rec.Code)
	}
}

func TestDeleteNewsWidgetHandler(t *testing.T) {
	mockService := &mockNewsWidgetService{
		DeleteNewsWidgetFunc: func(id int) error {
			return nil
		},
	}

	controller := NewNewsWidgetController(mockService)

	req := httptest.NewRequest("DELETE", "/news?id=1", nil)
	rec := httptest.NewRecorder()

	controller.DeleteNewsWidgetHandler(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("Expected status code 204, got %d", rec.Code)
	}
}

func TestGetNewsWidgetStatsHandler(t *testing.T) {
	mockService := &mockNewsWidgetService{
		GetNewsWidgetStatsFunc: func() (int, error) {
			return 5, nil
		},
	}

	controller := NewNewsWidgetController(mockService)

	req := httptest.NewRequest("GET", "/news/stats", nil)
	rec := httptest.NewRecorder()

	controller.GetNewsWidgetStatsHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", rec.Code)
	}
}

func TestGetGoodNewsHandler(t *testing.T) {
	mockService := &mockNewsWidgetService{
		GetGoodNewsFunc: func() (*service.NewsAPIResponse, error) {
			return &service.NewsAPIResponse{}, nil
		},
	}

	controller := NewNewsWidgetController(mockService)

	req := httptest.NewRequest("GET", "/news/goodnews", nil)
	rec := httptest.NewRecorder()

	controller.GetGoodNewsHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", rec.Code)
	}
}

func TestGetJokeHandler(t *testing.T) {
	mockService := &mockNewsWidgetService{
		GetJokeFunc: func() (string, error) {
			return "yo mama", nil
		},
	}

	controller := NewNewsWidgetController(mockService)

	req := httptest.NewRequest("GET", "/news/goodnews", nil)
	rec := httptest.NewRecorder()

	controller.GetJokeHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", rec.Code)
	}
}
