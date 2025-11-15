package widgets_tests

import (
	"database/sql"
	"log"
	"os"
	"testing"

	"server/internal/models/modules/widgets"

	_ "github.com/lib/pq"
	// "github.com/joho/godotenv"
)

var testDB *sql.DB

// Sets Up Postgres Connections
func TestMain(m *testing.M) {
	var err error
	connStr := "host=localhost port=5432 user=myuser password=mypassword dbname=mydb sslmode=disable"
	testDB, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	code := m.Run()
	testDB.Close()
	os.Exit(code)
}

// Helper Function to Clear Test Weather Table
func clearWeatherTable(t *testing.T) {
	_, err := testDB.Exec(`DELETE FROM "WeatherWidget";`)
	if err != nil {
		t.Fatalf("Failed to clear table: %v", err)
	}
}

// Tests CreateWeather and GetWeatherByID
func TestCreateWeatherWidget(t *testing.T) {
	clearWeatherTable(t)

	weather_widget := &widgets.WeatherWidget{
		ZIPCODE: "14623",
	}

	err := widgets.CreateWeather(testDB, weather_widget)
	if err != nil {
		t.Fatalf("CreateWeather Failed: %v", err)
	}
	if weather_widget.ID == 0 {
		t.Errorf("Weather Widget ID Not Set")
	}

	found_widget, err := widgets.GetWeatherByID(testDB, weather_widget.ID)
	if err != nil {
		t.Fatalf("GetWeatherByID Failed: %v", err)
	}
	if found_widget == nil {
		t.Fatalf("Widget Not Found")
	}
	if found_widget.ZIPCODE != "14623" {
		t.Errorf("Unexpected Widget Returned (CreateWeather): %+v", found_widget)
	}
}

// Tests UpdateWeather
func TestUpdateWeatherWidget(t *testing.T) {
	clearWeatherTable(t)

	weather_widget := &widgets.WeatherWidget{
		ZIPCODE: "14623",
	}

	err := widgets.CreateWeather(testDB, weather_widget)
	if err != nil {
		t.Fatalf("CreateWeather Failed: %v", err)
	}

	weather_widget.ZIPCODE = "14618"
	err = widgets.UpdateWeather(testDB, weather_widget)
	if err != nil {
		t.Fatalf("UpdateWeather Failed: %v", err)
	}
	found_widget, err := widgets.GetWeatherByID(testDB, weather_widget.ID)
	if found_widget.ZIPCODE != "14618" {
		t.Errorf("Unexpected Widget Returned (UpdateWeather): %+v", found_widget)
	}
}

// Tests DeleteWeather

func TestDeleteWeatherWidget(t *testing.T) {
	clearWeatherTable(t)

	weather_widget := &widgets.WeatherWidget{
		ZIPCODE: "14623",
	}

	err := widgets.CreateWeather(testDB, weather_widget)
	if err != nil {
		t.Fatalf("CreateWeather Failed: %v", err)
	}

	err = widgets.DeleteWeather(testDB, weather_widget.ID)
	if err != nil {
		t.Fatalf("DeleteWeather Failed: %v", err)
	}
	found_widget, err := widgets.GetWeatherByID(testDB, weather_widget.ID)
	if found_widget != nil {
		t.Errorf("Unexpected Widget Returned (DeleteWeather): %+v", found_widget)
	}
}

// Tests CreateWeatherBulk
func TestCreateWeatherBulkWidgets(t *testing.T) {
	clearWeatherTable(t)

	weather_widgets := []widgets.WeatherWidget{
		{ZIPCODE: "14623"},
		{ZIPCODE: "14618"},
		{ZIPCODE: "14626"},
	}
	err := widgets.CreateWeathersBulk(testDB, weather_widgets)
	if err != nil {
		t.Fatalf("CreateWeatherBulk Failed: %v", err)
	}
	found_widget, err := widgets.GetWeatherByID(testDB, weather_widgets[0].ID)
	if found_widget == nil {
		t.Error("CreateWeatherBulk widget[0] Not Found")
	}
	found_widget, err = widgets.GetWeatherByID(testDB, weather_widgets[1].ID)
	if found_widget == nil {
		t.Error("CreateWeatherBulk widget[1] Not Found")
	}
	found_widget, err = widgets.GetWeatherByID(testDB, weather_widgets[2].ID)
	if found_widget == nil {
		t.Error("CreateWeatherBulk widget[2] Not Found")
	}
}

// Test GetWeatherStats
func TestGetWeatherStatsWidget(t *testing.T) {
	clearWeatherTable(t)

	weather_widgets := []widgets.WeatherWidget{
		{ZIPCODE: "14623"},
		{ZIPCODE: "14618"},
		{ZIPCODE: "14626"},
	}
	err := widgets.CreateWeathersBulk(testDB, weather_widgets)
	if err != nil {
		t.Fatalf("CreateWeatherBulk Failed: %v", err)
	}

	count, err := widgets.GetWeatherStats(testDB)
	if count != 3 {
		t.Errorf("Unexpected Count Returned: %d", count)
	}
}
