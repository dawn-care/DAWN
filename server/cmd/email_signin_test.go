package main

import (
	"testing"

	authSvc "server/internal/services/auth"

	"github.com/stretchr/testify/assert"
)

const goodSecret = "0123456789abcdef0123456789abcdef"

func TestEnableEmailSignIn(t *testing.T) {
	cases := []struct {
		name string
		env  map[string]string
		on   bool
	}{
		{"production with a mail server and a secret", map[string]string{"SMTP_HOST": "smtp.example.com", "SMTP_FROM": "no-reply@example.com", "LOGIN_CODE_SECRET": goodSecret}, true},
		{"development with no mail server logs the email instead", map[string]string{"RUN_MODE": "dev", "LOGIN_CODE_SECRET": goodSecret}, true},
		{"production with no mail server", map[string]string{"LOGIN_CODE_SECRET": goodSecret}, false},
		{"no secret", map[string]string{"SMTP_HOST": "smtp.example.com", "SMTP_FROM": "no-reply@example.com"}, false},
		{"secret too short", map[string]string{"SMTP_HOST": "smtp.example.com", "SMTP_FROM": "no-reply@example.com", "LOGIN_CODE_SECRET": "short"}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			for _, key := range []string{"RUN_MODE", "SMTP_HOST", "SMTP_PORT", "SMTP_USER", "SMTP_PASS", "SMTP_FROM", "LOGIN_CODE_SECRET"} {
				t.Setenv(key, tc.env[key])
			}
			auth := authSvc.NewAuthService(nil, nil, nil, nil)

			enableEmailSignIn(auth, nil)

			assert.Equal(t, tc.on, auth.LoginCodes != nil)
		})
	}
}
