package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// call sends a request through the protection to a handler that answers 204
func call(t *testing.T, method string, headers map[string]string) *httptest.ResponseRecorder {
	t.Helper()
	protect, err := CrossOriginProtection([]string{"http://localhost:5173"})
	require.NoError(t, err)
	handler := protect(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	req := httptest.NewRequest(method, "http://api.example.com/api/auth/email/code", nil)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	return w
}

func TestCrossOriginProtection_AllowsTheServersOwnPages(t *testing.T) {
	assert.Equal(t, http.StatusNoContent, call(t, http.MethodPost, map[string]string{"Sec-Fetch-Site": "same-origin"}).Code)
	assert.Equal(t, http.StatusNoContent, call(t, http.MethodPost, map[string]string{"Origin": "http://api.example.com"}).Code, "an older browser: Origin matches Host")
}

func TestCrossOriginProtection_AllowsTrustedOrigins(t *testing.T) {
	// the caregiver app's development server is another origin on the same site
	w := call(t, http.MethodPost, map[string]string{"Sec-Fetch-Site": "same-site", "Origin": "http://localhost:5173"})
	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestCrossOriginProtection_BlocksOtherSites(t *testing.T) {
	cases := map[string]map[string]string{
		"cross-site":                   {"Sec-Fetch-Site": "cross-site", "Origin": "https://evil.example"},
		"cross-site without an Origin": {"Sec-Fetch-Site": "cross-site"},
		"same-site but not trusted":    {"Sec-Fetch-Site": "same-site", "Origin": "http://localhost:9999"},
		"an older browser":             {"Origin": "https://evil.example"},
	}
	for name, headers := range cases {
		t.Run(name, func(t *testing.T) {
			w := call(t, http.MethodPost, headers)

			assert.Equal(t, http.StatusForbidden, w.Code)
			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
			assert.JSONEq(t, `{"code":"CROSS_ORIGIN_BLOCKED","error":"This request came from another site."}`, w.Body.String())
		})
	}
}

func TestCrossOriginProtection_AllowsRequestsThatAreNotFromABrowserPage(t *testing.T) {
	assert.Equal(t, http.StatusNoContent, call(t, http.MethodPost, nil).Code, "no Origin and no Sec-Fetch-Site: curl, the tablet app, a server")
}

func TestCrossOriginProtection_OnlyChecksUnsafeMethods(t *testing.T) {
	w := call(t, http.MethodGet, map[string]string{"Sec-Fetch-Site": "cross-site", "Origin": "https://evil.example"})
	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestCrossOriginProtection_RejectsAMalformedTrustedOrigin(t *testing.T) {
	for _, origin := range []string{"localhost:5173", "http://localhost:5173/path", "*", ""} {
		protect, err := CrossOriginProtection([]string{origin})
		assert.Error(t, err, "%q", origin)
		assert.Nil(t, protect)
	}
}
