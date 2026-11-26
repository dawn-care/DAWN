package auth

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	authsvc "server/internal/services/auth"
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

	verified []string
}

func (f *fakeGoogle) Verify(ctx context.Context, idToken string) (*user.GoogleUser, error) {
	f.verified = append(f.verified, idToken)
	return f.user, f.err
}

type fakeSessions struct {
	createToken  string
	createErr    error
	foundSession authsvc.ValidatedSession
	findErr      error
	deleteErr    error

	createdFor []int
	deletedFor []int
	found      int
}

func (f *fakeSessions) Create(ctx context.Context, userID int) (string, error) {
	f.createdFor = append(f.createdFor, userID)
	return f.createToken, f.createErr
}

func (f *fakeSessions) Validate(ctx context.Context, token string) (authsvc.ValidatedSession, error) {
	return authsvc.ValidatedSession{}, errors.New("not used: logout looks the session up with Find, which does not renew it")
}

func (f *fakeSessions) Find(ctx context.Context, token string) (authsvc.ValidatedSession, error) {
	f.found++
	return f.foundSession, f.findErr
}

func (f *fakeSessions) DeleteAllForUser(ctx context.Context, userID int) error {
	f.deletedFor = append(f.deletedFor, userID)
	return f.deleteErr
}

func newController(t *testing.T, google authsvc.GoogleAuthService, sessions authsvc.SessionService) (*AuthController, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock database: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	service := authsvc.NewAuthService(google, caregiver.NewCaregiverService(db), nil, sessions)
	return &AuthController{AuthService: service}, mock
}

func expectExistingCaregiver(mock sqlmock.Sqlmock, id int) {
	mock.ExpectQuery(regexp.QuoteMeta(`FROM "Caregiver"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "role", "google_id", "email"}).AddRow(id, "Pat", "caregiver", "g-1", nil))
}

func login(c *AuthController) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/api/auth/google", strings.NewReader(`{"idToken":"id-token"}`))
	w := httptest.NewRecorder()
	c.GoogleLogin(w, req)
	return w
}

func logout(c *AuthController, cookie *http.Cookie) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)
	if cookie != nil {
		req.AddCookie(cookie)
	}
	w := httptest.NewRecorder()
	c.Logout(w, req)
	return w
}

func sessionCookie(token string) *http.Cookie {
	return &http.Cookie{Name: "session", Value: token}
}

func assertCleared(t *testing.T, w *httptest.ResponseRecorder) {
	t.Helper()
	cookies := w.Result().Cookies()
	require.Len(t, cookies, 1)
	assert.Equal(t, "session", cookies[0].Name)
	assert.Equal(t, "", cookies[0].Value)
	assert.Less(t, cookies[0].MaxAge, 0)
}

/* -------------------------------------------------------
   GOOGLE LOGIN
------------------------------------------------------- */

func TestGoogleLogin_SetsALongLivedStrictCookie(t *testing.T) {
	t.Setenv("RUN_MODE", "production")
	sessions := &fakeSessions{createToken: "new-token"}
	c, mock := newController(t, &fakeGoogle{user: &user.GoogleUser{ID: "g-1", Name: "Pat"}}, sessions)
	expectExistingCaregiver(mock, 7)

	w := login(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"status":"authenticated"}`, w.Body.String())
	assert.Equal(t, []int{7}, sessions.createdFor)
	cookies := w.Result().Cookies()
	require.Len(t, cookies, 1)
	cookie := cookies[0]
	assert.Equal(t, "session", cookie.Name)
	assert.Equal(t, "new-token", cookie.Value)
	assert.Equal(t, 90*24*60*60, cookie.MaxAge, "90 days, so it survives closing the browser")
	assert.True(t, cookie.HttpOnly)
	assert.True(t, cookie.Secure)
	assert.Equal(t, http.SameSiteStrictMode, cookie.SameSite)
	assert.Equal(t, "/", cookie.Path)
}

func TestGoogleLogin_DevCookieIsNotSecure(t *testing.T) {
	t.Setenv("RUN_MODE", "dev")
	c, mock := newController(t, &fakeGoogle{user: &user.GoogleUser{ID: "g-1"}}, &fakeSessions{createToken: "new-token"})
	expectExistingCaregiver(mock, 7)

	w := login(c)

	cookies := w.Result().Cookies()
	require.Len(t, cookies, 1)
	assert.False(t, cookies[0].Secure)
	assert.True(t, cookies[0].HttpOnly)
	assert.Equal(t, http.SameSiteStrictMode, cookies[0].SameSite)
}

func TestGoogleLogin_SendsTheRequestsTokenToGoogle(t *testing.T) {
	google := &fakeGoogle{user: &user.GoogleUser{ID: "g-1"}}
	c, mock := newController(t, google, &fakeSessions{createToken: "new-token"})
	expectExistingCaregiver(mock, 7)

	login(c)

	assert.Equal(t, []string{"id-token"}, google.verified)
}

func TestGoogleLogin_UnreadableBodyIsABadRequestAndSetsNoCookie(t *testing.T) {
	google := &fakeGoogle{user: &user.GoogleUser{ID: "g-1"}}
	sessions := &fakeSessions{}
	c, _ := newController(t, google, sessions)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/google", strings.NewReader(`not json`))
	w := httptest.NewRecorder()

	c.GoogleLogin(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Empty(t, w.Result().Cookies())
	assert.Empty(t, google.verified, "nothing is sent to Google")
	assert.Empty(t, sessions.createdFor)
}

func TestGoogleLogin_RejectedGoogleTokenSetsNoCookie(t *testing.T) {
	sessions := &fakeSessions{}
	c, _ := newController(t, &fakeGoogle{err: errors.New("bad token")}, sessions)

	w := login(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Empty(t, w.Result().Cookies())
	assert.Empty(t, sessions.createdFor)
}

// A failure on our side is a 500, so the caregiver is not told their Google sign-in was wrong
func TestGoogleLogin_SessionErrorIsAServerErrorAndSetsNoCookie(t *testing.T) {
	sessions := &fakeSessions{createErr: errors.New("db down")}
	c, mock := newController(t, &fakeGoogle{user: &user.GoogleUser{ID: "g-1"}}, sessions)
	expectExistingCaregiver(mock, 7)

	w := login(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Empty(t, w.Result().Cookies())
}

func TestGoogleLogin_CaregiverLookupErrorIsAServerError(t *testing.T) {
	sessions := &fakeSessions{}
	c, mock := newController(t, &fakeGoogle{user: &user.GoogleUser{ID: "g-1"}}, sessions)
	mock.ExpectQuery(regexp.QuoteMeta(`FROM "Caregiver"`)).WillReturnError(errors.New("db down"))

	w := login(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Empty(t, w.Result().Cookies())
	assert.Empty(t, sessions.createdFor)
}

/* -------------------------------------------------------
   LOGOUT
------------------------------------------------------- */

func TestLogout_EndsAllSessionsAndClearsTheCookie(t *testing.T) {
	t.Setenv("RUN_MODE", "production")
	sessions := &fakeSessions{foundSession: authsvc.ValidatedSession{UserID: 7}}
	c, _ := newController(t, nil, sessions)

	w := logout(c, sessionCookie("cookie-token"))

	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"status":"logged out"}`, w.Body.String())
	assert.Equal(t, []int{7}, sessions.deletedFor, "every session of that caregiver, not just this device's")
	assertCleared(t, w)
	cleared := w.Result().Cookies()[0]
	assert.True(t, cleared.HttpOnly)
	assert.True(t, cleared.Secure)
	assert.Equal(t, http.SameSiteStrictMode, cleared.SameSite)
}

func TestLogout_WithoutACookieStillSucceeds(t *testing.T) {
	sessions := &fakeSessions{}
	c, _ := newController(t, nil, sessions)

	w := logout(c, nil)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Zero(t, sessions.found)
	assert.Empty(t, sessions.deletedFor)
	assertCleared(t, w)
}

func TestLogout_UnknownOrExpiredCookieStillSucceeds(t *testing.T) {
	for _, findErr := range []error{authsvc.ErrSessionNotFound, authsvc.ErrSessionExpired} {
		sessions := &fakeSessions{findErr: findErr}
		c, _ := newController(t, nil, sessions)

		w := logout(c, sessionCookie("stale-token"))

		assert.Equal(t, http.StatusOK, w.Code, findErr.Error())
		assert.Empty(t, sessions.deletedFor, findErr.Error())
		assertCleared(t, w)
	}
}

func TestLogout_LookupErrorKeepsTheCookieSoTheCaregiverCanRetry(t *testing.T) {
	sessions := &fakeSessions{findErr: errors.New("db down")}
	c, _ := newController(t, nil, sessions)

	w := logout(c, sessionCookie("cookie-token"))

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Empty(t, w.Result().Cookies())
	assert.Empty(t, sessions.deletedFor)
}

func TestLogout_DeleteErrorKeepsTheCookieSoTheCaregiverCanRetry(t *testing.T) {
	sessions := &fakeSessions{foundSession: authsvc.ValidatedSession{UserID: 7}, deleteErr: errors.New("db down")}
	c, _ := newController(t, nil, sessions)

	w := logout(c, sessionCookie("cookie-token"))

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Empty(t, w.Result().Cookies(), "the sessions may still be valid, so do not pretend they ended")
}
