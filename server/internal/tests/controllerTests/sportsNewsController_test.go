package controllertests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	controllers "server/internal/controllers/modules/widgets"
	models "server/internal/models/modules/widgets"
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
	GetTeamGameFunc                                   func(league, team string) (string, error)
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

func (m *mockSportsNewsWidgetService) GetTeamGame(league, team string) (string, error) {
	return m.GetTeamGameFunc(league, team)
}

func TestCreateSportsNewsWidgetHandler(t *testing.T) {
	mockService := &mockSportsNewsWidgetService{
		CreateSportsNewsWidgetFunc: func(news_widget *models.NewsWidget) error {
			return nil
		},
	}
	controller := controllers.NewSportsNewsWidgetController(mockService)

	news_widget := models.NewsWidget{
		CATEGORY: "sports",
		LANGUAGE: "buf",
		COUNTRY:  "nfl",
	}
	body, _ := json.Marshal(news_widget)
	req := httptest.NewRequest("POST", "/sportsnews", bytes.NewBuffer(body))
	rec := httptest.NewRecorder()

	controller.CreateSportsNewsWidgetHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", rec.Code)
	}
}

func TestGetSportsNewsWidgetByIDHandler(t *testing.T) {
	mockService := &mockSportsNewsWidgetService{
		GetSportsNewsWidgetByIDFunc: func(id int) (*models.NewsWidget, error) {
			return &models.NewsWidget{
				ID:       id,
				CATEGORY: "sports",
				LANGUAGE: "buf",
				COUNTRY:  "nfl",
			}, nil
		},
	}

	controller := controllers.NewSportsNewsWidgetController(mockService)

	req := httptest.NewRequest("GET", "/sportsnews?id=1", nil)
	rec := httptest.NewRecorder()

	controller.GetSportsNewsWidgetByIDHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", rec.Code)
	}
}

func TestGetSportsNewsWidgetsByCarepartnerAndCaregiverHandler(t *testing.T) {
	mockService := &mockSportsNewsWidgetService{
		GetSportsNewsWidgetsByCarepartnerAndCaregiverFunc: func(carepartner_id int, caregiver_id int) ([]models.NewsWidget, error) {
			return []models.NewsWidget{{
				ID:       1,
				CATEGORY: "sports",
				LANGUAGE: "buf",
				COUNTRY:  "nfl",
			}}, nil
		},
	}

	controller := controllers.NewSportsNewsWidgetController(mockService)

	req := httptest.NewRequest("GET", "/sportsnews?carepartner_id=1&caregiver_id=1", nil)
	rec := httptest.NewRecorder()

	controller.GetSportsNewsWidgetsByCarepartnerAndCaregiverHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", rec.Code)
	}
}

func TestGetSportsNewsWidgetsByCarepartnerIDAndDayHandler(t *testing.T) {
	mockService := &mockSportsNewsWidgetService{
		GetSportsNewsWidgetsByCarepartnerIDAndDayFunc: func(carepartner_id int, day, month, year int) ([]models.NewsWidget, error) {
			return []models.NewsWidget{{
				ID:       1,
				CATEGORY: "sports",
				LANGUAGE: "buf",
				COUNTRY:  "nfl",
			}}, nil
		},
	}

	controller := controllers.NewSportsNewsWidgetController(mockService)

	req := httptest.NewRequest("GET", "/sportsnews?carepartner_id=1&day=1&month=1&year=2024", nil)
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

	controller := controllers.NewSportsNewsWidgetController(mockService)

	news_widget := models.NewsWidget{
		ID:       1,
		CATEGORY: "sports",
		LANGUAGE: "buf",
		COUNTRY:  "nfl",
	}
	body, _ := json.Marshal(news_widget)
	req := httptest.NewRequest("PUT", "/sportsnews", bytes.NewBuffer(body))
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

	controller := controllers.NewSportsNewsWidgetController(mockService)

	req := httptest.NewRequest("DELETE", "/sportsnews?id=1", nil)
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

	controller := controllers.NewSportsNewsWidgetController(mockService)

	req := httptest.NewRequest("GET", "/sportsnews/stats", nil)
	rec := httptest.NewRecorder()

	controller.GetSportsNewsWidgetStatsHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", rec.Code)
	}
}

func TestGetTeamGameHandler(t *testing.T) {
	mockService := &mockSportsNewsWidgetService{
		GetTeamGameFunc: func(league, team string) (string, error) {
			return "you win!", nil
		},
	}

	controller := controllers.NewSportsNewsWidgetController(mockService)

	req := httptest.NewRequest("GET", "/sportsnews?league=nfl&team=buf", nil)
	rec := httptest.NewRecorder()

	controller.GetTeamGameHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", rec.Code)
	}
}
