package widgets

import (
	"database/sql"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func setupMockDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock database: %v", err)
	}
	return db, mock
}

// CREATE & RETRIEVE

func TestCreateWeatherWidget(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	w := &WeatherWidget{ZIPCODE: "14623"}

	mock.ExpectQuery(regexp.QuoteMeta(`
		INSERT INTO "WeatherWidget" (zipcode)
		VALUES ($1)
		RETURNING id;
	`)).
		WithArgs(w.ZIPCODE).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, zipcode
		FROM "WeatherWidget"
		WHERE id = $1;
	`)).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "zipcode"}).AddRow(1, "14623"))

	err := CreateWeather(db, w)
	assert.NoError(t, err)
	assert.Equal(t, 1, w.ID)

	found, err := GetWeatherByID(db, 1)
	assert.NoError(t, err)
	assert.NotNil(t, found)
	assert.Equal(t, "14623", found.ZIPCODE)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateWeatherBulkWidgets(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	widgets := []WeatherWidget{
		{ZIPCODE: "14623"},
		{ZIPCODE: "14618"},
		{ZIPCODE: "14626"},
	}

	mock.ExpectBegin()

	mock.ExpectPrepare(regexp.QuoteMeta(`
		INSERT INTO "WeatherWidget" (zipcode)
		VALUES ($1)
		RETURNING id;
	`))

	for i, w := range widgets {
		mock.ExpectQuery(regexp.QuoteMeta(`
			INSERT INTO "WeatherWidget" (zipcode)
			VALUES ($1)
			RETURNING id;
		`)).
			WithArgs(w.ZIPCODE).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(i + 1))
	}

	mock.ExpectCommit()

	err := CreateWeathersBulk(db, widgets)
	assert.NoError(t, err)

	for i, w := range widgets {
		assert.Equal(t, i+1, w.ID)
	}

	assert.NoError(t, mock.ExpectationsWereMet())
}

// UPDATE

func TestUpdateWeatherWidget(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	w := &WeatherWidget{
		ID:      1,
		ZIPCODE: "14618",
	}

	mock.ExpectExec(regexp.QuoteMeta(`
		UPDATE "WeatherWidget"
		SET zipcode = $1
		WHERE id = $2;
	`)).
		WithArgs(w.ZIPCODE, w.ID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, zipcode
		FROM "WeatherWidget"
		WHERE id = $1;
	`)).
		WithArgs(w.ID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "zipcode"}).AddRow(1, "14618"))

	err := UpdateWeather(db, w)
	assert.NoError(t, err)

	found, err := GetWeatherByID(db, w.ID)
	assert.NoError(t, err)
	assert.Equal(t, "14618", found.ZIPCODE)

	assert.NoError(t, mock.ExpectationsWereMet())
}

// DELETE

func TestDeleteWeatherWidget(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	mock.ExpectExec(regexp.QuoteMeta(`
		DELETE FROM "WeatherWidget"
		WHERE id = $1;
	`)).
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, zipcode
		FROM "WeatherWidget"
		WHERE id = $1;
	`)).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{}))

	err := DeleteWeather(db, 1)
	assert.NoError(t, err)

	found, err := GetWeatherByID(db, 1)
	assert.NoError(t, err)
	assert.Nil(t, found)

	assert.NoError(t, mock.ExpectationsWereMet())
}

// ANALYSIS & STATS

func TestGetWeatherStatsWidget(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	// Expect count(*)
	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT COUNT(*)
		FROM "WeatherWidget";
	`)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(3))

	count, err := GetWeatherStats(db)
	assert.NoError(t, err)
	assert.Equal(t, 3, count)

	assert.NoError(t, mock.ExpectationsWereMet())
}
