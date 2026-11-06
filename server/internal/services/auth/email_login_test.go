package auth

import (
	"context"
	"errors"
	"regexp"
	"testing"

	"server/internal/services/auth/authtest"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var findByEmail = regexp.QuoteMeta(`WHERE lower(email) = lower($1)`)

var caregiverColumns = []string{"id", "name", "role", "google_id", "email"}

/* -------------------------------------------------------
   LoginWithEmailCode
------------------------------------------------------- */

func TestLoginWithEmailCode_ExistingCaregiverGetsASession(t *testing.T) {
	codes := &authtest.FakeCodes{}
	sessions := &fakeSessions{createToken: "new-token"}
	svc, mock := newAuthService(t, fakeGoogle{}, sessions)
	svc.WithLoginCodes(codes)
	mock.ExpectQuery(findByEmail).WithArgs("pat@example.com").
		WillReturnRows(sqlmock.NewRows(caregiverColumns).AddRow(9, "Pat", "caregiver", "g-1", "pat@example.com"))

	token, err := svc.LoginWithEmailCode(context.Background(), "Pat@Example.com", "123456", "203.0.113.9")

	require.NoError(t, err)
	assert.Equal(t, "new-token", token)
	assert.Equal(t, []int{9}, sessions.createdFor, "the caregiver with this email, including one who signed up with Google")
	assert.Equal(t, []string{"Pat@Example.com|123456|203.0.113.9"}, codes.Verified)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestLoginWithEmailCode_AWrongCodeStartsNoSessionAndCreatesNoCaregiver(t *testing.T) {
	for _, rejection := range []error{ErrCodeRejected, ErrTooManyAttempts, errors.New("db down")} {
		sessions := &fakeSessions{}
		svc, mock := newAuthService(t, fakeGoogle{}, sessions)
		svc.WithLoginCodes(&authtest.FakeCodes{VerifyErr: rejection})

		token, err := svc.LoginWithEmailCode(context.Background(), "pat@example.com", "000000", "203.0.113.9")

		assert.ErrorIs(t, err, rejection)
		assert.Empty(t, token)
		assert.Empty(t, sessions.createdFor)
		assert.NoError(t, mock.ExpectationsWereMet(), "the caregiver table is not touched without a right code")
	}
}
