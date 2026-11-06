package email

import (
	"bufio"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"fmt"
	"math/big"
	"net"
	"strings"
	"sync"
	"testing"
	"time"
)

// delivered is one message the fake server accepted
type delivered struct {
	from   string
	to     []string
	data   string
	authed bool
	tls    bool
}

// fakeSMTP is a small SMTP server on localhost, so the mailer is tested without a network
type fakeSMTP struct {
	startTLS   bool   // offers STARTTLS
	implicit   bool   // speaks TLS from the first byte
	user, pass string // when set, the server requires AUTH PLAIN with these
	rejectRcpt bool

	listener net.Listener
	cert     tls.Certificate
	roots    *x509.CertPool

	mu           sync.Mutex
	connections  int
	authAttempts int
	messages     []delivered
}

func newFakeSMTP(t *testing.T, configure func(*fakeSMTP)) *fakeSMTP {
	t.Helper()
	s := &fakeSMTP{}
	if configure != nil {
		configure(s)
	}
	s.cert, s.roots = testCertificate(t)

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	s.listener = ln
	t.Cleanup(func() { ln.Close() })

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			s.mu.Lock()
			s.connections++
			s.mu.Unlock()
			go s.handle(conn)
		}
	}()
	return s
}

func (s *fakeSMTP) port() int {
	return s.listener.Addr().(*net.TCPAddr).Port
}

func (s *fakeSMTP) delivered() []delivered {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]delivered(nil), s.messages...)
}

func (s *fakeSMTP) counts() (connections, authAttempts int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.connections, s.authAttempts
}

func (s *fakeSMTP) handle(conn net.Conn) {
	defer func() { conn.Close() }()
	serverTLS := &tls.Config{Certificates: []tls.Certificate{s.cert}}
	isTLS := false
	if s.implicit {
		conn = tls.Server(conn, serverTLS)
		isTLS = true
	}
	reader := bufio.NewReader(conn)
	reply := func(format string, args ...any) { fmt.Fprintf(conn, format+"\r\n", args...) }

	reply("220 localhost ESMTP")
	var from string
	var to []string
	authed := false
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return
		}
		line = strings.TrimRight(line, "\r\n")
		upper := strings.ToUpper(line)
		switch {
		case strings.HasPrefix(upper, "EHLO"), strings.HasPrefix(upper, "HELO"):
			reply("250-localhost")
			if s.startTLS && !isTLS {
				reply("250-STARTTLS")
			}
			if s.user != "" {
				reply("250-AUTH PLAIN")
			}
			reply("250 8BITMIME")
		case upper == "STARTTLS":
			if !s.startTLS {
				reply("502 not supported")
				continue
			}
			reply("220 ready to start TLS")
			secured := tls.Server(conn, serverTLS)
			if err := secured.Handshake(); err != nil {
				return
			}
			conn = secured
			reader = bufio.NewReader(secured)
			isTLS = true
		case strings.HasPrefix(upper, "AUTH PLAIN"):
			s.mu.Lock()
			s.authAttempts++
			s.mu.Unlock()
			fields := strings.Fields(line)
			if len(fields) < 3 {
				reply("501 expected credentials")
				continue
			}
			raw, err := base64.StdEncoding.DecodeString(fields[2])
			parts := strings.Split(string(raw), "\x00")
			if err == nil && len(parts) == 3 && parts[1] == s.user && parts[2] == s.pass {
				authed = true
				reply("235 authenticated")
			} else {
				reply("535 bad credentials")
			}
		case strings.HasPrefix(upper, "MAIL FROM:"):
			if s.user != "" && !authed {
				reply("530 authentication required")
				continue
			}
			from = addressIn(line)
			reply("250 ok")
		case strings.HasPrefix(upper, "RCPT TO:"):
			if s.rejectRcpt {
				reply("550 no such user")
				continue
			}
			to = append(to, addressIn(line))
			reply("250 ok")
		case upper == "DATA":
			reply("354 end data with a dot on its own line")
			var data strings.Builder
			for {
				part, err := reader.ReadString('\n')
				if err != nil {
					return
				}
				if part == ".\r\n" {
					break
				}
				data.WriteString(strings.TrimPrefix(part, "."))
			}
			s.mu.Lock()
			s.messages = append(s.messages, delivered{from: from, to: to, data: data.String(), authed: authed, tls: isTLS})
			s.mu.Unlock()
			reply("250 queued")
		case upper == "QUIT":
			reply("221 bye")
			return
		case upper == "RSET", upper == "NOOP":
			reply("250 ok")
		default:
			reply("500 unknown command")
		}
	}
}

// testCertificate makes a self-signed certificate for 127.0.0.1, and a pool that trusts it
func testCertificate(t *testing.T) (tls.Certificate, *x509.CertPool) {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	template := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "fake smtp"},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
		IPAddresses:           []net.IP{net.ParseIP("127.0.0.1")},
	}
	der, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := x509.ParseCertificate(der)
	if err != nil {
		t.Fatal(err)
	}
	pool := x509.NewCertPool()
	pool.AddCert(parsed)
	return tls.Certificate{Certificate: [][]byte{der}, PrivateKey: key}, pool
}

// mailerFor points a mailer at the fake server and trusts its certificate
func mailerFor(t *testing.T, s *fakeSMTP, change func(*Config)) *SMTPMailer {
	t.Helper()
	cfg := Config{
		Host: "127.0.0.1",
		Port: s.port(),
		From: "D.A.W.N. <no-reply@example.com>",
	}
	if change != nil {
		change(&cfg)
	}
	m, err := NewSMTPMailer(cfg)
	if err != nil {
		t.Fatal(err)
	}
	m.tlsConfig = &tls.Config{RootCAs: s.roots}
	return m
}

// addressIn returns what is between < and > in a MAIL FROM or RCPT TO line, ignoring parameters
// such as BODY=8BITMIME after it
func addressIn(line string) string {
	start, end := strings.Index(line, "<"), strings.Index(line, ">")
	if start < 0 || end < start {
		return ""
	}
	return line[start+1 : end]
}
