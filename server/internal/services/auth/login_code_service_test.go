package auth

import (
	"context"
	"errors"
	"regexp"
	"sync"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"server/internal/services/email"
)

/* -------------------------------------------------------
   SETUP
------------------------------------------------------- */

const testCodeSecret = "0123456789abcdef0123456789abcdef"

var fixedNow = time.Date(2026, 11, 6, 12, 0, 0, 0, time.UTC)

type sentEmail struct{ to, subject, text, html string }

// codeMailer records the emails the service sends
type codeMailer struct {
	mu   sync.Mutex
	sent []sentEmail
	err  error
}

func (m *codeMailer) Send(ctx context.Context, to, subject, text, html string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.err != nil {
		return m.err
	}
	m.sent = append(m.sent, sentEmail{to, subject, text, html})
	return nil
}

func (m *codeMailer) emails() []sentEmail {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]sentEmail(nil), m.sent...)
}

func newCodeServiceWithMock(t *testing.T) (*loginCodeService, sqlmock.Sqlmock, *codeMailer) {
	t.Helper()
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	mailer := &codeMailer{}
	svc, err := NewLoginCodeService(db, mailer, []byte(testCodeSecret))
	require.NoError(t, err)
	svc.now = func() time.Time { return fixedNow }
	svc.generate = func() (string, error) { return "123456", nil }
	return svc, mock, mailer
}

func q(fragment string) string { return regexp.QuoteMeta(fragment) }

const (
	lockSQL            = `SELECT pg_advisory_xact_lock(hashtext($1))`
	deleteRequestsSQL  = `DELETE FROM "LoginCodeRequest" WHERE created_at < $1`
	deleteCodesSQL     = `DELETE FROM "LoginCode" WHERE created_at < $1`
	deleteGuessesSQL   = `DELETE FROM "LoginCodeGuess" WHERE created_at < $1`
	countByEmailSQL    = `SELECT count(*) FILTER (WHERE created_at > $3), count(*) FROM "LoginCodeRequest" WHERE lower(email) = $1 AND created_at > $2`
	countByIPSQL       = `FROM "LoginCodeRequest" WHERE ip_hash = $1 AND created_at > $2`
	supersedeSQL       = `UPDATE "LoginCode" SET used_at = $2 WHERE lower(email) = $1 AND used_at IS NULL`
	insertCodeSQL      = `INSERT INTO "LoginCode" (email, code_hash, created_at, expires_at) VALUES ($1, $2, $3, $4)`
	insertRequestSQL   = `INSERT INTO "LoginCodeRequest" (email, ip_hash, created_at) VALUES ($1, $2, $3)`
	countGuessesSQL    = `SELECT count(*) FROM "LoginCodeGuess" WHERE ip_hash = $1 AND created_at > $2`
	insertGuessSQL     = `INSERT INTO "LoginCodeGuess" (ip_hash, created_at) VALUES ($1, $2)`
	selectCodeSQL      = `FROM "LoginCode" WHERE lower(email) = $1 AND used_at IS NULL AND expires_at > $2 ORDER BY created_at DESC, id DESC LIMIT 1 FOR UPDATE`
	countAttemptSQL    = `UPDATE "LoginCode" SET attempts = attempts + 1 WHERE id = $1`
	markUsedSQL        = `UPDATE "LoginCode" SET used_at = $2 WHERE id = $1`
	testAddress        = "pat@example.com"
	testIP             = "203.0.113.9"
	testIPSecondClient = "198.51.100.4"
)

// expectChecks expects everything a request does up to the rate limit decision
func expectChecks(mock sqlmock.Sqlmock, svc *loginCodeService, ip string, recent, today, fromIP int) {
	mock.ExpectBegin()
	mock.ExpectExec(q(lockSQL)).WithArgs(testAddress).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery(q(countByEmailSQL)).
		WithArgs(testAddress, fixedNow.Add(-24*time.Hour), fixedNow.Add(-15*time.Minute)).
		WillReturnRows(sqlmock.NewRows([]string{"recent", "today"}).AddRow(recent, today))
	mock.ExpectQuery(q(countByIPSQL)).
		WithArgs(svc.ipHash(ip), fixedNow.Add(-time.Hour)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(fromIP))
}

// expectStore expects the new code and the request log row to be written (not yet committed)
func expectStore(mock sqlmock.Sqlmock, svc *loginCodeService, ip, code string) {
	mock.ExpectExec(q(supersedeSQL)).WithArgs(testAddress, fixedNow).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(q(insertCodeSQL)).
		WithArgs(testAddress, svc.codeHash(testAddress, code), fixedNow, fixedNow.Add(10*time.Minute)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(q(insertRequestSQL)).WithArgs(testAddress, svc.ipHash(ip), fixedNow).WillReturnResult(sqlmock.NewResult(1, 1))
}

// expectCleanup expects the retention deletes that run after a request goes through
func expectCleanup(mock sqlmock.Sqlmock) {
	cutoff := fixedNow.Add(-24 * time.Hour)
	mock.ExpectExec(q(deleteRequestsSQL)).WithArgs(cutoff).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(q(deleteCodesSQL)).WithArgs(cutoff).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(q(deleteGuessesSQL)).WithArgs(cutoff).WillReturnResult(sqlmock.NewResult(0, 0))
}

/* -------------------------------------------------------
   NewLoginCodeService
------------------------------------------------------- */

func TestRandomLoginCode_IsSixDigits(t *testing.T) {
	seen := map[string]bool{}
	for i := 0; i < 300; i++ {
		code, err := randomLoginCode()
		require.NoError(t, err)
		require.Regexp(t, `^[0-9]{6}$`, code)
		seen[code] = true
	}
	assert.Greater(t, len(seen), 250, "codes are not repeating")
}

func TestHashes_AreKeyedAndSeparated(t *testing.T) {
	svc, _, _ := newCodeServiceWithMock(t)
	other, err := NewLoginCodeService(nil, &codeMailer{}, []byte("another-secret-another-secret-123"))
	require.NoError(t, err)

	assert.Regexp(t, `^[0-9a-f]{64}$`, svc.codeHash(testAddress, "123456"))
	assert.NotEqual(t, svc.codeHash(testAddress, "123456"), other.codeHash(testAddress, "123456"), "another secret gives another hash")
	assert.NotEqual(t, svc.codeHash(testAddress, "123456"), svc.codeHash("sam@example.com", "123456"), "a code is tied to its address")
	assert.NotEqual(t, svc.codeHash(testAddress, "123456"), svc.codeHash(testAddress, "654321"))
	assert.NotContains(t, svc.ipHash(testIP), "203.0.113.9")
	assert.NotEqual(t, svc.mac("ab", "c"), svc.mac("a", "bc"))
}

/* -------------------------------------------------------
   RequestCode
------------------------------------------------------- */

func TestRequestCode_StoresAHashedCodeThenEmailsIt(t *testing.T) {
	svc, mock, mailer := newCodeServiceWithMock(t)
	expectChecks(mock, svc, testIP, 0, 0, 0)
	expectStore(mock, svc, testIP, "123456")
	mock.ExpectCommit()
	expectCleanup(mock)

	err := svc.RequestCode(context.Background(), "  Pat@Example.COM ", testIP)

	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
	sent := mailer.emails()
	require.Len(t, sent, 1)
	assert.Equal(t, testAddress, sent[0].to, "sent to the normalized address")
	assert.Equal(t, "Your D.A.W.N. sign-in code", sent[0].subject)
	assert.Contains(t, sent[0].text, "123456")
	assert.Contains(t, sent[0].html, "123456")
	assert.Contains(t, sent[0].text, "10 minutes")
}

func TestRequestCode_RefusesAtEveryLimitAndSendsNothing(t *testing.T) {
	cases := map[string]struct {
		counts     [3]int
		retryAfter time.Duration
	}{
		"3 for this email in 15 minutes": {[3]int{maxCodesPerEmailPer15Min, 3, 0}, 15 * time.Minute},
		"10 for this email in a day":     {[3]int{2, maxCodesPerEmailPerDay, 0}, 24 * time.Hour},
		"10 from this IP in an hour":     {[3]int{0, 0, maxCodesPerIPPerHour}, time.Hour},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			svc, mock, mailer := newCodeServiceWithMock(t)
			expectChecks(mock, svc, testIP, tc.counts[0], tc.counts[1], tc.counts[2])
			mock.ExpectRollback()

			err := svc.RequestCode(context.Background(), testAddress, testIP)

			assert.ErrorIs(t, err, ErrTooManyRequests)
			var limit *LimitError
			require.ErrorAs(t, err, &limit)
			assert.Equal(t, tc.retryAfter, limit.RetryAfter)
			assert.Empty(t, mailer.emails(), "no email is sent")
			assert.NoError(t, mock.ExpectationsWereMet(), "nothing was stored")
		})
	}
}

func TestRequestCode_RejectsInvalidAddressesBeforeTouchingTheDatabase(t *testing.T) {
	for _, address := range []string{
		"", "   ", "nope", "@example.com", "pat@", "Pat <pat@example.com>",
		"pat@example.com\nBcc: boss@example.com", "pat@example.com, sam@example.com",
	} {
		svc, mock, mailer := newCodeServiceWithMock(t)
		err := svc.RequestCode(context.Background(), address, testIP)
		assert.ErrorIs(t, err, ErrInvalidEmail, "%q", address)
		assert.Empty(t, mailer.emails())
		assert.NoError(t, mock.ExpectationsWereMet(), "%q opened no transaction", address)
	}
}

func TestRequestCode_AFailedSendStoresAndCountsNothing(t *testing.T) {
	svc, mock, mailer := newCodeServiceWithMock(t)
	mailer.err = errors.New("smtp down")
	expectChecks(mock, svc, testIP, 0, 0, 0)
	expectStore(mock, svc, testIP, "123456")
	mock.ExpectRollback() // the new code, the replaced old code and the request row are all undone

	err := svc.RequestCode(context.Background(), testAddress, testIP)

	assert.ErrorIs(t, err, ErrSendFailed)
	assert.ErrorIs(t, err, mailer.err)
	assert.NotErrorIs(t, err, ErrTooManyRequests)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRequestCode_ACallerWhoLeavesDoesNotCutOffTheSend(t *testing.T) {
	svc, mock, mailer := newCodeServiceWithMock(t)
	expectChecks(mock, svc, testIP, 0, 0, 0)
	expectStore(mock, svc, testIP, "123456")
	mock.ExpectCommit()
	expectCleanup(mock)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := svc.RequestCode(ctx, testAddress, testIP)

	assert.NoError(t, err)
	assert.Len(t, mailer.emails(), 1)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRequestCode_DatabaseErrors(t *testing.T) {
	t.Run("cannot count", func(t *testing.T) {
		svc, mock, mailer := newCodeServiceWithMock(t)
		mock.ExpectBegin()
		mock.ExpectExec(q(lockSQL)).WillReturnResult(sqlmock.NewResult(0, 0))
		mock.ExpectQuery(q(countByEmailSQL)).WillReturnError(errors.New("db down"))
		mock.ExpectRollback()
		err := svc.RequestCode(context.Background(), testAddress, testIP)
		assert.Error(t, err)
		assert.NotErrorIs(t, err, ErrTooManyRequests, "a database failure is not a rate limit")
		assert.Empty(t, mailer.emails())
		assert.NoError(t, mock.ExpectationsWereMet(), "rolled back")
	})
}

/* -------------------------------------------------------
   VerifyCode
------------------------------------------------------- */

// accept is an onAccepted that records it ran
func accept(ran *bool) func() error {
	return func() error { *ran = true; return nil }
}

// expectGuessCount expects the per-IP wrong guess count
func expectGuessCount(mock sqlmock.Sqlmock, svc *loginCodeService, wrong int) {
	mock.ExpectBegin()
	mock.ExpectQuery(q(countGuessesSQL)).WithArgs(svc.ipHash(testIP), fixedNow.Add(-time.Hour)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(wrong))
}

func expectCodeRow(mock sqlmock.Sqlmock, svc *loginCodeService, storedCode string, attempts int) {
	expectGuessCount(mock, svc, 0)
	mock.ExpectQuery(q(selectCodeSQL)).WithArgs(testAddress, fixedNow).
		WillReturnRows(sqlmock.NewRows([]string{"id", "code_hash", "attempts"}).
			AddRow(5, svc.codeHash(testAddress, storedCode), attempts))
}

// expectWrongGuessKept expects the wrong guess to be logged and committed
func expectWrongGuessKept(mock sqlmock.Sqlmock, svc *loginCodeService) {
	mock.ExpectExec(q(insertGuessSQL)).WithArgs(svc.ipHash(testIP), fixedNow).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
}

func TestVerifyCode_RightCodeSignsInThenIsUsedUp(t *testing.T) {
	svc, mock, _ := newCodeServiceWithMock(t)
	expectCodeRow(mock, svc, "123456", 0)
	mock.ExpectExec(q(countAttemptSQL)).WithArgs(5).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(q(markUsedSQL)).WithArgs(5, fixedNow).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	var ran bool

	err := svc.VerifyCode(context.Background(), " PAT@example.com", "123456", testIP, accept(&ran))

	assert.NoError(t, err)
	assert.True(t, ran)
	assert.NoError(t, mock.ExpectationsWereMet(), "a right guess is not logged against the IP")
}

func TestVerifyCode_WrongCodeCountsAnAttemptAndKeepsIt(t *testing.T) {
	svc, mock, _ := newCodeServiceWithMock(t)
	expectCodeRow(mock, svc, "123456", 1)
	mock.ExpectExec(q(countAttemptSQL)).WithArgs(5).WillReturnResult(sqlmock.NewResult(0, 1))
	expectWrongGuessKept(mock, svc) // committed, not rolled back, so the attempt is not lost
	var ran bool

	err := svc.VerifyCode(context.Background(), testAddress, "000000", testIP, accept(&ran))

	assert.ErrorIs(t, err, ErrCodeRejected)
	assert.False(t, ran)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestVerifyCode_TheFifthWrongCodeBurnsIt(t *testing.T) {
	svc, mock, _ := newCodeServiceWithMock(t)
	expectCodeRow(mock, svc, "123456", loginCodeMaxAttempts-1)
	mock.ExpectExec(q(countAttemptSQL)).WithArgs(5).WillReturnResult(sqlmock.NewResult(0, 1))
	expectWrongGuessKept(mock, svc)

	err := svc.VerifyCode(context.Background(), testAddress, "000000", testIP, accept(new(bool)))

	assert.ErrorIs(t, err, ErrTooManyAttempts)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestVerifyCode_ABurnedCodeIsRefusedEvenWhenTheGuessIsRight(t *testing.T) {
	svc, mock, _ := newCodeServiceWithMock(t)
	expectCodeRow(mock, svc, "123456", loginCodeMaxAttempts)
	expectWrongGuessKept(mock, svc)
	var ran bool

	err := svc.VerifyCode(context.Background(), testAddress, "123456", testIP, accept(&ran))

	assert.ErrorIs(t, err, ErrTooManyAttempts)
	assert.False(t, ran)
	assert.NoError(t, mock.ExpectationsWereMet(), "no further attempt is counted and nothing is used")
}

func TestVerifyCode_NoActiveCode(t *testing.T) {
	svc, mock, _ := newCodeServiceWithMock(t)
	expectGuessCount(mock, svc, 0)
	mock.ExpectQuery(q(selectCodeSQL)).WithArgs(testAddress, fixedNow).
		WillReturnRows(sqlmock.NewRows([]string{"id", "code_hash", "attempts"}))
	expectWrongGuessKept(mock, svc) // guessing at addresses with no code counts too

	err := svc.VerifyCode(context.Background(), testAddress, "123456", testIP, accept(new(bool)))

	assert.ErrorIs(t, err, ErrCodeRejected, "an expired, used or never sent code looks the same as a wrong one")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestVerifyCode_MalformedInputNeverReachesTheDatabase(t *testing.T) {
	for _, tc := range []struct{ address, code string }{
		{testAddress, ""},
		{testAddress, "12345"},
		{testAddress, "1234567"},
		{testAddress, "12345a"},
		{testAddress, " 12345"},
		{testAddress, "١٢٣٤٥٦"}, // six digits in another script
		{"", "123456"},
		{"not-an-email", "123456"},
	} {
		svc, mock, _ := newCodeServiceWithMock(t)
		err := svc.VerifyCode(context.Background(), tc.address, tc.code, testIP, accept(new(bool)))
		assert.ErrorIs(t, err, ErrCodeRejected, "%q %q", tc.address, tc.code)
		assert.NoError(t, mock.ExpectationsWereMet())
	}
}

func TestVerifyCode_DatabaseErrors(t *testing.T) {
	t.Run("cannot look up the code", func(t *testing.T) {
		svc, mock, _ := newCodeServiceWithMock(t)
		expectGuessCount(mock, svc, 0)
		mock.ExpectQuery(q(selectCodeSQL)).WillReturnError(errors.New("db down"))
		mock.ExpectRollback()
		err := svc.VerifyCode(context.Background(), testAddress, "123456", testIP, accept(new(bool)))
		assert.Error(t, err)
		assert.NotErrorIs(t, err, ErrCodeRejected)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
	t.Run("cannot mark the code used", func(t *testing.T) {
		svc, mock, _ := newCodeServiceWithMock(t)
		expectCodeRow(mock, svc, "123456", 0)
		mock.ExpectExec(q(countAttemptSQL)).WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectExec(q(markUsedSQL)).WillReturnError(errors.New("db down"))
		mock.ExpectRollback()
		assert.Error(t, svc.VerifyCode(context.Background(), testAddress, "123456", testIP, accept(new(bool))), "a code that cannot be used up does not sign anyone in")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

var _ email.Mailer = (*codeMailer)(nil)
