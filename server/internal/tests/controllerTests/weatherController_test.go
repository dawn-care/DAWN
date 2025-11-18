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

type mockWeatherWidgetService struct {
	CreateWeatherWidgetFunc                        func(weather_widget *models.WeatherWidget) error
	CreateWeatherWidgetsBulkFunc                   func(weather_widgets []models.WeatherWidget) error
	GetWeatherWidgetByIDFunc                       func(id int) (*models.WeatherWidget, error)
	GetWeatherWidgetsByCarepartnerAndCaregiverFunc func(carepartner_id int, caregiver_id int) ([]models.WeatherWidget, error)
	GetWeatherWidgetsByCarepartnerIDAndDayFunc     func(carepartner_id int, day, month, year int) ([]models.WeatherWidget, error)
	UpdateWeatherWidgetFunc                        func(weather_widget *models.WeatherWidget) error
	DeleteWeatherWidgetFunc                        func(id int) error
	GetWeatherOverviewFunc                         func(weather_widget *models.WeatherWidget) (string, error)
	GetWeatherWidgetStatsFunc                      func() (int, error)
}

func (m *mockWeatherWidgetService) CreateWeatherWidget(weather_widget *models.WeatherWidget) error {
	return m.CreateWeatherWidgetFunc(weather_widget)
}

func (m *mockWeatherWidgetService) CreateWeatherWidgetsBulk(weather_widgets []models.WeatherWidget) error {
	return m.CreateWeatherWidgetsBulkFunc(weather_widgets)
}

func (m *mockWeatherWidgetService) GetWeatherWidgetByID(id int) (*models.WeatherWidget, error) {
	return m.GetWeatherWidgetByIDFunc(id)
}

func (m *mockWeatherWidgetService) GetWeatherWidgetsByCarepartnerAndCaregiver(carepartner_id int, caregiver_id int) ([]models.WeatherWidget, error) {
	return m.GetWeatherWidgetsByCarepartnerAndCaregiverFunc(carepartner_id, caregiver_id)
}

func (m *mockWeatherWidgetService) GetWeatherWidgetsByCarepartnerIDAndDay(carepartner_id int, day, month, year int) ([]models.WeatherWidget, error) {
	return m.GetWeatherWidgetsByCarepartnerIDAndDayFunc(carepartner_id, day, month, year)
}

func (m *mockWeatherWidgetService) UpdateWeatherWidget(weather_widget *models.WeatherWidget) error {
	return m.UpdateWeatherWidgetFunc(weather_widget)
}

func (m *mockWeatherWidgetService) DeleteWeatherWidget(id int) error {
	return m.DeleteWeatherWidgetFunc(id)
}

func (m *mockWeatherWidgetService) GetWeatherOverview(weather_widget *models.WeatherWidget) (string, error) {
	return m.GetWeatherOverviewFunc(weather_widget)
}

func (m *mockWeatherWidgetService) GetWeatherWidgetStats() (int, error) {
	return m.GetWeatherWidgetStatsFunc()
}

func TestCreateWeatherWidgetHandler(t *testing.T) {
	mockService := &mockWeatherWidgetService{
		CreateWeatherWidgetFunc: func(weather_widget *models.WeatherWidget) error {
			return nil
		},
	}
	controller := controllers.NewWeatherWidgetController(mockService)

	weather_widget := models.WeatherWidget{
		ZIPCODE: "12345",
	}
	body, _ := json.Marshal(weather_widget)
	req := httptest.NewRequest("POST", "/weather", bytes.NewBuffer(body))
	rec := httptest.NewRecorder()

	controller.CreateWeatherWidgetHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", rec.Code)
	}
}

func TestGetWeatherWidgetByIDHandler(t *testing.T) {
	mockService := &mockWeatherWidgetService{
		GetWeatherWidgetByIDFunc: func(id int) (*models.WeatherWidget, error) {
			return &models.WeatherWidget{ID: id, ZIPCODE: "12345"}, nil
		},
	}

	controller := controllers.NewWeatherWidgetController(mockService)

	req := httptest.NewRequest("GET", "/weather?id=1", nil)
	rec := httptest.NewRecorder()

	controller.GetWeatherWidgetByIDHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", rec.Code)
	}
}

func TestGetWeatherWidgetsByCarepartnerAndCaregiverHandler(t *testing.T) {
	mockService := &mockWeatherWidgetService{
		GetWeatherWidgetsByCarepartnerAndCaregiverFunc: func(carepartner_id int, caregiver_id int) ([]models.WeatherWidget, error) {
			return []models.WeatherWidget{{ID: 1, ZIPCODE: "12345"}}, nil
		},
	}

	controller := controllers.NewWeatherWidgetController(mockService)

	req := httptest.NewRequest("GET", "/weather?carepartner_id=1&caregiver_id=1", nil)
	rec := httptest.NewRecorder()

	controller.GetWeatherWidgetsByCarepartnerAndCaregiverHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", rec.Code)
	}
}

func TestGetWeatherWidgetsByCarepartnerIDAndDayHandler(t *testing.T) {
	mockService := &mockWeatherWidgetService{
		GetWeatherWidgetsByCarepartnerIDAndDayFunc: func(carepartner_id int, day, month, year int) ([]models.WeatherWidget, error) {
			return []models.WeatherWidget{{ID: 1, ZIPCODE: "12345"}}, nil
		},
	}

	controller := controllers.NewWeatherWidgetController(mockService)

	req := httptest.NewRequest("GET", "/weather?carepartner_id=1&day=1&month=1&year=2024", nil)
	rec := httptest.NewRecorder()

	controller.GetWeatherByDayAndCarepartnerHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", rec.Code)
	}
}

func TestUpdateWeatherWidgetHandler(t *testing.T) {
	mockService := &mockWeatherWidgetService{
		UpdateWeatherWidgetFunc: func(weather_widget *models.WeatherWidget) error {
			return nil
		},
	}

	controller := controllers.NewWeatherWidgetController(mockService)

	weather_widget := models.WeatherWidget{
		ID:      1,
		ZIPCODE: "12345",
	}
	body, _ := json.Marshal(weather_widget)
	req := httptest.NewRequest("PUT", "/weather", bytes.NewBuffer(body))
	rec := httptest.NewRecorder()

	controller.UpdateWeatherWidgetHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", rec.Code)
	}
}

func TestDeleteWeatherWidgetHandler(t *testing.T) {
	mockService := &mockWeatherWidgetService{
		DeleteWeatherWidgetFunc: func(id int) error {
			return nil
		},
	}

	controller := controllers.NewWeatherWidgetController(mockService)

	req := httptest.NewRequest("DELETE", "/weather?id=1", nil)
	rec := httptest.NewRecorder()

	controller.DeleteWeatherWidgetHandler(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("Expected status code 204, got %d", rec.Code)
	}
}

func TestWeatherOverviewHandler(t *testing.T) {
	mockService := &mockWeatherWidgetService{
		GetWeatherOverviewFunc: func(weather_widget *models.WeatherWidget) (string, error) {
			return "", nil
		},
	}

	controller := controllers.NewWeatherWidgetController(mockService)

	weather_widget := models.WeatherWidget{
		ID:      1,
		ZIPCODE: "12345",
	}
	body, _ := json.Marshal(weather_widget)
	req := httptest.NewRequest("POST", "/weather/overview", bytes.NewBuffer(body))
	rec := httptest.NewRecorder()

	controller.GetWeatherOverviewHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", rec.Code)
	}
}

func TestGetWeatherWidgetStatsHandler(t *testing.T) {
	mockService := &mockWeatherWidgetService{
		GetWeatherWidgetStatsFunc: func() (int, error) {
			return 5, nil
		},
	}

	controller := controllers.NewWeatherWidgetController(mockService)

	req := httptest.NewRequest("GET", "/weather/stats", nil)
	rec := httptest.NewRecorder()

	controller.GetWeatherWidgetStatsHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", rec.Code)
	}
}
