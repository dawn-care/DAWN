package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGoogleUserFromClaims(t *testing.T) {
	cases := []struct {
		name         string
		claims       map[string]any
		wantEmail    string
		wantName     string
		wantVerified bool
	}{
		{"all claims", map[string]any{"email": "pat@example.com", "name": "Pat", "email_verified": true}, "pat@example.com", "Pat", true},
		{"unverified email", map[string]any{"email": "pat@example.com", "name": "Pat", "email_verified": false}, "pat@example.com", "Pat", false},
		{"verified as text", map[string]any{"email": "pat@example.com", "email_verified": "true"}, "pat@example.com", "", true},
		{"no email_verified claim", map[string]any{"email": "pat@example.com", "name": "Pat"}, "pat@example.com", "Pat", false},
		{"no email claim", map[string]any{"name": "Pat", "email_verified": true}, "", "Pat", false},
		{"claims of the wrong type", map[string]any{"email": 5, "name": true, "email_verified": 1}, "", "", false},
		{"nil claims", nil, "", "", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			u := googleUserFromClaims("sub-1", tc.claims)
			assert.Equal(t, "sub-1", u.ID)
			assert.Equal(t, tc.wantEmail, u.Email)
			assert.Equal(t, tc.wantName, u.Name)
			assert.Equal(t, tc.wantVerified, u.EmailVerified)
		})
	}
}
