package auth

import (
	"testing"

	"server/internal/services/user"

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
		{"workspace domain", map[string]any{"email": "pat@stanns.org", "email_verified": true, "hd": "stanns.org"}, "pat@stanns.org", "", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			u := googleUserFromClaims("sub-1", tc.claims)
			assert.Equal(t, "sub-1", u.ID)
			assert.Equal(t, tc.wantEmail, u.Email)
			assert.Equal(t, tc.wantName, u.Name)
			assert.Equal(t, tc.wantVerified, u.EmailVerified)
			assert.Equal(t, tc.claims["hd"] != nil, u.HostedDomain != "")
		})
	}
}

func TestTrustedEmail(t *testing.T) {
	cases := []struct {
		name string
		user user.GoogleUser
		want string
	}{
		{"verified Gmail", user.GoogleUser{Email: "Pat@Gmail.com", EmailVerified: true}, "Pat@Gmail.com"},
		{"unverified Gmail", user.GoogleUser{Email: "pat@gmail.com"}, ""},
		{"verified Workspace address", user.GoogleUser{Email: "pat@stanns.org", EmailVerified: true, HostedDomain: "stanns.org"}, "pat@stanns.org"},
		{"verified outside address on a personal account", user.GoogleUser{Email: "bob@corp.org", EmailVerified: true}, ""},
		{"a look-alike Gmail domain", user.GoogleUser{Email: "pat@notgmail.com", EmailVerified: true}, ""},
		{"no email", user.GoogleUser{EmailVerified: true, HostedDomain: "stanns.org"}, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, tc.user.TrustedEmail())
		})
	}
}
