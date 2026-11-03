package auth

import (
	"context"
	"errors"
	"regexp"
	"testing"

	"server/internal/services/user"
	"server/internal/services/user/caregiver"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

/* -------------------------------------------------------
   SETUP
------------------------------------------------------- */

type fakeGoogle struct {
	user *user.GoogleUser
	err  error
}

func (f fakeGoogle) Verify(ctx context.Context, idToken string) (*user.GoogleUser, error) {
	return f.user, f.err
}

// fakeSessions records what the auth service asked of the session service
type fakeSessions struct {
	createToken  string
	createErr    error
	foundSession ValidatedSession
	findErr      error
	deleteErr    error

	createdFor  []int
	foundTokens []string
	deletedFor  []int
}

func (f *fakeSessions) Create(ctx context.Context, userID int) (string, error) {
	f.createdFor = append(f.createdFor, userID)
	return f.createToken, f.createErr
}

func (f *fakeSessions) Validate(ctx context.Context, token string) (ValidatedSession, error) {
	return ValidatedSession{}, errors.New("not used: logout looks the session up with Find, which does not renew it")
}

func (f *fakeSessions) Find(ctx context.Context, token string) (ValidatedSession, error) {
	f.foundTokens = append(f.foundTokens, token)
	return f.foundSession, f.findErr
}

func (f *fakeSessions) DeleteAllForUser(ctx context.Context, userID int) error {
	f.deletedFor = append(f.deletedFor, userID)
	return f.deleteErr
}

func newAuthService(t *testing.T, google GoogleAuthService, sessions SessionService) (*AuthService, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock database: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return NewAuthService(google, caregiver.NewCaregiverService(db), nil, sessions), mock
}

/* -------------------------------------------------------
   LOGIN
------------------------------------------------------- */

func TestLoginWithGoogle_ExistingCaregiverGetsASession(t *testing.T) {
	sessions := &fakeSessions{createToken: "new-token"}
	svc, mock := newAuthService(t, fakeGoogle{user: &user.GoogleUser{ID: "g-1", Name: "Pat"}}, sessions)
	mock.ExpectQuery(regexp.QuoteMeta(`FROM "Caregiver"`)).
		WithArgs("g-1").
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "role", "google_id"}).AddRow(7, "Pat", "caregiver", "g-1"))

	token, err := svc.LoginWithGoogle(context.Background(), "id-token")

	require.NoError(t, err)
	assert.Equal(t, "new-token", token)
	assert.Equal(t, []int{7}, sessions.createdFor)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestLoginWithGoogle_NewCaregiverIsCreatedThenGetsASession(t *testing.T) {
	sessions := &fakeSessions{createToken: "new-token"}
	svc, mock := newAuthService(t, fakeGoogle{user: &user.GoogleUser{ID: "g-2", Name: "Sam"}}, sessions)
	mock.ExpectQuery(regexp.QuoteMeta(`FROM "Caregiver"`)).
		WithArgs("g-2").
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "role", "google_id"}))
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "Caregiver"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(12))

	token, err := svc.LoginWithGoogle(context.Background(), "id-token")

	require.NoError(t, err)
	assert.Equal(t, "new-token", token)
	assert.Equal(t, []int{12}, sessions.createdFor)
}

func TestLoginWithGoogle_CaregiverLookupErrorStartsNoSession(t *testing.T) {
	sessions := &fakeSessions{}
	svc, mock := newAuthService(t, fakeGoogle{user: &user.GoogleUser{ID: "g-1"}}, sessions)
	mock.ExpectQuery(regexp.QuoteMeta(`FROM "Caregiver"`)).WillReturnError(errors.New("db down"))

	token, err := svc.LoginWithGoogle(context.Background(), "id-token")

	assert.Error(t, err)
	assert.NotErrorIs(t, err, ErrGoogleTokenRejected, "a database failure is not a bad Google token")
	assert.Empty(t, token)
	assert.Empty(t, sessions.createdFor)
}

func TestLoginWithGoogle_FailedGoogleCheckStartsNoSession(t *testing.T) {
	sessions := &fakeSessions{}
	svc, _ := newAuthService(t, fakeGoogle{err: errors.New("bad token")}, sessions)

	token, err := svc.LoginWithGoogle(context.Background(), "id-token")

	assert.ErrorIs(t, err, ErrGoogleTokenRejected)
	assert.ErrorContains(t, err, "bad token", "Google's reason is kept for the log")
	assert.Empty(t, token)
	assert.Empty(t, sessions.createdFor)
}

func TestStartSession(t *testing.T) {
	sessions := &fakeSessions{createToken: "tok"}
	svc, _ := newAuthService(t, nil, sessions)

	token, err := svc.StartSession(context.Background(), 5)

	require.NoError(t, err)
	assert.Equal(t, "tok", token)
	assert.Equal(t, []int{5}, sessions.createdFor)
}

func TestStartSession_Error(t *testing.T) {
	sessions := &fakeSessions{createErr: errors.New("db down")}
	svc, _ := newAuthService(t, nil, sessions)

	token, err := svc.StartSession(context.Background(), 5)

	assert.Error(t, err)
	assert.Empty(t, token)
}

/* -------------------------------------------------------
   LOGOUT EVERYWHERE
------------------------------------------------------- */

func TestLogoutEverywhere_EndsEverySessionOfThatCaregiver(t *testing.T) {
	sessions := &fakeSessions{foundSession: ValidatedSession{UserID: 7}}
	svc, _ := newAuthService(t, nil, sessions)

	err := svc.LogoutEverywhere(context.Background(), "cookie-token")

	require.NoError(t, err)
	assert.Equal(t, []string{"cookie-token"}, sessions.foundTokens)
	assert.Equal(t, []int{7}, sessions.deletedFor)
}

func TestLogoutEverywhere_UnknownOrExpiredTokenHasNothingToEnd(t *testing.T) {
	for _, findErr := range []error{ErrSessionNotFound, ErrSessionExpired} {
		sessions := &fakeSessions{findErr: findErr}
		svc, _ := newAuthService(t, nil, sessions)

		err := svc.LogoutEverywhere(context.Background(), "cookie-token")

		assert.NoError(t, err, findErr.Error())
		assert.Empty(t, sessions.deletedFor, findErr.Error())
	}
}

func TestLogoutEverywhere_DatabaseErrorOnLookupIsReported(t *testing.T) {
	sessions := &fakeSessions{findErr: errors.New("db down")}
	svc, _ := newAuthService(t, nil, sessions)

	err := svc.LogoutEverywhere(context.Background(), "cookie-token")

	assert.Error(t, err)
	assert.Empty(t, sessions.deletedFor)
}

func TestLogoutEverywhere_DeleteErrorIsReported(t *testing.T) {
	sessions := &fakeSessions{foundSession: ValidatedSession{UserID: 7}, deleteErr: errors.New("db down")}
	svc, _ := newAuthService(t, nil, sessions)

	err := svc.LogoutEverywhere(context.Background(), "cookie-token")

	assert.Error(t, err, "saying logged out while sessions remain would be wrong")
	assert.ErrorContains(t, err, "caregiver 7", "the log says whose sessions are left")
}
