package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func onlyCookie(t *testing.T, w *httptest.ResponseRecorder) *http.Cookie {
	t.Helper()
	cookies := w.Result().Cookies()
	require.Len(t, cookies, 1)
	return cookies[0]
}

func TestSetSessionCookie_Production(t *testing.T) {
	t.Setenv("RUN_MODE", "production")
	w := httptest.NewRecorder()

	SetSessionCookie(w, "the-token", 90*24*time.Hour)

	c := onlyCookie(t, w)
	assert.Equal(t, "session", c.Name)
	assert.Equal(t, "the-token", c.Value)
	assert.Equal(t, "/", c.Path)
	assert.Equal(t, 90*24*60*60, c.MaxAge, "a real expiry, so it survives closing the browser")
	assert.True(t, c.HttpOnly)
	assert.True(t, c.Secure)
	assert.Equal(t, http.SameSiteStrictMode, c.SameSite)
}

func TestSetSessionCookie_UnsetRunModeIsStillSecure(t *testing.T) {
	t.Setenv("RUN_MODE", "")
	w := httptest.NewRecorder()

	SetSessionCookie(w, "the-token", time.Hour)

	assert.True(t, onlyCookie(t, w).Secure, "only an explicit dev mode drops Secure")
}

func TestSetSessionCookie_DevIsNotSecure(t *testing.T) {
	t.Setenv("RUN_MODE", "dev")
	w := httptest.NewRecorder()

	SetSessionCookie(w, "the-token", time.Hour)

	c := onlyCookie(t, w)
	assert.False(t, c.Secure, "plain HTTP on localhost would never send a Secure cookie")
	assert.True(t, c.HttpOnly)
	assert.Equal(t, http.SameSiteStrictMode, c.SameSite)
}

func TestSetSessionCookie_RoundsDownToWholeSeconds(t *testing.T) {
	w := httptest.NewRecorder()

	SetSessionCookie(w, "the-token", 90*time.Second+900*time.Millisecond)

	assert.Equal(t, 90, onlyCookie(t, w).MaxAge)
}

func TestSetSessionCookie_NothingLeftClearsInsteadOfBecomingABrowserSessionCookie(t *testing.T) {
	for _, left := range []time.Duration{0, -time.Hour, 500 * time.Millisecond} {
		w := httptest.NewRecorder()

		SetSessionCookie(w, "the-token", left)

		c := onlyCookie(t, w)
		assert.Equal(t, "", c.Value, "left=%v", left)
		assert.Less(t, c.MaxAge, 0, "left=%v: Max-Age 0 means delete", left)
	}
}

func TestClearSessionCookie(t *testing.T) {
	t.Setenv("RUN_MODE", "production")
	w := httptest.NewRecorder()

	ClearSessionCookie(w)

	c := onlyCookie(t, w)
	assert.Equal(t, "session", c.Name)
	assert.Equal(t, "", c.Value)
	assert.Equal(t, "/", c.Path, "the path has to match for the browser to delete it")
	assert.Less(t, c.MaxAge, 0)
	assert.True(t, c.HttpOnly)
	assert.True(t, c.Secure)
	assert.Equal(t, http.SameSiteStrictMode, c.SameSite)
}
