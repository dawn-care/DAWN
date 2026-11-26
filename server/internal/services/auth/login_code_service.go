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
	"log"
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
	// Wrong guesses are limited per IP across every address, so one caller cannot burn other
	// people's codes, or guess at many addresses, in bulk. A right guess does not count.
	maxWrongGuessesPerIPPerHour = 20

	// Codes, request logs and wrong guesses older than this are deleted the next time a code
	// request goes through. There is no scheduled cleanup.
	loginCodeRetention = 24 * time.Hour

	minLoginCodeSecretBytes = 32
)

var (
	ErrInvalidEmail = errors.New("that is not a valid email address")
	// ErrTooManyRequests means a sending limit was reached (429). It comes as a *LimitError.
	ErrTooManyRequests = errors.New("too many sign-in codes requested")
	// ErrSendFailed means the code could not be emailed (502). Nothing was stored or counted,
	// so an earlier code still works.
	ErrSendFailed = errors.New("the sign-in code could not be emailed")
	// ErrCodeRejected means the code is wrong, expired or already used (401)
	ErrCodeRejected = errors.New("the sign-in code is wrong or has expired")
	// ErrTooManyAttempts means the code was guessed wrong too many times and no longer works (429)
	ErrTooManyAttempts = errors.New("too many wrong sign-in codes")
	// ErrTooManyGuesses means this IP made too many wrong guesses lately (429), whatever the
	// address. It comes as a *LimitError. The code was not checked, so it still works.
	ErrTooManyGuesses = errors.New("too many wrong sign-in codes from this network")
)

// LimitError is a rate limit that was reached. errors.Is matches it to Limit
// (ErrTooManyRequests or ErrTooManyGuesses).
type LimitError struct {
	Limit error
	// RetryAfter is how long until another try would be allowed: until enough of the requests
	// (or wrong guesses) that fill each full window have aged out of it. Whole seconds, at least 1.
	RetryAfter time.Duration
}

func (e *LimitError) Error() string { return e.Limit.Error() }
func (e *LimitError) Unwrap() error { return e.Limit }

// LoginCodeService sends and checks emailed sign-in codes. It knows nothing about caregivers
// or sessions: AuthService turns a checked code into a session.
type LoginCodeService interface {
	// RequestCode emails a new code to the address. An earlier unused code stops working.
	// ip is the caller's address, used only for rate limiting.
	RequestCode(ctx context.Context, address, ip string) error
	// VerifyCode checks a code and, if it is right, runs onAccepted and uses the code up. If
	// onAccepted fails, the code is not used up, so the caregiver can enter it again.
	// ip is the caller's address, used only for rate limiting.
	VerifyCode(ctx context.Context, address, code, ip string, onAccepted func() error) error
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

// RequestCode checks the rate limits, stores the new code and emails it, all in one
// transaction: the code is committed only once the mail server has accepted the email. A
// failed send therefore leaves the earlier code working and does not count toward the limits.
// It holds a lock on the address, so two requests at once cannot both pass a limit. The IP
// limit is not locked, so many requests at the very same moment can pass it by a few.
func (s *loginCodeService) RequestCode(ctx context.Context, rawAddress, ip string) error {
	address := caregiver.NormalizeEmail(rawAddress)
	if !validAddress(address) {
		return ErrInvalidEmail
	}
	code, err := s.generate()
	if err != nil {
		return fmt.Errorf("making a sign-in code: %w", err)
	}
	// A caller who closes the page must not cut off a send that is under way, or roll back a
	// code the mail server has already accepted. The mailer has its own timeout.
	ctx = context.WithoutCancel(ctx)
	now := s.now()

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("storing a sign-in code: %w", err)
	}
	defer tx.Rollback()

	if err := s.checkSendLimits(ctx, tx, address, s.ipHash(ip), now); err != nil {
		return err
	}
	if err := s.storeCode(ctx, tx, address, s.ipHash(ip), s.codeHash(address, code), now); err != nil {
		return fmt.Errorf("storing a sign-in code: %w", err)
	}

	text, html := email.LoginCodeMessage(code, LoginCodeTTL)
	if err := s.mailer.Send(ctx, address, email.LoginCodeSubject, text, html); err != nil {
		return fmt.Errorf("%w: %w", ErrSendFailed, err)
	}
	// The email is out; if this fails the code in it will not work, which is rare and is
	// reported as an error so the caregiver asks again
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("saving a sign-in code that was already emailed: %w", err)
	}

	s.deleteOldRows(ctx, now)
	return nil
}

// checkSendLimits takes the address's lock and checks the sending limits
func (s *loginCodeService) checkSendLimits(ctx context.Context, tx *sql.Tx, address, ipHash string, now time.Time) error {
	if _, err := tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock(hashtext($1))`, address); err != nil {
		return fmt.Errorf("locking the address for a sign-in code: %w", err)
	}

	var recentForEmail, todayForEmail int
	err := tx.QueryRowContext(ctx, `
		SELECT count(*) FILTER (WHERE created_at > $3), count(*)
		FROM "LoginCodeRequest"
		WHERE lower(email) = $1 AND created_at > $2`,
		address, now.Add(-24*time.Hour), now.Add(-15*time.Minute),
	).Scan(&recentForEmail, &todayForEmail)
	if err != nil {
		return fmt.Errorf("counting sign-in codes for the address: %w", err)
	}
	var lastHourForIP int
	err = tx.QueryRowContext(ctx, `
		SELECT count(*) FROM "LoginCodeRequest" WHERE ip_hash = $1 AND created_at > $2`,
		ipHash, now.Add(-time.Hour),
	).Scan(&lastHourForIP)
	if err != nil {
		return fmt.Errorf("counting sign-in codes for the IP: %w", err)
	}

	// A new code needs room in every window, so the wait is the longest of the full windows' waits
	windows := []struct {
		full   bool
		filter string
		key    string
		limit  int
		length time.Duration
	}{
		{todayForEmail >= maxCodesPerEmailPerDay, emailFilter, address, maxCodesPerEmailPerDay, 24 * time.Hour},
		{recentForEmail >= maxCodesPerEmailPer15Min, emailFilter, address, maxCodesPerEmailPer15Min, 15 * time.Minute},
		{lastHourForIP >= maxCodesPerIPPerHour, ipFilter, ipHash, maxCodesPerIPPerHour, time.Hour},
	}
	var wait time.Duration
	for _, window := range windows {
		if !window.full {
			continue
		}
		reopens, err := reopensIn(ctx, tx, "LoginCodeRequest", window.filter, window.key, window.limit, window.length, now)
		if err != nil {
			return fmt.Errorf("working out when sign-in codes can be sent again: %w", err)
		}
		wait = max(wait, reopens)
	}
	if wait > 0 {
		return &LimitError{Limit: ErrTooManyRequests, RetryAfter: wait}
	}
	return nil
}

// The rows a limit counts: by address (stored lowercase) or by hashed IP
const (
	emailFilter = `lower(email) = $1`
	ipFilter    = `ip_hash = $1`
)

// reopensIn says how long until a full window has room again. With limit rows counted in the
// window, the limit-th newest is the one that has to age out for the count to drop below the
// limit (older ones go first). A row stops counting once it is length old. If that row has
// already gone (a cleanup, say), the whole length is the safe answer.
func reopensIn(ctx context.Context, tx *sql.Tx, table, filter, key string, limit int, length time.Duration, now time.Time) (time.Duration, error) {
	var createdAt time.Time
	err := tx.QueryRowContext(ctx, `
		SELECT created_at FROM "`+table+`"
		WHERE `+filter+` AND created_at > $2
		ORDER BY created_at DESC
		OFFSET $3 LIMIT 1`,
		key, now.Add(-length), limit-1,
	).Scan(&createdAt)
	if errors.Is(err, sql.ErrNoRows) {
		return length, nil
	}
	if err != nil {
		return 0, err
	}
	return wholeSecondsUp(createdAt.Add(length).Sub(now)), nil
}

// wholeSecondsUp rounds up to a whole second, at least 1, so a caller who waits that long is
// never early
func wholeSecondsUp(d time.Duration) time.Duration {
	rounded := d.Truncate(time.Second)
	if rounded < d {
		rounded += time.Second
	}
	return max(rounded, time.Second)
}

// storeCode replaces the address's code with a new one and logs the request
func (s *loginCodeService) storeCode(ctx context.Context, tx *sql.Tx, address, ipHash, codeHash string, now time.Time) error {
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
	_, err := tx.ExecContext(ctx, `
		INSERT INTO "LoginCodeRequest" (email, ip_hash, created_at) VALUES ($1, $2, $3)`,
		address, ipHash, now)
	return err
}

// deleteOldRows removes codes, request logs and wrong guesses past their retention. It runs
// after a request has gone through and only logs a failure: cleanup must never stop sign-in.
func (s *loginCodeService) deleteOldRows(ctx context.Context, now time.Time) {
	cutoff := now.Add(-loginCodeRetention)
	for _, table := range []string{"LoginCodeRequest", "LoginCode", "LoginCodeGuess"} {
		if _, err := s.db.ExecContext(ctx, `DELETE FROM "`+table+`" WHERE created_at < $1`, cutoff); err != nil {
			log.Printf("deleting old %s rows failed: %v", table, err)
		}
	}
}

func (s *loginCodeService) VerifyCode(ctx context.Context, rawAddress, code, ip string, onAccepted func() error) error {
	address := caregiver.NormalizeEmail(rawAddress)
	if !validAddress(address) || !validCodeFormat(code) {
		return ErrCodeRejected
	}
	now := s.now()
	ipHash := s.ipHash(ip)

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("checking a sign-in code: %w", err)
	}
	defer tx.Rollback()

	// Checked before the code is looked at, so a caller over the limit cannot burn it
	var wrongFromIP int
	if err := tx.QueryRowContext(ctx, `
		SELECT count(*) FROM "LoginCodeGuess" WHERE ip_hash = $1 AND created_at > $2`,
		ipHash, now.Add(-time.Hour)).Scan(&wrongFromIP); err != nil {
		return fmt.Errorf("counting wrong sign-in codes for the IP: %w", err)
	}
	if wrongFromIP >= maxWrongGuessesPerIPPerHour {
		wait, err := reopensIn(ctx, tx, "LoginCodeGuess", ipFilter, ipHash, maxWrongGuessesPerIPPerHour, time.Hour, now)
		if err != nil {
			return fmt.Errorf("working out when sign-in codes can be checked again: %w", err)
		}
		return &LimitError{Limit: ErrTooManyGuesses, RetryAfter: wait}
	}

	// reject records a wrong guess from this IP, keeps it and the code's attempt count, and
	// returns the rejection
	reject := func(rejection error) error {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO "LoginCodeGuess" (ip_hash, created_at) VALUES ($1, $2)`, ipHash, now); err != nil {
			return fmt.Errorf("recording a wrong sign-in code: %w", err)
		}
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("recording a wrong sign-in code: %w", err)
		}
		return rejection
	}

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
		return reject(ErrCodeRejected)
	}
	if err != nil {
		return fmt.Errorf("finding the sign-in code: %w", err)
	}
	if attempts >= loginCodeMaxAttempts {
		return reject(ErrTooManyAttempts)
	}

	// The guess is counted before it is compared, and kept if it is wrong
	if _, err := tx.ExecContext(ctx, `UPDATE "LoginCode" SET attempts = attempts + 1 WHERE id = $1`, id); err != nil {
		return fmt.Errorf("counting a sign-in code guess: %w", err)
	}

	if !hmac.Equal([]byte(storedHash), []byte(s.codeHash(address, code))) {
		if attempts+1 >= loginCodeMaxAttempts {
			return reject(ErrTooManyAttempts)
		}
		return reject(ErrCodeRejected)
	}

	// The code is right. It is used up only if signing in works, so a failure here (the
	// database, say) leaves it for another try; the row lock still stops a second use meanwhile.
	if err := onAccepted(); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE "LoginCode" SET used_at = $2 WHERE id = $1`, id, now); err != nil {
		return fmt.Errorf("using up the sign-in code: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("using up the sign-in code: %w", err)
	}
	return nil
}
