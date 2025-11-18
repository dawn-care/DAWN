package prompts

import (
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

// CREATE

func TestCreateCarepartnerLocation(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	location := &CarepartnerLocationPrompt{
		Location: "Care facility",
		Message:  "You are here.",
		URI:      "C:\\image1.png",
	}

	mock.ExpectQuery(regexp.QuoteMeta(
		`INSERT INTO "CarepartnerLocationPrompt" (location, message, image_uri) VALUES ($1, $2, $3) RETURNING id;`,
	)).
		WithArgs(location.Location, location.Message, location.URI).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	err := CreateCarepartnerLocation(db, location)
	assert.NoError(t, err)
	assert.Equal(t, 1, location.ID)

	assert.NoError(t, mock.ExpectationsWereMet())
}

// READ

func TestReadCarepartnerLocationByID(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	rows := sqlmock.NewRows([]string{"id", "location", "message", "image_uri"}).
		AddRow(
			1,
			"Care facility",
			"You are here.",
			"C:\\image1.png",
		)

	mock.ExpectQuery(`(?s)SELECT\s+id,\s+location,\s+message,\s+image_uri\s+FROM\s+"CarepartnerLocationPrompt"\s+WHERE\s+id\s*=\s*\$1`).
		WithArgs(1).
		WillReturnRows(rows)

	got, err := GetCarepartnerLocationByID(db, 1)
	assert.NoError(t, err)
	assert.NotNil(t, got)
	assert.Equal(t, "Care facility", got.Location)
	assert.Equal(t, "You are here.", got.Message)
	assert.Equal(t, "C:\\image1.png", got.URI)

	assert.NoError(t, mock.ExpectationsWereMet())
}

// UPDATE

func TestUpdateCarepartnerLocation(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	location := &CarepartnerLocationPrompt{
		ID:       1,
		Location: "Care facility",
		Message:  "You chose to be here, your family will visit!",
		URI:      "C:\\image1.png",
	}

	mock.ExpectExec(`(?s)UPDATE\s+"CarepartnerLocationPrompt"\s+SET\s+location\s*=\s*\$1,\s+message\s*=\s*\$2,\s+image_uri\s*=\s*\$3\s+WHERE\s+id\s*=\s*\$4`).
		WithArgs(location.Location, location.Message, location.URI, location.ID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := UpdateCarepartnerLocation(db, location)
	assert.NoError(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

// DELETE

func TestDeleteCarepartnerLocation(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	mock.ExpectExec(`(?s)DELETE\s+FROM\s+"CarepartnerLocationPrompt"\s+WHERE\s+id\s*=\s*\$1`).
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := DeleteCarepartnerLocation(db, 1)
	assert.NoError(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}
