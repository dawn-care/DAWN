package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// CrossOriginProtection blocks requests that a page on another site makes through the
// browser. The session cookie is SameSite=Strict, but a sign-in request has no cookie yet, so
// the cookie cannot stop a page on another site from signing a browser in to the attacker's
// account ("login CSRF"). This checks the Sec-Fetch-Site and Origin headers the browser sets.
//
// A request from the server's own origin passes, and so does one from trustedOrigins (the
// caregiver app's development server, for example). A request with neither header is not from
// a browser page, so it passes: it has no cookie to misuse. Only unsafe methods (POST, PUT,
// PATCH, DELETE) are checked.
func CrossOriginProtection(trustedOrigins []string) (func(http.Handler) http.Handler, error) {
	protection := http.NewCrossOriginProtection()
	for _, origin := range trustedOrigins {
		if err := protection.AddTrustedOrigin(origin); err != nil {
			return nil, fmt.Errorf("trusted origin %q: %w", origin, err)
		}
	}
	protection.SetDenyHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{
			"code":  "CROSS_ORIGIN_BLOCKED",
			"error": "This request came from another site.",
		})
	}))
	return protection.Handler, nil
}
