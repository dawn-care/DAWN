// Package email sends the emails the server needs, today only the sign-in code.
package email

import (
	"context"
	"fmt"
	"log"
	"net/mail"
	"strconv"
	"strings"
)

// Mailer sends one email with a plain text and an HTML version of the same message
type Mailer interface {
	Send(ctx context.Context, to, subject, text, html string) error
}

// Config says how to reach the mail server. The provider is chosen only by these values,
// so switching providers changes the environment, never the code.
type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	// From is the sender, for example `D.A.W.N. <no-reply@example.com>`
	From string
	// ImplicitTLS connects over TLS from the first byte (port 465). Otherwise the connection
	// starts as plain text and is upgraded with STARTTLS (port 587) before anything is sent.
	ImplicitTLS bool
	// AllowInsecure lets the mailer send over a connection with no TLS when the server does not
	// offer STARTTLS. Only local development uses it, to reach a test inbox such as Mailpit.
	AllowInsecure bool
}

// NewMailerFromEnv builds the mailer from SMTP_HOST, SMTP_PORT (default 587), SMTP_USER,
// SMTP_PASS and SMTP_FROM.
//
// With no SMTP_HOST, development gets a LogMailer that prints the email to the server log, and
// anything else gets a nil Mailer and no error: sending is not set up, and the caller turns
// email sign-in off. A set but invalid configuration is an error.
func NewMailerFromEnv(getenv func(string) string, devMode bool) (Mailer, error) {
	host := strings.TrimSpace(getenv("SMTP_HOST"))
	if host == "" {
		if devMode {
			return NewLogMailer(log.Default()), nil
		}
		return nil, nil
	}

	port := 587
	if raw := strings.TrimSpace(getenv("SMTP_PORT")); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed < 1 || parsed > 65535 {
			return nil, fmt.Errorf("email: SMTP_PORT %q is not a port number", raw)
		}
		port = parsed
	}

	user := strings.TrimSpace(getenv("SMTP_USER"))
	password := getenv("SMTP_PASS")
	if (user == "") != (password == "") {
		return nil, fmt.Errorf("email: set both SMTP_USER and SMTP_PASS, or neither")
	}

	return NewSMTPMailer(Config{
		Host:          host,
		Port:          port,
		User:          user,
		Password:      password,
		From:          strings.TrimSpace(getenv("SMTP_FROM")),
		ImplicitTLS:   port == 465,
		AllowInsecure: devMode,
	})
}

// LogMailer prints emails to a log instead of sending them. Development only: the log shows
// the sign-in code, so it must never be used in production.
type LogMailer struct {
	logger *log.Logger
}

func NewLogMailer(logger *log.Logger) *LogMailer {
	return &LogMailer{logger: logger}
}

func (m *LogMailer) Send(ctx context.Context, to, subject, text, html string) error {
	if _, err := parseRecipient(to); err != nil {
		return err
	}
	m.logger.Printf("email (not sent, SMTP is not set up) to %s\nSubject: %s\n\n%s", to, subject, text)
	return nil
}

// parseRecipient accepts one plain address. A display name, a list or a line break could add
// recipients or headers to the message.
func parseRecipient(to string) (mail.Address, error) {
	to = strings.TrimSpace(to)
	if to == "" || strings.ContainsAny(to, "\r\n") {
		return mail.Address{}, fmt.Errorf("email: %q is not a valid recipient", to)
	}
	parsed, err := mail.ParseAddress(to)
	if err != nil || parsed.Name != "" || parsed.Address != to {
		return mail.Address{}, fmt.Errorf("email: %q is not a valid recipient", to)
	}
	return *parsed, nil
}
