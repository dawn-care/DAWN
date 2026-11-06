package email

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"io"
	"log"
	"mime"
	"mime/multipart"
	"mime/quotedprintable"
	"net"
	"net/mail"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testCode = "482913"

func sendTestCode(t *testing.T, m Mailer, to string) error {
	t.Helper()
	text, html := LoginCodeMessage(testCode, 10*time.Minute)
	return m.Send(context.Background(), to, LoginCodeSubject, text, html)
}

// readParts parses a delivered message and returns its headers and its parts by content type
func readParts(t *testing.T, raw string) (mail.Header, map[string]string) {
	t.Helper()
	msg, err := mail.ReadMessage(strings.NewReader(raw))
	require.NoError(t, err)
	mediaType, params, err := mime.ParseMediaType(msg.Header.Get("Content-Type"))
	require.NoError(t, err)
	require.Equal(t, "multipart/alternative", mediaType)

	parts := map[string]string{}
	reader := multipart.NewReader(msg.Body, params["boundary"])
	for {
		part, err := reader.NextPart()
		if errors.Is(err, io.EOF) {
			break
		}
		require.NoError(t, err)
		body, err := io.ReadAll(quotedprintable.NewReader(part))
		require.NoError(t, err)
		contentType, _, err := mime.ParseMediaType(part.Header.Get("Content-Type"))
		require.NoError(t, err)
		parts[contentType] = string(body)
	}
	return msg.Header, parts
}

/* -------------------------------------------------------
   SMTPMailer
------------------------------------------------------- */

func TestSMTPMailer_StartTLSWithAuthDeliversTheCodeEmail(t *testing.T) {
	srv := newFakeSMTP(t, func(s *fakeSMTP) { s.startTLS = true; s.user = "apiuser"; s.pass = "secret" })
	m := mailerFor(t, srv, func(c *Config) { c.User = "apiuser"; c.Password = "secret" })

	require.NoError(t, sendTestCode(t, m, "pat@example.com"))

	got := srv.delivered()
	require.Len(t, got, 1)
	assert.True(t, got[0].tls, "the message travelled over TLS")
	assert.True(t, got[0].authed)
	assert.Equal(t, "no-reply@example.com", got[0].from)
	assert.Equal(t, []string{"pat@example.com"}, got[0].to)

	header, parts := readParts(t, got[0].data)
	assert.Equal(t, "Your D.A.W.N. sign-in code", header.Get("Subject"))
	assert.Equal(t, `"D.A.W.N." <no-reply@example.com>`, header.Get("From"))
	assert.Equal(t, "<pat@example.com>", header.Get("To"))
	assert.Equal(t, "1.0", header.Get("Mime-Version"))
	assert.Contains(t, header.Get("Message-Id"), "@example.com>")
	_, err := mail.ParseDate(header.Get("Date"))
	assert.NoError(t, err)

	require.Contains(t, parts, "text/plain")
	require.Contains(t, parts, "text/html")
	for kind, body := range parts {
		assert.Contains(t, body, testCode, kind)
		assert.Contains(t, body, "10 minutes", kind)
	}
}

func TestSMTPMailer_ImplicitTLS(t *testing.T) {
	srv := newFakeSMTP(t, func(s *fakeSMTP) { s.implicit = true })
	m := mailerFor(t, srv, func(c *Config) { c.ImplicitTLS = true })

	require.NoError(t, sendTestCode(t, m, "pat@example.com"))

	got := srv.delivered()
	require.Len(t, got, 1)
	assert.True(t, got[0].tls)
}

func TestSMTPMailer_RefusesAServerWithoutTLSAndSendsNothing(t *testing.T) {
	srv := newFakeSMTP(t, func(s *fakeSMTP) { s.startTLS = false; s.user = "apiuser"; s.pass = "secret" })
	m := mailerFor(t, srv, func(c *Config) { c.User = "apiuser"; c.Password = "secret" })

	err := sendTestCode(t, m, "pat@example.com")

	assert.ErrorIs(t, err, ErrInsecureServer)
	assert.Empty(t, srv.delivered())
	_, authAttempts := srv.counts()
	assert.Zero(t, authAttempts, "the password is never sent over a connection with no TLS")
}

func TestSMTPMailer_DevelopmentMayUseAServerWithoutTLS(t *testing.T) {
	srv := newFakeSMTP(t, nil) // like Mailpit: no TLS, no login
	m := mailerFor(t, srv, func(c *Config) { c.AllowInsecure = true })

	require.NoError(t, sendTestCode(t, m, "pat@example.com"))

	got := srv.delivered()
	require.Len(t, got, 1)
	assert.False(t, got[0].tls)
}

func TestSMTPMailer_DevelopmentStillUsesTLSWhenTheServerOffersIt(t *testing.T) {
	srv := newFakeSMTP(t, func(s *fakeSMTP) { s.startTLS = true })
	m := mailerFor(t, srv, func(c *Config) { c.AllowInsecure = true })

	require.NoError(t, sendTestCode(t, m, "pat@example.com"))

	got := srv.delivered()
	require.Len(t, got, 1)
	assert.True(t, got[0].tls)
}

func TestSMTPMailer_DoesNotTrustAnUnknownCertificate(t *testing.T) {
	srv := newFakeSMTP(t, func(s *fakeSMTP) { s.startTLS = true })
	m := mailerFor(t, srv, nil)
	m.tlsConfig = &tls.Config{} // no longer trusts the test certificate

	err := sendTestCode(t, m, "pat@example.com")

	assert.Error(t, err)
	assert.Empty(t, srv.delivered())
}

func TestSMTPMailer_WrongPassword(t *testing.T) {
	srv := newFakeSMTP(t, func(s *fakeSMTP) { s.startTLS = true; s.user = "apiuser"; s.pass = "secret" })
	m := mailerFor(t, srv, func(c *Config) { c.User = "apiuser"; c.Password = "wrong" })

	err := sendTestCode(t, m, "pat@example.com")

	assert.Error(t, err)
	assert.NotContains(t, err.Error(), "wrong", "the password is not put in the error")
	assert.Empty(t, srv.delivered())
}

func TestSMTPMailer_RejectedRecipient(t *testing.T) {
	srv := newFakeSMTP(t, func(s *fakeSMTP) { s.startTLS = true; s.rejectRcpt = true })
	m := mailerFor(t, srv, nil)

	err := sendTestCode(t, m, "nobody@example.com")

	assert.ErrorContains(t, err, "recipient rejected")
	assert.Empty(t, srv.delivered())
}

func TestSMTPMailer_InvalidRecipientsNeverReachTheServer(t *testing.T) {
	srv := newFakeSMTP(t, func(s *fakeSMTP) { s.startTLS = true })
	m := mailerFor(t, srv, nil)

	for _, to := range []string{
		"",
		"   ",
		"not-an-email",
		"pat@example.com\r\nBcc: boss@example.com",
		"pat@example.com\nBcc: boss@example.com",
		"Pat <pat@example.com>",
		"pat@example.com, other@example.com",
	} {
		assert.Error(t, sendTestCode(t, m, to), "%q", to)
	}
	connections, _ := srv.counts()
	assert.Zero(t, connections, "no connection was opened")
}

func TestSMTPMailer_SubjectWithALineBreakIsRefused(t *testing.T) {
	srv := newFakeSMTP(t, func(s *fakeSMTP) { s.startTLS = true })
	m := mailerFor(t, srv, nil)

	err := m.Send(context.Background(), "pat@example.com", "Hi\r\nBcc: boss@example.com", "text", "<p>html</p>")

	assert.Error(t, err)
	connections, _ := srv.counts()
	assert.Zero(t, connections)
}

func TestSMTPMailer_ServerThatNeverAnswersTimesOut(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() { ln.Close() })
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			t.Cleanup(func() { conn.Close() })
		}
	}()
	m, err := NewSMTPMailer(Config{Host: "127.0.0.1", Port: ln.Addr().(*net.TCPAddr).Port, From: "no-reply@example.com"})
	require.NoError(t, err)
	m.timeout = 300 * time.Millisecond

	start := time.Now()
	err = sendTestCode(t, m, "pat@example.com")

	assert.Error(t, err)
	assert.Less(t, time.Since(start), 5*time.Second)
}

func TestSMTPMailer_CancelledContextStopsTheSend(t *testing.T) {
	srv := newFakeSMTP(t, func(s *fakeSMTP) { s.startTLS = true })
	m := mailerFor(t, srv, nil)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := m.Send(ctx, "pat@example.com", LoginCodeSubject, "text", "<p>html</p>")

	assert.Error(t, err)
	assert.Empty(t, srv.delivered())
}

func TestSMTPMailer_ServerThatIsDown(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	port := ln.Addr().(*net.TCPAddr).Port
	ln.Close()
	m, err := NewSMTPMailer(Config{Host: "127.0.0.1", Port: port, From: "no-reply@example.com"})
	require.NoError(t, err)

	assert.Error(t, sendTestCode(t, m, "pat@example.com"))
}

func TestSMTPMailer_NonASCIISubjectIsEncoded(t *testing.T) {
	srv := newFakeSMTP(t, func(s *fakeSMTP) { s.startTLS = true })
	m := mailerFor(t, srv, nil)

	require.NoError(t, m.Send(context.Background(), "pat@example.com", "Votre code café", "texte", "<p>texte</p>"))

	got := srv.delivered()
	require.Len(t, got, 1)
	assert.NotContains(t, got[0].data, "café", "raw non-ASCII text is not put in a header")
	header, _ := readParts(t, got[0].data)
	decoded, err := new(mime.WordDecoder).DecodeHeader(header.Get("Subject"))
	require.NoError(t, err)
	assert.Equal(t, "Votre code café", decoded)
}

func TestNewSMTPMailer_Validates(t *testing.T) {
	_, err := NewSMTPMailer(Config{From: "no-reply@example.com"})
	assert.Error(t, err, "a host is required")

	_, err = NewSMTPMailer(Config{Host: "smtp.example.com", From: ""})
	assert.Error(t, err, "a sender is required")

	_, err = NewSMTPMailer(Config{Host: "smtp.example.com", From: "not a sender"})
	assert.Error(t, err, "the sender must be an address")

	m, err := NewSMTPMailer(Config{Host: "smtp.example.com", From: "D.A.W.N. <no-reply@example.com>"})
	require.NoError(t, err)
	assert.Equal(t, 587, m.cfg.Port, "the port defaults to 587")
}

/* -------------------------------------------------------
   NewMailerFromEnv
------------------------------------------------------- */

func env(values map[string]string) func(string) string {
	return func(key string) string { return values[key] }
}

func TestNewMailerFromEnv_NoHost(t *testing.T) {
	m, err := NewMailerFromEnv(env(nil), true)
	require.NoError(t, err)
	assert.IsType(t, &LogMailer{}, m, "development logs the email")

	m, err = NewMailerFromEnv(env(nil), false)
	require.NoError(t, err)
	assert.Nil(t, m, "production without SMTP sends nothing, so sign-in by email can be turned off")
}

func TestNewMailerFromEnv_BuildsAnSMTPMailer(t *testing.T) {
	values := map[string]string{
		"SMTP_HOST": " smtp.example.com ",
		"SMTP_USER": "apiuser",
		"SMTP_PASS": "secret",
		"SMTP_FROM": "D.A.W.N. <no-reply@example.com>",
	}

	m, err := NewMailerFromEnv(env(values), false)
	require.NoError(t, err)
	smtpMailer, ok := m.(*SMTPMailer)
	require.True(t, ok)
	assert.Equal(t, Config{
		Host: "smtp.example.com", Port: 587, User: "apiuser", Password: "secret",
		From: "D.A.W.N. <no-reply@example.com>", ImplicitTLS: false, AllowInsecure: false,
	}, smtpMailer.cfg, "production never allows a connection without TLS")

	m, err = NewMailerFromEnv(env(values), true)
	require.NoError(t, err)
	assert.True(t, m.(*SMTPMailer).cfg.AllowInsecure, "development may reach a test inbox without TLS")
}

func TestNewMailerFromEnv_Port465UsesImplicitTLS(t *testing.T) {
	m, err := NewMailerFromEnv(env(map[string]string{
		"SMTP_HOST": "smtp.example.com", "SMTP_PORT": "465", "SMTP_FROM": "no-reply@example.com",
	}), false)
	require.NoError(t, err)
	cfg := m.(*SMTPMailer).cfg
	assert.Equal(t, 465, cfg.Port)
	assert.True(t, cfg.ImplicitTLS)
}

func TestNewMailerFromEnv_RejectsAnInvalidConfiguration(t *testing.T) {
	base := func(extra map[string]string) map[string]string {
		values := map[string]string{"SMTP_HOST": "smtp.example.com", "SMTP_FROM": "no-reply@example.com"}
		for k, v := range extra {
			values[k] = v
		}
		return values
	}
	cases := map[string]map[string]string{
		"port is not a number":     base(map[string]string{"SMTP_PORT": "abc"}),
		"port is zero":             base(map[string]string{"SMTP_PORT": "0"}),
		"port is too large":        base(map[string]string{"SMTP_PORT": "70000"}),
		"user without a password":  base(map[string]string{"SMTP_USER": "apiuser"}),
		"password without a user":  base(map[string]string{"SMTP_PASS": "secret"}),
		"no sender":                {"SMTP_HOST": "smtp.example.com"},
		"sender is not an address": base(map[string]string{"SMTP_FROM": "not a sender"}),
	}
	for name, values := range cases {
		t.Run(name, func(t *testing.T) {
			m, err := NewMailerFromEnv(env(values), false)
			assert.Error(t, err)
			assert.True(t, m == nil, "a plain nil, so the caller's == nil check works")
		})
	}
}

/* -------------------------------------------------------
   LogMailer
------------------------------------------------------- */

func TestLogMailer_PrintsTheEmail(t *testing.T) {
	var out bytes.Buffer
	m := NewLogMailer(log.New(&out, "", 0))

	require.NoError(t, sendTestCode(t, m, "pat@example.com"))

	assert.Contains(t, out.String(), "pat@example.com")
	assert.Contains(t, out.String(), LoginCodeSubject)
	assert.Contains(t, out.String(), testCode, "development reads the code from the log")
}

func TestLogMailer_RejectsAnInvalidRecipient(t *testing.T) {
	var out bytes.Buffer
	m := NewLogMailer(log.New(&out, "", 0))

	assert.Error(t, sendTestCode(t, m, "not-an-email"))
	assert.Empty(t, out.String())
}

/* -------------------------------------------------------
   Sign-in code email
------------------------------------------------------- */

func TestLoginCodeMessage(t *testing.T) {
	assert.Equal(t, "Your D.A.W.N. sign-in code", LoginCodeSubject)

	text, html := LoginCodeMessage("123456", 10*time.Minute)
	for name, body := range map[string]string{"text": text, "html": html} {
		assert.Contains(t, body, "123456", name)
		assert.Contains(t, body, "expires in 10 minutes", name)
		assert.Contains(t, body, "ignore this email", name)
	}
	assert.Contains(t, html, "<html>")
	assert.NotContains(t, text, "<", "the plain text has no markup")
}

func TestLoginCodeMessage_OneMinute(t *testing.T) {
	text, _ := LoginCodeMessage("123456", time.Minute)
	assert.Contains(t, text, "expires in 1 minute.")
}

func TestLoginCodeMessage_EscapesTheCodeInHTML(t *testing.T) {
	_, html := LoginCodeMessage("<b>1</b>", 10*time.Minute)
	assert.NotContains(t, html, "<b>1</b>")
	assert.Contains(t, html, "&lt;b&gt;1&lt;/b&gt;")
}
