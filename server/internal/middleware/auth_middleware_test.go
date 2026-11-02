package middleware

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"server/internal/services/auth"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

/* -------------------------------------------------------
   SETUP
------------------------------------------------------- */

type fakeSessions struct {
	session auth.ValidatedSession
	err     error

	validated []string
}

func (f *fakeSessions) Create(ctx context.Context, userID int) (string, error) {
	return "", errors.New("not used by the middleware")
}

func (f *fakeSessions) Validate(ctx context.Context, token string) (auth.ValidatedSession, error) {
	f.validated = append(f.validated, token)
	return f.session, f.err
}

func (f *fakeSessions) DeleteAllForUser(ctx context.Context, userID int) error {
	return errors.New("not used by the middleware")
}

// serve runs one request through RequireAuth and reports whether the protected handler ran,
// and who it saw as the signed-in caregiver
func serve(sessions auth.SessionService, cookie *http.Cookie) (w *httptest.ResponseRecorder, reached bool, userID any) {
	protected := RequireAuth(sessions)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reached = true
		userID = r.Context().Value(UserIDKey)
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	if cookie != nil {
		req.AddCookie(cookie)
	}
	w = httptest.NewRecorder()
	protected.ServeHTTP(w, req)
	return w, reached, userID
}

func sessionCookieFor(token string) *http.Cookie {
	return &http.Cookie{Name: SessionCookieName, Value: token}
}

/* -------------------------------------------------------
   REQUIRE AUTH
------------------------------------------------------- */

func TestRequireAuth_NoCookie(t *testing.T) {
	sessions := &fakeSessions{}

	w, reached, _ := serve(sessions, nil)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.False(t, reached)
	assert.Empty(t, sessions.validated, "nothing to look up")
	assert.Empty(t, w.Result().Cookies(), "there is no cookie to delete")
}

func TestRequireAuth_ValidSessionReachesTheHandlerAsThatCaregiver(t *testing.T) {
	sessions := &fakeSessions{session: auth.ValidatedSession{UserID: 7, ExpiresAt: time.Now().Add(30 * 24 * time.Hour)}}

	w, reached, userID := serve(sessions, sessionCookieFor("cookie-token"))

	assert.Equal(t, http.StatusOK, w.Code)
	assert.True(t, reached)
	assert.Equal(t, 7, userID)
	assert.Equal(t, []string{"cookie-token"}, sessions.validated)
}

func TestRequireAuth_SessionThatDidNotRenewSendsNoCookie(t *testing.T) {
	sessions := &fakeSessions{session: auth.ValidatedSession{UserID: 7, ExpiresAt: time.Now().Add(30 * 24 * time.Hour), Refreshed: false}}

	w, _, _ := serve(sessions, sessionCookieFor("cookie-token"))

	assert.Empty(t, w.Result().Cookies())
}

func TestRequireAuth_RenewedSessionSendsTheSameTokenWithTheNewExpiry(t *testing.T) {
	t.Setenv("RUN_MODE", "production")
	// Near the one year cutoff the new expiry is sooner than 90 days, and the cookie must match it
	expires := time.Now().Add(65 * 24 * time.Hour)
	sessions := &fakeSessions{session: auth.ValidatedSession{UserID: 7, ExpiresAt: expires, Refreshed: true}}

	w, reached, _ := serve(sessions, sessionCookieFor("cookie-token"))

	assert.Equal(t, http.StatusOK, w.Code)
	assert.True(t, reached)
	cookies := w.Result().Cookies()
	require.Len(t, cookies, 1)
	c := cookies[0]
	assert.Equal(t, "session", c.Name)
	assert.Equal(t, "cookie-token", c.Value, "the browser keeps the same token, only its lifetime moves")
	assert.InDelta(t, (65 * 24 * time.Hour).Seconds(), c.MaxAge, 5)
	assert.True(t, c.HttpOnly)
	assert.True(t, c.Secure)
	assert.Equal(t, http.SameSiteStrictMode, c.SameSite)
	assert.Equal(t, "/", c.Path)
}

func TestRequireAuth_RenewedSessionHasItsCookieBeforeTheHandlerWrites(t *testing.T) {
	sessions := &fakeSessions{session: auth.ValidatedSession{UserID: 7, ExpiresAt: time.Now().Add(time.Hour * 24), Refreshed: true}}
	protected := RequireAuth(sessions)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("body already started"))
	}))
	req := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	req.AddCookie(sessionCookieFor("cookie-token"))
	w := httptest.NewRecorder()

	protected.ServeHTTP(w, req)

	assert.Len(t, w.Result().Cookies(), 1, "headers cannot be added once the body has started")
}

func TestRequireAuth_UnknownOrExpiredSessionIsDeniedAndItsCookieDeleted(t *testing.T) {
	for _, validateErr := range []error{
		auth.ErrSessionNotFound,
		auth.ErrSessionExpired,
		fmt.Errorf("wrapped: %w", auth.ErrSessionExpired),
	} {
		t.Run(validateErr.Error(), func(t *testing.T) {
			t.Setenv("RUN_MODE", "production")
			sessions := &fakeSessions{err: validateErr}

			w, reached, _ := serve(sessions, sessionCookieFor("stale-token"))

			assert.Equal(t, http.StatusUnauthorized, w.Code)
			assert.False(t, reached)
			cookies := w.Result().Cookies()
			require.Len(t, cookies, 1)
			assert.Equal(t, "session", cookies[0].Name)
			assert.Equal(t, "", cookies[0].Value)
			assert.Less(t, cookies[0].MaxAge, 0)
			assert.Equal(t, "/", cookies[0].Path)
		})
	}
}

func TestRequireAuth_DatabaseErrorIsNotALogout(t *testing.T) {
	sessions := &fakeSessions{err: errors.New("db down")}

	w, reached, _ := serve(sessions, sessionCookieFor("cookie-token"))

	assert.Equal(t, http.StatusInternalServerError, w.Code, "not 401, which the app reads as logged out")
	assert.False(t, reached)
	assert.Empty(t, w.Result().Cookies(), "the cookie is still good once the database is back")
}
