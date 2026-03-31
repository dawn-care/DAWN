package widgets

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	models "server/internal/models/widgets/news"
	services "server/internal/services/widgets/news"

	"github.com/go-chi/chi/v5"
)

type mockSportsNewsWidgetService struct {
	CreateSportsNewsWidgetFunc                        func(news_widget *models.NewsWidget) error
	CreateSportsNewsWidgetsBulkFunc                   func(news_widgets []models.NewsWidget) error
	GetSportsNewsWidgetByIDFunc                       func(id int) (*models.NewsWidget, error)
	GetSportsNewsWidgetsByCarepartnerAndCaregiverFunc func(carepartner_id int, caregiver_id int) ([]models.NewsWidget, error)
	GetSportsNewsWidgetsByCarepartnerIDAndDayFunc     func(carepartner_id int, day, month, year int) ([]models.NewsWidget, error)
	UpdateSportsNewsWidgetFunc                        func(news_widget *models.NewsWidget) error
	DeleteSportsNewsWidgetFunc                        func(id int) error
	GetSportsNewsWidgetStatsFunc                      func() (int, error)
	GetTeamGameFunc                                   func(news_widget *models.NewsWidget) (*services.GameSummaryResponse, error)
}

func (m *mockSportsNewsWidgetService) CreateSportsNewsWidget(news_widget *models.NewsWidget) error {
	return m.CreateSportsNewsWidgetFunc(news_widget)
}
func (m *mockSportsNewsWidgetService) CreateSportsNewsWidgetsBulk(news_widgets []models.NewsWidget) error {
	return m.CreateSportsNewsWidgetsBulkFunc(news_widgets)
}
func (m *mockSportsNewsWidgetService) GetSportsNewsWidgetByID(id int) (*models.NewsWidget, error) {
	return m.GetSportsNewsWidgetByIDFunc(id)
}
func (m *mockSportsNewsWidgetService) GetSportsNewsWidgetsByCarepartnerAndCaregiver(carepartner_id int, caregiver_id int) ([]models.NewsWidget, error) {
	return m.GetSportsNewsWidgetsByCarepartnerAndCaregiverFunc(carepartner_id, caregiver_id)
}
func (m *mockSportsNewsWidgetService) GetSportsNewsWidgetsByCarepartnerIDAndDay(carepartner_id int, day, month, year int) ([]models.NewsWidget, error) {
	return m.GetSportsNewsWidgetsByCarepartnerIDAndDayFunc(carepartner_id, day, month, year)
}
func (m *mockSportsNewsWidgetService) UpdateSportsNewsWidget(news_widget *models.NewsWidget) error {
	return m.UpdateSportsNewsWidgetFunc(news_widget)
}
func (m *mockSportsNewsWidgetService) DeleteSportsNewsWidget(id int) error {
	return m.DeleteSportsNewsWidgetFunc(id)
}
func (m *mockSportsNewsWidgetService) GetSportsNewsWidgetStats() (int, error) {
	return m.GetSportsNewsWidgetStatsFunc()
}
func (m *mockSportsNewsWidgetService) GetTeamGame(news_widget *models.NewsWidget) (*services.GameSummaryResponse, error) {
	return m.GetTeamGameFunc(news_widget)
}

func TestCreateSportsNewsWidgetHandler(t *testing.T) {
	mockService := &mockSportsNewsWidgetService{
		CreateSportsNewsWidgetFunc: func(news_widget *models.NewsWidget) error {
			return nil
		},
	}

	controller := NewSportsNewsWidgetController(mockService)

	news_widget := models.NewsWidget{
		CATEGORY: "sports",
		LANGUAGE: "buf",
		COUNTRY:  "nfl",
	}

	body, _ := json.Marshal(news_widget)
	req := httptest.NewRequest("POST", "/sports-news-widgets", bytes.NewBuffer(body))
	rec := httptest.NewRecorder()

	controller.CreateSportsNewsWidgetHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", rec.Code)
	}
}

func TestGetSportsNewsWidgetByIDHandler(t *testing.T) {
	mockService := &mockSportsNewsWidgetService{
		GetSportsNewsWidgetByIDFunc: func(id int) (*models.NewsWidget, error) {
			return &models.NewsWidget{ID: id}, nil
		},
	}

	controller := NewSportsNewsWidgetController(mockService)

	r := chi.NewRouter()
	r.Get("/sports-news-widgets/{id}", controller.GetSportsNewsWidgetByIDHandler)

	req := httptest.NewRequest("GET", "/sports-news-widgets/1", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", rec.Code)
	}
}

func TestGetSportsNewsWidgetsByCarepartnerAndCaregiverHandler(t *testing.T) {
	mockService := &mockSportsNewsWidgetService{
		GetSportsNewsWidgetsByCarepartnerAndCaregiverFunc: func(carepartner_id int, caregiver_id int) ([]models.NewsWidget, error) {
			return []models.NewsWidget{{ID: 1}}, nil
		},
	}

	controller := NewSportsNewsWidgetController(mockService)

	req := httptest.NewRequest("GET", "/sports-news-widgets/by-carepartner-and-caregiver?carepartner_id=1&caregiver_id=1", nil)
	rec := httptest.NewRecorder()

	controller.GetSportsNewsWidgetsByCarepartnerAndCaregiverHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", rec.Code)
	}
}

func TestGetSportsNewsByDayAndCarepartnerHandler(t *testing.T) {
	mockService := &mockSportsNewsWidgetService{
		GetSportsNewsWidgetsByCarepartnerIDAndDayFunc: func(carepartner_id int, day, month, year int) ([]models.NewsWidget, error) {
			return []models.NewsWidget{{ID: 1}}, nil
		},
	}

	controller := NewSportsNewsWidgetController(mockService)

	req := httptest.NewRequest("GET", "/sports-news-widgets/by-day?carepartner_id=1&day=1&month=1&year=2024", nil)
	rec := httptest.NewRecorder()

	controller.GetSportsNewsByDayAndCarepartnerHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", rec.Code)
	}
}

func TestUpdateSportsNewsWidgetHandler(t *testing.T) {
	mockService := &mockSportsNewsWidgetService{
		UpdateSportsNewsWidgetFunc: func(news_widget *models.NewsWidget) error {
			return nil
		},
	}

	controller := NewSportsNewsWidgetController(mockService)

	news_widget := models.NewsWidget{ID: 1}
	body, _ := json.Marshal(news_widget)

	req := httptest.NewRequest("PUT", "/sports-news-widgets", bytes.NewBuffer(body))
	rec := httptest.NewRecorder()

	controller.UpdateSportsNewsWidgetHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", rec.Code)
	}
}

func TestDeleteSportsNewsWidgetHandler(t *testing.T) {
	mockService := &mockSportsNewsWidgetService{
		DeleteSportsNewsWidgetFunc: func(id int) error {
			return nil
		},
	}

	controller := NewSportsNewsWidgetController(mockService)

	req := httptest.NewRequest("DELETE", "/sports-news-widgets?id=1", nil)
	rec := httptest.NewRecorder()

	controller.DeleteSportsNewsWidgetHandler(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("Expected status code 204, got %d", rec.Code)
	}
}

func TestGetSportsNewsWidgetStatsHandler(t *testing.T) {
	mockService := &mockSportsNewsWidgetService{
		GetSportsNewsWidgetStatsFunc: func() (int, error) {
			return 5, nil
		},
	}

	controller := NewSportsNewsWidgetController(mockService)

	req := httptest.NewRequest("GET", "/sports-news-widgets/stats", nil)
	rec := httptest.NewRecorder()

	controller.GetSportsNewsWidgetStatsHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", rec.Code)
	}
}

func TestGetTeamGameHandler(t *testing.T) {
	mockService := &mockSportsNewsWidgetService{
		GetTeamGameFunc: func(news_widget *models.NewsWidget) (*services.GameSummaryResponse, error) {
			return &services.GameSummaryResponse{
				UserTeamName:     "Bills",
				OpponentTeamName: "Chiefs",
				UserScore:        24,
				OpponentScore:    17,
				Message:          "you win!",
			}, nil
		},
	}

	controller := NewSportsNewsWidgetController(mockService)

	body := models.NewsWidget{
		CATEGORY: "sports",
		LANGUAGE: "buf",
		COUNTRY:  "nfl",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("GET", "/sports-news-widgets/team-game", bytes.NewBuffer(jsonBody))
	rec := httptest.NewRecorder()

	controller.GetTeamGameHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", rec.Code)
	}
}
