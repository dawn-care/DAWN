package auth

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	authsvc "server/internal/services/auth"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

/* -------------------------------------------------------
   SETUP
------------------------------------------------------- */

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

// newEmailController builds a controller with the fake code service; nil codes means email
// sign-in is not set up on the server
func newEmailController(t *testing.T, codes *fakeCodes, sessions *fakeSessions) (*AuthController, sqlmock.Sqlmock) {
	t.Helper()
	c, mock := newController(t, &fakeGoogle{}, sessions)
	if codes != nil {
		c.AuthService.WithLoginCodes(codes)
	}
	return c, mock
}

func post(handler http.HandlerFunc, remoteAddr, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	if remoteAddr != "" {
		req.RemoteAddr = remoteAddr
	}
	w := httptest.NewRecorder()
	handler(w, req)
	return w
}

func requireError(t *testing.T, w *httptest.ResponseRecorder, status int, code string) {
	t.Helper()
	assert.Equal(t, status, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
	var body AuthError
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, code, body.Code)
	assert.NotEmpty(t, body.Error, "a message the caregiver can read")
}

/* -------------------------------------------------------
   POST /api/auth/email/code
------------------------------------------------------- */

func TestRequestEmailCode_SendsACodeForTheAddressAndClientIP(t *testing.T) {
	codes := &fakeCodes{}
	c, _ := newEmailController(t, codes, &fakeSessions{})

	w := post(c.RequestEmailCode, "203.0.113.9:51234", `{"email":"pat@example.com"}`)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"status":"sent"}`, w.Body.String())
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
	assert.Equal(t, []string{"pat@example.com|203.0.113.9"}, codes.requested, "the IP without its port")
	assert.Empty(t, w.Result().Cookies(), "asking for a code signs nobody in")
}

func TestRequestEmailCode_UsesTheConnectionsIPNotForwardingHeaders(t *testing.T) {
	codes := &fakeCodes{}
	c, _ := newEmailController(t, codes, &fakeSessions{})
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"email":"pat@example.com"}`))
	req.RemoteAddr = "203.0.113.9:51234"
	req.Header.Set("X-Forwarded-For", "198.51.100.77")
	req.Header.Set("X-Real-IP", "198.51.100.78")

	c.RequestEmailCode(httptest.NewRecorder(), req)

	assert.Equal(t, []string{"pat@example.com|203.0.113.9"}, codes.requested, "a client cannot pick its own IP to get around the per-IP limit")
}

func TestClientIP(t *testing.T) {
	for remoteAddr, want := range map[string]string{
		"203.0.113.9:51234":   "203.0.113.9",
		"[2001:db8::1]:51234": "2001:db8::1",
		"203.0.113.9":         "203.0.113.9",
		"":                    "",
	} {
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		req.RemoteAddr = remoteAddr
		assert.Equal(t, want, clientIP(req), remoteAddr)
	}
}

func TestRequestEmailCode_Errors(t *testing.T) {
	cases := map[string]struct {
		err    error
		status int
		code   string
	}{
		"invalid address": {authsvc.ErrInvalidEmail, http.StatusBadRequest, CodeInvalidEmail},
		"rate limited":    {authsvc.ErrTooManyRequests, http.StatusTooManyRequests, CodeRateLimited},
		"another failure": {errors.New("smtp: password rejected for hunter2"), http.StatusInternalServerError, CodeInternal},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			c, _ := newEmailController(t, &fakeCodes{requestErr: tc.err}, &fakeSessions{})

			w := post(c.RequestEmailCode, "203.0.113.9:1", `{"email":"pat@example.com"}`)

			requireError(t, w, tc.status, tc.code)
			assert.NotContains(t, w.Body.String(), "hunter2", "internal error text is not sent to the browser")
		})
	}
}

func TestRequestEmailCode_UnavailableWhenNotSetUp(t *testing.T) {
	c, _ := newEmailController(t, nil, &fakeSessions{})

	w := post(c.RequestEmailCode, "203.0.113.9:1", `{"email":"pat@example.com"}`)

	requireError(t, w, http.StatusServiceUnavailable, CodeEmailLoginUnavailable)
}

func TestEmailLoginEndpoints_RejectUnreadableBodies(t *testing.T) {
	bodies := map[string]string{
		"not json":    `email=pat@example.com`,
		"broken json": `{"email":`,
		"too large":   `{"email":"` + strings.Repeat("a", 5000) + `"}`,
	}
	for name, body := range bodies {
		t.Run(name, func(t *testing.T) {
			codes := &fakeCodes{}
			c, _ := newEmailController(t, codes, &fakeSessions{})

			requireError(t, post(c.RequestEmailCode, "203.0.113.9:1", body), http.StatusBadRequest, CodeInvalidRequest)
			requireError(t, post(c.VerifyEmailCode, "203.0.113.9:1", body), http.StatusBadRequest, CodeInvalidRequest)
			assert.Empty(t, codes.requested)
			assert.Empty(t, codes.verified)
		})
	}
}

/* -------------------------------------------------------
   POST /api/auth/email/verify
------------------------------------------------------- */

func TestVerifyEmailCode_SignsInWithTheSameCookieAsGoogle(t *testing.T) {
	t.Setenv("RUN_MODE", "production")
	codes := &fakeCodes{}
	sessions := &fakeSessions{createToken: "new-token"}
	c, mock := newEmailController(t, codes, sessions)
	expectExistingCaregiver(mock, 7)

	w := post(c.VerifyEmailCode, "203.0.113.9:1", `{"email":"pat@example.com","code":"123456"}`)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"status":"authenticated"}`, w.Body.String())
	assert.Equal(t, []string{"pat@example.com|123456"}, codes.verified)
	assert.Equal(t, []int{7}, sessions.createdFor)
	cookies := w.Result().Cookies()
	require.Len(t, cookies, 1)
	cookie := cookies[0]
	assert.Equal(t, "session", cookie.Name)
	assert.Equal(t, "new-token", cookie.Value)
	assert.Equal(t, 90*24*60*60, cookie.MaxAge)
	assert.True(t, cookie.HttpOnly)
	assert.True(t, cookie.Secure)
	assert.Equal(t, http.SameSiteStrictMode, cookie.SameSite)
	assert.Equal(t, "/", cookie.Path)
}

func TestVerifyEmailCode_Errors(t *testing.T) {
	cases := map[string]struct {
		err    error
		status int
		code   string
	}{
		"wrong or expired code": {authsvc.ErrCodeRejected, http.StatusUnauthorized, CodeCodeRejected},
		"too many attempts":     {authsvc.ErrTooManyAttempts, http.StatusTooManyRequests, CodeTooManyAttempts},
		"another failure":       {errors.New("db: password authentication failed"), http.StatusInternalServerError, CodeInternal},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			sessions := &fakeSessions{createToken: "new-token"}
			c, mock := newEmailController(t, &fakeCodes{verifyErr: tc.err}, sessions)

			w := post(c.VerifyEmailCode, "203.0.113.9:1", `{"email":"pat@example.com","code":"000000"}`)

			requireError(t, w, tc.status, tc.code)
			assert.Empty(t, w.Result().Cookies(), "no cookie without a right code")
			assert.Empty(t, sessions.createdFor)
			assert.NoError(t, mock.ExpectationsWereMet(), "the caregiver table was not touched")
			assert.NotContains(t, w.Body.String(), "password")
		})
	}
}

func TestVerifyEmailCode_UnavailableWhenNotSetUp(t *testing.T) {
	c, _ := newEmailController(t, nil, &fakeSessions{})

	w := post(c.VerifyEmailCode, "203.0.113.9:1", `{"email":"pat@example.com","code":"123456"}`)

	requireError(t, w, http.StatusServiceUnavailable, CodeEmailLoginUnavailable)
	assert.Empty(t, w.Result().Cookies())
}

func TestVerifyEmailCode_NoCookieWhenTheSessionCannotBeStarted(t *testing.T) {
	sessions := &fakeSessions{createErr: errors.New("db down")}
	c, mock := newEmailController(t, &fakeCodes{}, sessions)
	expectExistingCaregiver(mock, 7)

	w := post(c.VerifyEmailCode, "203.0.113.9:1", `{"email":"pat@example.com","code":"123456"}`)

	requireError(t, w, http.StatusInternalServerError, CodeInternal)
	assert.Empty(t, w.Result().Cookies())
}
