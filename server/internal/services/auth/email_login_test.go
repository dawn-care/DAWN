package auth

import (
	"context"
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeCodes is a LoginCodeService that records what it was asked
type fakeCodes struct {
	requested  []string // "address|ip"
	verified   []string // "address|code"
	requestErr error
	verifyErr  error
}

func (f *fakeCodes) RequestCode(ctx context.Context, address, ip string) error {
	f.requested = append(f.requested, address+"|"+ip)
	return f.requestErr
}

func (f *fakeCodes) VerifyCode(ctx context.Context, address, code string) error {
	f.verified = append(f.verified, address+"|"+code)
	return f.verifyErr
}

var findByEmail = regexp.QuoteMeta(`WHERE lower(email) = lower($1)`)

var caregiverColumns = []string{"id", "name", "role", "google_id", "email"}

/* -------------------------------------------------------
   RequestEmailCode
------------------------------------------------------- */

func TestEmailLogin_DisabledWhenNotSetUp(t *testing.T) {
	sessions := &fakeSessions{}
	svc, mock := newAuthService(t, fakeGoogle{}, sessions)

	assert.ErrorIs(t, svc.RequestEmailCode(context.Background(), "pat@example.com", "203.0.113.9"), ErrEmailLoginDisabled)
	token, err := svc.LoginWithEmailCode(context.Background(), "pat@example.com", "123456")
	assert.ErrorIs(t, err, ErrEmailLoginDisabled)
	assert.Empty(t, token)
	assert.Empty(t, sessions.createdFor)
	assert.NoError(t, mock.ExpectationsWereMet())
}

/* -------------------------------------------------------
   LoginWithEmailCode
------------------------------------------------------- */

func TestLoginWithEmailCode_ExistingCaregiverGetsASession(t *testing.T) {
	codes := &fakeCodes{}
	sessions := &fakeSessions{createToken: "new-token"}
	svc, mock := newAuthService(t, fakeGoogle{}, sessions)
	svc.WithLoginCodes(codes)
	mock.ExpectQuery(findByEmail).WithArgs("pat@example.com").
		WillReturnRows(sqlmock.NewRows(caregiverColumns).AddRow(9, "Pat", "caregiver", "g-1", "pat@example.com"))

	token, err := svc.LoginWithEmailCode(context.Background(), "Pat@Example.com", "123456")

	require.NoError(t, err)
	assert.Equal(t, "new-token", token)
	assert.Equal(t, []int{9}, sessions.createdFor, "the caregiver with this email, including one who signed up with Google")
	assert.Equal(t, []string{"Pat@Example.com|123456"}, codes.verified)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestLoginWithEmailCode_AWrongCodeStartsNoSessionAndCreatesNoCaregiver(t *testing.T) {
	for _, rejection := range []error{ErrCodeRejected, ErrTooManyAttempts, errors.New("db down")} {
		sessions := &fakeSessions{}
		svc, mock := newAuthService(t, fakeGoogle{}, sessions)
		svc.WithLoginCodes(&fakeCodes{verifyErr: rejection})

		token, err := svc.LoginWithEmailCode(context.Background(), "pat@example.com", "000000")

		assert.ErrorIs(t, err, rejection)
		assert.Empty(t, token)
		assert.Empty(t, sessions.createdFor)
		assert.NoError(t, mock.ExpectationsWereMet(), "the caregiver table is not touched without a right code")
	}
}
