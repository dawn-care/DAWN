package prompts

import (
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

// SETUP

func setupMockDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock database: %v", err)
	}
	return db, mock
}

// CREATE

func TestCreateCaregiverLocation(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	location := &CaregiverLocationPrompt{
		Location: "1 Lomb Memorial Driver, Rochester NY, 14623",
		Message:  "Helena is only 14 minutes away!",
		URI:      "C:\\image1.png",
	}

	mock.ExpectQuery(`INSERT\s+INTO\s+"CaregiverLocationPrompt"`).
		WithArgs(location.Location, location.Message, location.URI).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	err := CreateCaregiverLocation(db, location)
	assert.NoError(t, err)
	assert.Equal(t, 1, location.ID)

	assert.NoError(t, mock.ExpectationsWereMet())
}

// READ

func TestReadCaregiverLocationByID(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	rows := sqlmock.NewRows([]string{"id", "location", "message", "image_uri"}).
		AddRow(
			1,
			"1 Lomb Memorial Driver, Rochester NY, 14623",
			"Helena is only 14 minutes away!",
			"C:\\image1.png",
		)

	mock.ExpectQuery(`(?s)SELECT\s+id,\s+location,\s+message,\s+image_uri\s+FROM\s+"CaregiverLocationPrompt"\s+WHERE\s+id\s*=\s*\$1`).
		WithArgs(1).
		WillReturnRows(rows)

	got, err := GetCaregiverLocationByID(db, 1)
	assert.NoError(t, err)
	assert.NotNil(t, got)
	assert.Equal(t, "1 Lomb Memorial Driver, Rochester NY, 14623", got.Location)
	assert.Equal(t, "Helena is only 14 minutes away!", got.Message)
	assert.Equal(t, "C:\\image1.png", got.URI)

	assert.NoError(t, mock.ExpectationsWereMet())
}

// UPDATE

func TestUpdateCaregiverLocation(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	location := &CaregiverLocationPrompt{
		ID:       1,
		Location: "1 Lomb Memorial Driver, Rochester NY, 14623",
		Message:  "Helena is only 10 minutes away!",
		URI:      "C:\\image1.png",
	}

	mock.ExpectExec(`UPDATE\s+"CaregiverLocationPrompt"\s+SET`).
		WithArgs(location.Location, location.Message, location.URI, location.ID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := UpdateCaregiverLocation(db, location)
	assert.NoError(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

// DELETE

func TestDeleteCaregiverLocation(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	mock.ExpectExec(`DELETE\s+FROM\s+"CaregiverLocationPrompt"`).
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := DeleteCaregiverLocation(db, 1)
	assert.NoError(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}
