package auth

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"net/mail"
	"time"

	"server/internal/services/email"
	"server/internal/services/user/caregiver"
)

const (
	// LoginCodeTTL is how long a sign-in code works
	LoginCodeTTL = 10 * time.Minute
	// A code is burned after this many wrong guesses, so 1,000,000 possible codes cannot be tried
	loginCodeMaxAttempts = 5

	// Sending codes is limited per email and per IP. The windows slide: a request stops counting
	// once it is older than the window.
	maxCodesPerEmailPer15Min = 3
	maxCodesPerEmailPerDay   = 10
	maxCodesPerIPPerHour     = 10

	// Codes and request logs are kept this long, then deleted
	loginCodeRetention = 24 * time.Hour

	minLoginCodeSecretBytes = 32
)

var (
	ErrInvalidEmail    = errors.New("that is not a valid email address")
	ErrTooManyRequests = errors.New("too many sign-in codes requested")
	// ErrCodeRejected means the code is wrong, expired or already used (401)
	ErrCodeRejected = errors.New("the sign-in code is wrong or has expired")
	// ErrTooManyAttempts means the code was guessed wrong too many times and no longer works (429)
	ErrTooManyAttempts = errors.New("too many wrong sign-in codes")
)

// LoginCodeService sends and checks emailed sign-in codes. It knows nothing about caregivers
// or sessions: AuthService turns a checked code into a session.
type LoginCodeService interface {
	// RequestCode emails a new code to the address. An earlier unused code stops working.
	// ip is the caller's address, used only for rate limiting.
	RequestCode(ctx context.Context, address, ip string) error
	// VerifyCode checks a code and uses it up, so it works once.
	VerifyCode(ctx context.Context, address, code string) error
}

type loginCodeService struct {
	db     *sql.DB
	mailer email.Mailer
	secret []byte
	now    func() time.Time
	// generate makes a new code; tests replace it to know the code
	generate func() (string, error)
}

// NewLoginCodeService needs a mailer and a secret of at least 32 bytes. The secret is the key
// codes and IP addresses are hashed with, so a copy of the database alone cannot be used to
// find a code (there are only 1,000,000) or to see who asked.
func NewLoginCodeService(db *sql.DB, mailer email.Mailer, secret []byte) (*loginCodeService, error) {
	if mailer == nil {
		return nil, errors.New("sign-in codes need a mailer")
	}
	if len(secret) < minLoginCodeSecretBytes {
		return nil, fmt.Errorf("LOGIN_CODE_SECRET must be at least %d characters", minLoginCodeSecretBytes)
	}
	return &loginCodeService{db: db, mailer: mailer, secret: secret, now: time.Now, generate: randomLoginCode}, nil
}

// randomLoginCode makes 6 digits from the system's secure random source
func randomLoginCode() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1_000_000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}

// mac is a keyed hash (HMAC-SHA256). The parts are separated, so ("ab", "c") and ("a", "bc")
// give different hashes.
func (s *loginCodeService) mac(parts ...string) string {
	h := hmac.New(sha256.New, s.secret)
	for _, part := range parts {
		h.Write([]byte(part))
		h.Write([]byte{0})
	}
	return hex.EncodeToString(h.Sum(nil))
}

func (s *loginCodeService) codeHash(address, code string) string {
	return s.mac("login-code", address, code)
}

func (s *loginCodeService) ipHash(ip string) string {
	return s.mac("login-ip", ip)
}

// validAddress accepts one plain email address
func validAddress(address string) bool {
	if address == "" || len(address) > 254 {
		return false
	}
	parsed, err := mail.ParseAddress(address)
	return err == nil && parsed.Name == "" && parsed.Address == address
}

// validCodeFormat is exactly 6 digits
func validCodeFormat(code string) bool {
	if len(code) != 6 {
		return false
	}
	for _, c := range code {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

func (s *loginCodeService) RequestCode(ctx context.Context, rawAddress, ip string) error {
	address := caregiver.NormalizeEmail(rawAddress)
	if !validAddress(address) {
		return ErrInvalidEmail
	}
	code, err := s.generate()
	if err != nil {
		return fmt.Errorf("making a sign-in code: %w", err)
	}

	now := s.now()
	if err := s.storeCode(ctx, address, s.ipHash(ip), s.codeHash(address, code), now); err != nil {
		return err
	}

	// The code is stored before it is sent, so a code that arrives always works
	text, html := email.LoginCodeMessage(code, LoginCodeTTL)
	if err := s.mailer.Send(ctx, address, email.LoginCodeSubject, text, html); err != nil {
		return fmt.Errorf("sending the sign-in code: %w", err)
	}
	return nil
}

// storeCode checks the rate limits and, if they allow it, replaces the address's code with a
// new one. It holds a lock on the address, so two requests at once cannot both pass a limit.
// The IP limit is not locked, so many requests at the very same moment can pass it by a few.
func (s *loginCodeService) storeCode(ctx context.Context, address, ipHash, codeHash string, now time.Time) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock(hashtext($1))`, address); err != nil {
		return err
	}

	retentionCutoff := now.Add(-loginCodeRetention)
	if _, err := tx.ExecContext(ctx, `DELETE FROM "LoginCodeRequest" WHERE created_at < $1`, retentionCutoff); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM "LoginCode" WHERE created_at < $1`, retentionCutoff); err != nil {
		return err
	}

	var recentForEmail, todayForEmail int
	err = tx.QueryRowContext(ctx, `
		SELECT count(*) FILTER (WHERE created_at > $3), count(*)
		FROM "LoginCodeRequest"
		WHERE lower(email) = $1 AND created_at > $2`,
		address, now.Add(-24*time.Hour), now.Add(-15*time.Minute),
	).Scan(&recentForEmail, &todayForEmail)
	if err != nil {
		return err
	}
	var lastHourForIP int
	err = tx.QueryRowContext(ctx, `
		SELECT count(*) FROM "LoginCodeRequest" WHERE ip_hash = $1 AND created_at > $2`,
		ipHash, now.Add(-time.Hour),
	).Scan(&lastHourForIP)
	if err != nil {
		return err
	}
	if recentForEmail >= maxCodesPerEmailPer15Min || todayForEmail >= maxCodesPerEmailPerDay || lastHourForIP >= maxCodesPerIPPerHour {
		return ErrTooManyRequests
	}

	// Only the newest code works
	if _, err := tx.ExecContext(ctx, `
		UPDATE "LoginCode" SET used_at = $2 WHERE lower(email) = $1 AND used_at IS NULL`,
		address, now); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO "LoginCode" (email, code_hash, created_at, expires_at) VALUES ($1, $2, $3, $4)`,
		address, codeHash, now, now.Add(LoginCodeTTL)); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO "LoginCodeRequest" (email, ip_hash, created_at) VALUES ($1, $2, $3)`,
		address, ipHash, now); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *loginCodeService) VerifyCode(ctx context.Context, rawAddress, code string) error {
	address := caregiver.NormalizeEmail(rawAddress)
	if !validAddress(address) || !validCodeFormat(code) {
		return ErrCodeRejected
	}
	now := s.now()

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// The row stays locked until the end, so guesses at one code are counted one by one
	// and a correct code cannot be used twice at the same moment
	var id, attempts int
	var storedHash string
	err = tx.QueryRowContext(ctx, `
		SELECT id, code_hash, attempts FROM "LoginCode"
		WHERE lower(email) = $1 AND used_at IS NULL AND expires_at > $2
		ORDER BY created_at DESC, id DESC
		LIMIT 1
		FOR UPDATE`, address, now).Scan(&id, &storedHash, &attempts)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrCodeRejected
	}
	if err != nil {
		return err
	}
	if attempts >= loginCodeMaxAttempts {
		return ErrTooManyAttempts
	}

	// The guess is counted before it is compared, and kept whatever the outcome
	if _, err := tx.ExecContext(ctx, `UPDATE "LoginCode" SET attempts = attempts + 1 WHERE id = $1`, id); err != nil {
		return err
	}

	if hmac.Equal([]byte(storedHash), []byte(s.codeHash(address, code))) {
		if _, err := tx.ExecContext(ctx, `UPDATE "LoginCode" SET used_at = $2 WHERE id = $1`, id, now); err != nil {
			return err
		}
		return tx.Commit()
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	if attempts+1 >= loginCodeMaxAttempts {
		return ErrTooManyAttempts
	}
	return ErrCodeRejected
}
