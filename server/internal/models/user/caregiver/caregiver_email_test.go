package caregiver

import (
	"errors"
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

func TestGetCaregiverByEmail_NotFoundIsNilNotAnError(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()
	mock.ExpectQuery(regexp.QuoteMeta(`WHERE lower(email) = lower($1)`)).
		WithArgs("none@example.com").
		WillReturnRows(sqlmock.NewRows(caregiverColumns))

	c, err := GetCaregiverByEmail(db, "none@example.com")
	assert.NoError(t, err)
	assert.Nil(t, c)
}

func TestGetCaregiverByEmail_DatabaseError(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()
	mock.ExpectQuery(regexp.QuoteMeta(`WHERE lower(email) = lower($1)`)).WillReturnError(errors.New("db down"))

	c, err := GetCaregiverByEmail(db, "pat@example.com")
	assert.Error(t, err)
	assert.Nil(t, c)
}

func TestSetGoogleID_OnlyFillsAnEmptyGoogleID(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()
	mock.ExpectExec(regexp.QuoteMeta(`SET google_id = $1 WHERE id = $2 AND google_id IS NULL`)).
		WithArgs("g-1", 4).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(`SET google_id = $1 WHERE id = $2 AND google_id IS NULL`)).
		WithArgs("g-2", 5).
		WillReturnResult(sqlmock.NewResult(0, 0))

	changed, err := SetGoogleID(db, 4, "g-1")
	require.NoError(t, err)
	assert.True(t, changed)

	changed, err = SetGoogleID(db, 5, "g-2")
	require.NoError(t, err)
	assert.False(t, changed, "a caregiver that already has a Google ID is left alone")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSetGoogleID_DatabaseError(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()
	mock.ExpectExec(regexp.QuoteMeta(`SET google_id`)).WillReturnError(errors.New("db down"))

	changed, err := SetGoogleID(db, 4, "g-1")
	assert.Error(t, err)
	assert.False(t, changed)
}

func TestSetEmail_OnlyFillsAnEmptyEmailThatNobodyElseUses(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()
	const fragment = `SET email = $1 WHERE id = $2 AND email IS NULL AND NOT EXISTS (SELECT 1 FROM "Caregiver" WHERE lower(email) = lower($1))`
	mock.ExpectExec(regexp.QuoteMeta(fragment)).WithArgs("pat@example.com", 4).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(fragment)).WithArgs("taken@example.com", 5).WillReturnResult(sqlmock.NewResult(0, 0))

	changed, err := SetEmail(db, 4, "pat@example.com")
	require.NoError(t, err)
	assert.True(t, changed)

	changed, err = SetEmail(db, 5, "taken@example.com")
	require.NoError(t, err)
	assert.False(t, changed, "an email another caregiver uses, or one already set, changes nothing")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSetEmail_DatabaseError(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()
	mock.ExpectExec(regexp.QuoteMeta(`SET email`)).WillReturnError(errors.New("db down"))

	changed, err := SetEmail(db, 4, "pat@example.com")
	assert.Error(t, err)
	assert.False(t, changed)
}
