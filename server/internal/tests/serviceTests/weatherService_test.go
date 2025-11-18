package servicetests

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"

	modelsWidget "server/internal/models/modules/widgets"
	servicesWeatherWidget "server/internal/services/modules/widgets/weather"
	"server/internal/tests"
)

func TestCreateWeatherWidget(t *testing.T) {
	db, mock := tests.SetupMockDB(t)
	defer db.Close()

	weather_service := servicesWeatherWidget.NewWeatherService(db, "test_api_key")

	weather_widget := &modelsWidget.WeatherWidget{
		ZIPCODE: "11356",
	}

	mock.ExpectQuery(regexp.QuoteMeta(`
		INSERT INTO "WeatherWidget" (zipcode)
		VALUES ($1)
		RETURNING id;
	`)).
		WithArgs("11356").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	err := weather_service.CreateWeatherWidget(weather_widget)
	assert.NoError(t, err)
	assert.Equal(t, 1, weather_widget.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetWeatherWidgetByID(t *testing.T) {
	db, mock := tests.SetupMockDB(t)
	defer db.Close()

	weather_service := servicesWeatherWidget.NewWeatherService(db, "test_api_key")

	rows := sqlmock.NewRows([]string{"id", "zidcode"}).
		AddRow(1, "11356")

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, zipcode
		FROM "WeatherWidget"
		WHERE id = $1;
	`)).
		WithArgs(1).
		WillReturnRows(rows)

	weather_widget, err := weather_service.GetWeatherWidgetByID(1)
	assert.NoError(t, err)
	assert.NotNil(t, weather_widget)
	assert.Equal(t, 1, weather_widget.ID)
	assert.Equal(t, "11356", weather_widget.ZIPCODE)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateWeatherWidget(t *testing.T) {
	db, mock := tests.SetupMockDB(t)
	defer db.Close()

	weather_service := servicesWeatherWidget.NewWeatherService(db, "test_api_key")

	weather_widget := &modelsWidget.WeatherWidget{
		ID:      1,
		ZIPCODE: "11357",
	}

	mock.ExpectExec(regexp.QuoteMeta(`
		UPDATE "WeatherWidget"
		SET zipcode = $1
		WHERE id = $2;
	`)).
		WithArgs(weather_widget.ZIPCODE, weather_widget.ID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := weather_service.UpdateWeatherWidget(weather_widget)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteWeatherWidget(t *testing.T) {
	db, mock := tests.SetupMockDB(t)
	defer db.Close()

	weather_service := servicesWeatherWidget.NewWeatherService(db, "test_api_key")

	mock.ExpectExec(regexp.QuoteMeta(`
		DELETE FROM "WeatherWidget"
		WHERE id = $1;
	`)).
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := weather_service.DeleteWeatherWidget(1)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetWeatherOverview(t *testing.T) {
	db, _ := tests.SetupMockDB(t)
	defer db.Close()
	err := godotenv.Load("../../../../.env")
	if err != nil {
		assert.NoError(t, err)
	}
	apiKey := os.Getenv("OPENWEATHER_API_KEY")

	weather_service := servicesWeatherWidget.NewWeatherService(db, apiKey)

	weather_widget := &modelsWidget.WeatherWidget{
		ID:      1,
		ZIPCODE: "11357",
	}

	weather_overview, err := weather_service.GetWeatherOverview(weather_widget)

	fmt.Printf("%s", weather_overview)
	assert.NoError(t, err)
}
