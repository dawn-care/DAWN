package email

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/tls"
	"encoding/hex"
	"errors"
	"fmt"
	"mime"
	"mime/multipart"
	"mime/quotedprintable"
	"net"
	"net/mail"
	"net/smtp"
	"net/textproto"
	"strconv"
	"strings"
	"time"
)

// ErrInsecureServer means the mail server does not offer TLS, so nothing was sent
var ErrInsecureServer = errors.New("email: the mail server does not offer TLS (STARTTLS), so nothing was sent")

const defaultTimeout = 20 * time.Second

// SMTPMailer sends email through an SMTP server. It only sends over TLS: either from the first
// byte (ImplicitTLS) or after STARTTLS, and the server's certificate must be valid. Without
// TLS it refuses, in development too. The password is only sent
// over TLS (or to localhost).
type SMTPMailer struct {
	cfg     Config
	from    mail.Address
	timeout time.Duration
	// tlsConfig replaces the default TLS settings; tests use it to trust a test certificate
	tlsConfig *tls.Config
}

func NewSMTPMailer(cfg Config) (*SMTPMailer, error) {
	if cfg.Host == "" {
		return nil, errors.New("email: SMTP_HOST is required")
	}
	if cfg.Port == 0 {
		cfg.Port = 587
	}
	from, err := mail.ParseAddress(cfg.From)
	if err != nil {
		return nil, fmt.Errorf("email: SMTP_FROM %q is not a valid sender: %w", cfg.From, err)
	}
	return &SMTPMailer{cfg: cfg, from: *from, timeout: defaultTimeout}, nil
}

func (m *SMTPMailer) Send(ctx context.Context, to, subject, text, html string) error {
	rcpt, err := parseRecipient(to)
	if err != nil {
		return err
	}
	msg, err := buildMessage(m.from, rcpt, subject, text, html, time.Now())
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(ctx, m.timeout)
	defer cancel()

	addr := net.JoinHostPort(m.cfg.Host, strconv.Itoa(m.cfg.Port))
	conn, err := m.dial(ctx, addr)
	if err != nil {
		return fmt.Errorf("email: connect to %s: %w", addr, err)
	}
	defer conn.Close()
	if deadline, ok := ctx.Deadline(); ok {
		conn.SetDeadline(deadline)
	}
	// Closing the connection also stops a server that has gone quiet mid-conversation
	stop := context.AfterFunc(ctx, func() { conn.Close() })
	defer stop()

	client, err := smtp.NewClient(conn, m.cfg.Host)
	if err != nil {
		return fmt.Errorf("email: greeting from %s: %w", addr, err)
	}
	defer client.Close()

	if !m.cfg.ImplicitTLS {
		if ok, _ := client.Extension("STARTTLS"); ok {
			if err := client.StartTLS(m.tlsSettings()); err != nil {
				return fmt.Errorf("email: start TLS with %s: %w", addr, err)
			}
		} else {
			return ErrInsecureServer
		}
	}

	if m.cfg.User != "" {
		// PlainAuth itself refuses to send the password over a connection with no TLS,
		// except to localhost
		if err := client.Auth(smtp.PlainAuth("", m.cfg.User, m.cfg.Password, m.cfg.Host)); err != nil {
			return fmt.Errorf("email: sign in to %s: %w", addr, err)
		}
	}

	if err := client.Mail(m.from.Address); err != nil {
		return fmt.Errorf("email: sender rejected: %w", err)
	}
	if err := client.Rcpt(rcpt.Address); err != nil {
		return fmt.Errorf("email: recipient rejected: %w", err)
	}
	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("email: start message: %w", err)
	}
	if _, err := w.Write(msg); err != nil {
		return fmt.Errorf("email: send message: %w", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("email: message not accepted: %w", err)
	}
	// The server has accepted the message, so a failure to say goodbye changes nothing
	client.Quit()
	return nil
}

func (m *SMTPMailer) dial(ctx context.Context, addr string) (net.Conn, error) {
	if m.cfg.ImplicitTLS {
		dialer := tls.Dialer{NetDialer: &net.Dialer{}, Config: m.tlsSettings()}
		return dialer.DialContext(ctx, "tcp", addr)
	}
	return (&net.Dialer{}).DialContext(ctx, "tcp", addr)
}

func (m *SMTPMailer) tlsSettings() *tls.Config {
	var cfg *tls.Config
	if m.tlsConfig != nil {
		cfg = m.tlsConfig.Clone()
	} else {
		cfg = &tls.Config{}
	}
	cfg.ServerName = m.cfg.Host
	if cfg.MinVersion < tls.VersionTLS12 {
		cfg.MinVersion = tls.VersionTLS12
	}
	return cfg
}

// buildMessage writes the message with a plain text and an HTML version; mail programs show
// the last one they understand, so the HTML goes last.
func buildMessage(from, to mail.Address, subject, text, html string, now time.Time) ([]byte, error) {
	if strings.ContainsAny(subject, "\r\n") {
		return nil, errors.New("email: the subject must be a single line")
	}
	id, err := messageID(from.Address)
	if err != nil {
		return nil, err
	}

	var body bytes.Buffer
	parts := multipart.NewWriter(&body)
	if err := writePart(parts, "text/plain", text); err != nil {
		return nil, err
	}
	if err := writePart(parts, "text/html", html); err != nil {
		return nil, err
	}
	if err := parts.Close(); err != nil {
		return nil, err
	}

	var msg bytes.Buffer
	header := func(name, value string) { msg.WriteString(name + ": " + value + "\r\n") }
	header("From", from.String())
	header("To", to.String())
	header("Subject", mime.QEncoding.Encode("utf-8", subject))
	header("Date", now.Format(time.RFC1123Z))
	header("Message-ID", id)
	header("MIME-Version", "1.0")
	header("Content-Type", mime.FormatMediaType("multipart/alternative", map[string]string{"boundary": parts.Boundary()}))
	msg.WriteString("\r\n")
	msg.Write(body.Bytes())
	return msg.Bytes(), nil
}

func writePart(parts *multipart.Writer, contentType, content string) error {
	header := textproto.MIMEHeader{}
	header.Set("Content-Type", contentType+"; charset=utf-8")
	header.Set("Content-Transfer-Encoding", "quoted-printable")
	w, err := parts.CreatePart(header)
	if err != nil {
		return err
	}
	encoded := quotedprintable.NewWriter(w)
	if _, err := encoded.Write([]byte(content)); err != nil {
		return err
	}
	return encoded.Close()
}

func messageID(sender string) (string, error) {
	random := make([]byte, 16)
	if _, err := rand.Read(random); err != nil {
		return "", err
	}
	domain := "localhost"
	if at := strings.LastIndex(sender, "@"); at >= 0 {
		domain = sender[at+1:]
	}
	return "<" + hex.EncodeToString(random) + "@" + domain + ">", nil
}
