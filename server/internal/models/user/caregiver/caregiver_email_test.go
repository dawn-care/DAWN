package caregiver

import (
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var caregiverColumns = []string{"id", "name", "role", "google_id", "email"}

func TestCreateCaregiver_SavesEmail(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()
	email := "pat@example.com"
	c := &Caregiver{Name: "pat", Role: "caregiver", Email: &email}

	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "Caregiver" (name, role, google_id, email)`)).
		WithArgs("pat", "caregiver", nil, "pat@example.com").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(3))

	require.NoError(t, CreateCaregiver(db, c))
	assert.Equal(t, 3, c.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetCaregiverByEmail_FindsIgnoringCase(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	mock.ExpectQuery(regexp.QuoteMeta(`WHERE lower(email) = lower($1)`)).
		WithArgs("PAT@Example.com").
		WillReturnRows(sqlmock.NewRows(caregiverColumns).AddRow(4, "pat", "caregiver", nil, "pat@example.com"))

	c, err := GetCaregiverByEmail(db, "PAT@Example.com")
	require.NoError(t, err)
	require.NotNil(t, c)
	assert.Equal(t, 4, c.ID)
	assert.Nil(t, c.GoogleID)
	assert.Equal(t, "pat@example.com", *c.Email)
	assert.NoError(t, mock.ExpectationsWereMet())
}
