package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const goodSecret = "0123456789abcdef0123456789abcdef"

func TestNewLoginCodes(t *testing.T) {
	cases := []struct {
		name string
		env  map[string]string
		ok   bool
	}{
		{"production with a mail server and a secret", map[string]string{"SMTP_HOST": "smtp.example.com", "SMTP_FROM": "no-reply@example.com", "LOGIN_CODE_SECRET": goodSecret}, true},
		{"development with no mail server logs the email instead", map[string]string{"RUN_MODE": "dev", "LOGIN_CODE_SECRET": goodSecret}, true},
		{"development with no secret makes one for the run", map[string]string{"RUN_MODE": "dev"}, true},
		{"production with no mail server does not start", map[string]string{"LOGIN_CODE_SECRET": goodSecret}, false},
		{"production with no secret does not start", map[string]string{"SMTP_HOST": "smtp.example.com", "SMTP_FROM": "no-reply@example.com"}, false},
		{"a secret that is too short does not start", map[string]string{"SMTP_HOST": "smtp.example.com", "SMTP_FROM": "no-reply@example.com", "LOGIN_CODE_SECRET": "short"}, false},
		{"development with a short secret does not start either", map[string]string{"RUN_MODE": "dev", "LOGIN_CODE_SECRET": "short"}, false},
		{"an invalid mail setting does not start", map[string]string{"SMTP_HOST": "smtp.example.com", "SMTP_PORT": "nope", "SMTP_FROM": "no-reply@example.com", "LOGIN_CODE_SECRET": goodSecret}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			getenv := func(key string) string { return tc.env[key] }

			codes, err := newLoginCodes(getenv, nil)

			if tc.ok {
				assert.NoError(t, err)
				assert.NotNil(t, codes)
			} else {
				assert.Error(t, err)
				assert.Nil(t, codes)
			}
		})
	}
}
