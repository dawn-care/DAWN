package caregiver

import (
	"regexp"
	"testing"

	"server/internal/services/user"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	byGoogleIDSQL  = `WHERE google_id = $1`
	byEmailSQL     = `WHERE lower(email) = lower($1)`
	setEmailSQL    = `SET email = $1`
	setGoogleIDSQL = `SET google_id = $1`
	insertSQL      = `INSERT INTO "Caregiver" (name, role, google_id, email)`
)

var columns = []string{"id", "name", "role", "google_id", "email"}

func newMockedService(t *testing.T) (*CaregiverService, sqlmock.Sqlmock) {
	db, mock := setupMockDB(t)
	t.Cleanup(func() { db.Close() })
	return NewCaregiverService(db), mock
}

func googleUser(email string, verified bool) *user.GoogleUser {
	return &user.GoogleUser{ID: "g-1", Name: "Pat", Email: email, EmailVerified: verified}
}

/* -------------------------------------------------------
   FindOrCreateFromGoogle
------------------------------------------------------- */

func TestFindOrCreateFromGoogle_KnownGoogleIDWithoutEmailSavesTheVerifiedEmail(t *testing.T) {
	svc, mock := newMockedService(t)
	mock.ExpectQuery(regexp.QuoteMeta(byGoogleIDSQL)).WithArgs("g-1").
		WillReturnRows(sqlmock.NewRows(columns).AddRow(7, "Pat", "caregiver", "g-1", nil))
	mock.ExpectExec(regexp.QuoteMeta(setEmailSQL)).WithArgs("pat@gmail.com", 7).
		WillReturnResult(sqlmock.NewResult(0, 1))

	c, err := svc.FindOrCreateFromGoogle(googleUser("  Pat@Gmail.COM ", true))

	require.NoError(t, err)
	assert.Equal(t, 7, c.ID)
	require.NotNil(t, c.Email)
	assert.Equal(t, "pat@gmail.com", *c.Email, "stored trimmed and lowercase")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestFindOrCreateFromGoogle_KnownGoogleIDDoesNotSaveAnUnverifiedEmail(t *testing.T) {
	svc, mock := newMockedService(t)
	mock.ExpectQuery(regexp.QuoteMeta(byGoogleIDSQL)).WithArgs("g-1").
		WillReturnRows(sqlmock.NewRows(columns).AddRow(7, "Pat", "caregiver", "g-1", nil))

	c, err := svc.FindOrCreateFromGoogle(googleUser("pat@gmail.com", false))

	require.NoError(t, err)
	assert.Nil(t, c.Email)
	assert.NoError(t, mock.ExpectationsWereMet(), "no update was expected")
}

func TestFindOrCreateFromGoogle_VerifiedEmailOfAnEmailCaregiverLinksTheGoogleAccount(t *testing.T) {
	svc, mock := newMockedService(t)
	mock.ExpectQuery(regexp.QuoteMeta(byGoogleIDSQL)).WithArgs("g-1").
		WillReturnRows(sqlmock.NewRows(columns))
	mock.ExpectQuery(regexp.QuoteMeta(byEmailSQL)).WithArgs("pat@gmail.com").
		WillReturnRows(sqlmock.NewRows(columns).AddRow(9, "pat", "caregiver", nil, "pat@gmail.com"))
	mock.ExpectExec(regexp.QuoteMeta(setGoogleIDSQL)).WithArgs("g-1", 9).
		WillReturnResult(sqlmock.NewResult(0, 1))

	c, err := svc.FindOrCreateFromGoogle(googleUser("Pat@gmail.com", true))

	require.NoError(t, err)
	assert.Equal(t, 9, c.ID, "the existing account, not a new one")
	require.NotNil(t, c.GoogleID)
	assert.Equal(t, "g-1", *c.GoogleID)
	assert.NoError(t, mock.ExpectationsWereMet(), "nothing is inserted")
}

func TestFindOrCreateFromGoogle_EmailAlreadyLinkedToAnotherGoogleAccountIsNotTakenOver(t *testing.T) {
	svc, mock := newMockedService(t)
	mock.ExpectQuery(regexp.QuoteMeta(byGoogleIDSQL)).WithArgs("g-1").
		WillReturnRows(sqlmock.NewRows(columns))
	mock.ExpectQuery(regexp.QuoteMeta(byEmailSQL)).WithArgs("pat@gmail.com").
		WillReturnRows(sqlmock.NewRows(columns).AddRow(9, "Old Pat", "caregiver", "g-other", "pat@gmail.com"))
	// A new caregiver is created without the address, which the old account still holds
	mock.ExpectQuery(regexp.QuoteMeta(insertSQL)).WithArgs("Pat", "caregiver", "g-1", nil).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(20))

	c, err := svc.FindOrCreateFromGoogle(googleUser("pat@gmail.com", true))

	require.NoError(t, err)
	assert.Equal(t, 20, c.ID)
	assert.Nil(t, c.Email)
	assert.NoError(t, mock.ExpectationsWereMet(), "the other account was not linked")
}

func TestFindOrCreateFromGoogle_UnverifiedEmailIsNeverMatchedOrStored(t *testing.T) {
	svc, mock := newMockedService(t)
	mock.ExpectQuery(regexp.QuoteMeta(byGoogleIDSQL)).WithArgs("g-1").
		WillReturnRows(sqlmock.NewRows(columns))
	// no email lookup is expected: matching on an address Google did not verify would let
	// anyone claim an account by registering its address at Google
	mock.ExpectQuery(regexp.QuoteMeta(insertSQL)).WithArgs("Pat", "caregiver", "g-1", nil).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(22))

	c, err := svc.FindOrCreateFromGoogle(googleUser("victim@gmail.com", false))

	require.NoError(t, err)
	assert.Equal(t, 22, c.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestFindOrCreateFromGoogle_NewCaregiverGetsTheVerifiedEmail(t *testing.T) {
	svc, mock := newMockedService(t)
	mock.ExpectQuery(regexp.QuoteMeta(byGoogleIDSQL)).WithArgs("g-1").
		WillReturnRows(sqlmock.NewRows(columns))
	mock.ExpectQuery(regexp.QuoteMeta(byEmailSQL)).WithArgs("pat@gmail.com").
		WillReturnRows(sqlmock.NewRows(columns))
	mock.ExpectQuery(regexp.QuoteMeta(insertSQL)).WithArgs("Pat", "caregiver", "g-1", "pat@gmail.com").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(23))

	c, err := svc.FindOrCreateFromGoogle(googleUser("Pat@Gmail.com", true))

	require.NoError(t, err)
	assert.Equal(t, 23, c.ID)
	assert.Equal(t, "pat@gmail.com", *c.Email)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestFindOrCreateFromEmail_ReturnsTheExistingCaregiverIgnoringCase(t *testing.T) {
	svc, mock := newMockedService(t)
	mock.ExpectQuery(regexp.QuoteMeta(byEmailSQL)).WithArgs("pat@gmail.com").
		WillReturnRows(sqlmock.NewRows(columns).AddRow(9, "Pat", "caregiver", "g-1", "pat@gmail.com"))

	c, err := svc.FindOrCreateFromEmail("  PAT@Gmail.com ")

	require.NoError(t, err)
	assert.Equal(t, 9, c.ID, "the Google caregiver with this email, not a new account")
	assert.NoError(t, mock.ExpectationsWereMet(), "nothing is inserted")
}

func TestFindOrCreateFromEmail_CreatesACaregiverNamedAfterTheAddress(t *testing.T) {
	svc, mock := newMockedService(t)
	mock.ExpectQuery(regexp.QuoteMeta(byEmailSQL)).WithArgs("jo.smith@gmail.com").
		WillReturnRows(sqlmock.NewRows(columns))
	mock.ExpectQuery(regexp.QuoteMeta(insertSQL)).WithArgs("jo.smith", "caregiver", nil, "jo.smith@gmail.com").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(30))

	c, err := svc.FindOrCreateFromEmail("Jo.Smith@Gmail.com")

	require.NoError(t, err)
	assert.Equal(t, 30, c.ID)
	assert.Equal(t, "jo.smith", c.Name)
	assert.Equal(t, "caregiver", c.Role)
	assert.Nil(t, c.GoogleID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestFindOrCreateFromEmail_UsesTheCaregiverAnotherSignInCreatedFirst(t *testing.T) {
	svc, mock := newMockedService(t)
	mock.ExpectQuery(regexp.QuoteMeta(byEmailSQL)).WithArgs("pat@gmail.com").WillReturnRows(sqlmock.NewRows(columns))
	mock.ExpectQuery(regexp.QuoteMeta(insertSQL)).
		WillReturnError(&pq.Error{Code: "23505", Constraint: "Caregiver_email_lower_key"})
	mock.ExpectQuery(regexp.QuoteMeta(byEmailSQL)).WithArgs("pat@gmail.com").
		WillReturnRows(sqlmock.NewRows(columns).AddRow(31, "pat", "caregiver", nil, "pat@gmail.com"))

	c, err := svc.FindOrCreateFromEmail("pat@gmail.com")

	require.NoError(t, err)
	assert.Equal(t, 31, c.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestFindOrCreateFromEmail_RejectsAnEmptyEmail(t *testing.T) {
	svc, mock := newMockedService(t)

	c, err := svc.FindOrCreateFromEmail("   ")

	assert.Error(t, err)
	assert.Nil(t, c)
	assert.NoError(t, mock.ExpectationsWereMet(), "no query ran")
}

func TestNormalizeEmail(t *testing.T) {
	assert.Equal(t, "pat@gmail.com", NormalizeEmail("  Pat@GMAIL.com\t"))
	assert.Equal(t, "", NormalizeEmail("   "))
}
