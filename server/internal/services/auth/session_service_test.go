package auth

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

/* -------------------------------------------------------
   SETUP
------------------------------------------------------- */

var testNow = time.Date(2026, time.November, 1, 12, 0, 0, 0, time.UTC)

func setupSessionService(t *testing.T) (*sessionService, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock database: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	svc := NewSessionService(db)
	svc.now = func() time.Time { return testNow }
	return svc, mock
}

// captureArg records the value a query was called with, so a test can look at it afterwards
type captureArg struct{ value driver.Value }

func (c *captureArg) Match(v driver.Value) bool {
	c.value = v
	return true
}

func sessionRow(userID int, createdAt, expiresAt, lastSeenAt time.Time) *sqlmock.Rows {
	return sqlmock.NewRows([]string{"user_id", "created_at", "expires_at", "last_seen_at"}).
		AddRow(userID, createdAt, expiresAt, lastSeenAt)
}

// The whole statements, so a test fails if a WHERE clause or comparison changes. sqlmock squashes
// whitespace before matching, so the code's line breaks don't matter.
const (
	cleanupSessions   = `DELETE FROM "Session" WHERE user_id = $1 AND (expires_at < $2 OR created_at < $3)`
	insertSession     = `INSERT INTO "Session" (token_hash, user_id, created_at, expires_at, last_seen_at) VALUES ($1, $2, $3, $4, $3)`
	selectSession     = `SELECT user_id, created_at, expires_at, last_seen_at FROM "Session" WHERE token_hash = $1`
	updateSession     = `UPDATE "Session" SET expires_at = $2, last_seen_at = $3 WHERE token_hash = $1`
	deleteSession     = `DELETE FROM "Session" WHERE token_hash = $1`
	deleteAllSessions = `DELETE FROM "Session" WHERE user_id = $1`
)

// exactly matches the whole statement, not just a part of it
func exactly(statement string) string {
	return "^" + regexp.QuoteMeta(statement) + "$"
}

func days(n int) time.Duration { return time.Duration(n) * 24 * time.Hour }

/* -------------------------------------------------------
   HASHING
------------------------------------------------------- */

func TestHashToken(t *testing.T) {
	hash := hashToken("token")
	assert.Equal(t, hash, hashToken("token"))
	assert.NotEqual(t, hash, hashToken("token2"))
	assert.Len(t, hash, 64)
	assert.NotContains(t, hash, "token")
	// SHA-256 of "abc", so the hash is the standard one a database tool could also compute
	assert.Equal(t, "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad", hashToken("abc"))
}

func TestGenerateSessionToken_IsRandomAndLongEnough(t *testing.T) {
	a, err := generateSessionToken()
	require.NoError(t, err)
	b, err := generateSessionToken()
	require.NoError(t, err)
	assert.NotEqual(t, a, b)
	assert.GreaterOrEqual(t, len(a), 43) // 32 bytes, base64
}

/* -------------------------------------------------------
   CREATE
------------------------------------------------------- */

func TestCreate_StoresOnlyTheHash(t *testing.T) {
	svc, mock := setupSessionService(t)

	mock.ExpectExec(exactly(cleanupSessions)).
		WithArgs(7, testNow, testNow.Add(-SessionAbsoluteTTL)).
		WillReturnResult(sqlmock.NewResult(0, 0))
	stored := &captureArg{}
	mock.ExpectExec(exactly(insertSession)).
		WithArgs(stored, 7, testNow, testNow.Add(90*24*time.Hour)).
		WillReturnResult(sqlmock.NewResult(1, 1))

	token, err := svc.Create(context.Background(), 7)

	require.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.Equal(t, hashToken(token), stored.value, "the database gets the hash of the token")
	assert.NotEqual(t, token, stored.value, "the database never gets the token itself")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreate_InsertError(t *testing.T) {
	svc, mock := setupSessionService(t)

	mock.ExpectExec(exactly(cleanupSessions)).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(exactly(insertSession)).WillReturnError(errors.New("db down"))

	token, err := svc.Create(context.Background(), 7)

	assert.Error(t, err)
	assert.Empty(t, token)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreate_CleanupFailureDoesNotBlockLogin(t *testing.T) {
	svc, mock := setupSessionService(t)

	mock.ExpectExec(exactly(cleanupSessions)).WillReturnError(errors.New("db hiccup"))
	mock.ExpectExec(exactly(insertSession)).WillReturnResult(sqlmock.NewResult(1, 1))

	token, err := svc.Create(context.Background(), 7)

	require.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.NoError(t, mock.ExpectationsWereMet())
}

/* -------------------------------------------------------
   VALIDATE
------------------------------------------------------- */

func TestValidate_LooksUpByHash(t *testing.T) {
	svc, mock := setupSessionService(t)

	mock.ExpectQuery(exactly(selectSession)).
		WithArgs(hashToken("cookie-token")).
		WillReturnRows(sessionRow(7, testNow.Add(-days(1)), testNow.Add(days(89)), testNow.Add(-time.Hour)))

	got, err := svc.Validate(context.Background(), "cookie-token")

	require.NoError(t, err)
	assert.Equal(t, 7, got.UserID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestValidate_UnknownToken(t *testing.T) {
	svc, mock := setupSessionService(t)

	mock.ExpectQuery(exactly(selectSession)).WillReturnError(sql.ErrNoRows)

	_, err := svc.Validate(context.Background(), "nope")

	assert.ErrorIs(t, err, ErrSessionNotFound)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestValidate_DatabaseErrorIsNotReportedAsUnknownToken(t *testing.T) {
	svc, mock := setupSessionService(t)

	mock.ExpectQuery(exactly(selectSession)).WillReturnError(errors.New("db down"))

	_, err := svc.Validate(context.Background(), "token")

	assert.Error(t, err)
	assert.NotErrorIs(t, err, ErrSessionNotFound)
	assert.NotErrorIs(t, err, ErrSessionExpired)
}

func TestValidate_UsedRecentlyIsNotWrittenAgain(t *testing.T) {
	svc, mock := setupSessionService(t)
	expires := testNow.Add(days(40))

	// Last renewed 23h59m ago: under a day, so no UPDATE (sqlmock fails on one nobody expected)
	mock.ExpectQuery(exactly(selectSession)).
		WillReturnRows(sessionRow(7, testNow.Add(-days(50)), expires, testNow.Add(-24*time.Hour+time.Minute)))

	got, err := svc.Validate(context.Background(), "token")

	require.NoError(t, err)
	assert.False(t, got.Refreshed)
	assert.Equal(t, expires, got.ExpiresAt)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestValidate_RenewsAfterADay(t *testing.T) {
	svc, mock := setupSessionService(t)
	created := testNow.Add(-days(50))
	oldExpiry := testNow.Add(days(40))

	mock.ExpectQuery(exactly(selectSession)).
		WillReturnRows(sessionRow(7, created, oldExpiry, testNow.Add(-24*time.Hour)))
	mock.ExpectExec(exactly(updateSession)).
		WithArgs(hashToken("token"), testNow.Add(SessionIdleTTL), testNow).
		WillReturnResult(sqlmock.NewResult(0, 1))

	got, err := svc.Validate(context.Background(), "token")

	require.NoError(t, err)
	assert.True(t, got.Refreshed)
	assert.Equal(t, testNow.Add(SessionIdleTTL), got.ExpiresAt)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestValidate_RenewalStopsAtTheOneYearCutoff(t *testing.T) {
	svc, mock := setupSessionService(t)
	created := testNow.Add(-days(300))
	cutoff := created.Add(SessionAbsoluteTTL) // 65 days from now, sooner than 90

	mock.ExpectQuery(exactly(selectSession)).
		WillReturnRows(sessionRow(7, created, testNow.Add(days(10)), testNow.Add(-days(2))))
	mock.ExpectExec(exactly(updateSession)).
		WithArgs(hashToken("token"), cutoff, testNow).
		WillReturnResult(sqlmock.NewResult(0, 1))

	got, err := svc.Validate(context.Background(), "token")

	require.NoError(t, err)
	assert.True(t, got.Refreshed)
	assert.Equal(t, cutoff, got.ExpiresAt)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestValidate_AtTheCutoffNothingMovesSoNoNewCookie(t *testing.T) {
	svc, mock := setupSessionService(t)
	created := testNow.Add(-days(364))
	cutoff := created.Add(SessionAbsoluteTTL) // tomorrow, and expires_at is already there

	mock.ExpectQuery(exactly(selectSession)).
		WillReturnRows(sessionRow(7, created, cutoff, testNow.Add(-days(2))))
	// last_seen_at is still written, or every request from here on would hit the database
	mock.ExpectExec(exactly(updateSession)).
		WithArgs(hashToken("token"), cutoff, testNow).
		WillReturnResult(sqlmock.NewResult(0, 1))

	got, err := svc.Validate(context.Background(), "token")

	require.NoError(t, err)
	assert.False(t, got.Refreshed)
	assert.Equal(t, cutoff, got.ExpiresAt)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestValidate_FailedRenewalKeepsTheCaregiverLoggedIn(t *testing.T) {
	svc, mock := setupSessionService(t)
	oldExpiry := testNow.Add(days(40))

	mock.ExpectQuery(exactly(selectSession)).
		WillReturnRows(sessionRow(7, testNow.Add(-days(50)), oldExpiry, testNow.Add(-days(2))))
	mock.ExpectExec(exactly(updateSession)).WillReturnError(errors.New("db hiccup"))

	got, err := svc.Validate(context.Background(), "token")

	require.NoError(t, err)
	assert.Equal(t, 7, got.UserID)
	assert.False(t, got.Refreshed)
	assert.Equal(t, oldExpiry, got.ExpiresAt)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// A logout on another device can delete the session between the lookup and the renewal. It no
// longer exists, so it must not be served or get a new cookie.
func TestValidate_SessionEndedDuringRenewalIsNotFound(t *testing.T) {
	svc, mock := setupSessionService(t)

	mock.ExpectQuery(exactly(selectSession)).
		WillReturnRows(sessionRow(7, testNow.Add(-days(50)), testNow.Add(days(40)), testNow.Add(-days(2))))
	mock.ExpectExec(exactly(updateSession)).WillReturnResult(sqlmock.NewResult(0, 0))

	got, err := svc.Validate(context.Background(), "token")

	assert.ErrorIs(t, err, ErrSessionNotFound)
	assert.False(t, got.Refreshed)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestValidate_IdleExpiredSessionIsDeleted(t *testing.T) {
	svc, mock := setupSessionService(t)

	mock.ExpectQuery(exactly(selectSession)).
		WillReturnRows(sessionRow(7, testNow.Add(-days(200)), testNow.Add(-time.Second), testNow.Add(-days(90))))
	mock.ExpectExec(exactly(deleteSession)).
		WithArgs(hashToken("token")).
		WillReturnResult(sqlmock.NewResult(0, 1))

	_, err := svc.Validate(context.Background(), "token")

	assert.ErrorIs(t, err, ErrSessionExpired)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestValidate_PastTheOneYearCutoffIsDeletedEvenIfNotIdle(t *testing.T) {
	svc, mock := setupSessionService(t)

	// Used yesterday and expires_at is in the future, but it was created 366 days ago
	mock.ExpectQuery(exactly(selectSession)).
		WillReturnRows(sessionRow(7, testNow.Add(-days(366)), testNow.Add(days(80)), testNow.Add(-days(1))))
	mock.ExpectExec(exactly(deleteSession)).
		WithArgs(hashToken("token")).
		WillReturnResult(sqlmock.NewResult(0, 1))

	_, err := svc.Validate(context.Background(), "token")

	assert.ErrorIs(t, err, ErrSessionExpired)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestValidate_ExpiredStaysExpiredWhenDeleteFails(t *testing.T) {
	svc, mock := setupSessionService(t)

	mock.ExpectQuery(exactly(selectSession)).
		WillReturnRows(sessionRow(7, testNow.Add(-days(200)), testNow.Add(-time.Second), testNow.Add(-days(90))))
	mock.ExpectExec(exactly(deleteSession)).WillReturnError(errors.New("db hiccup"))

	_, err := svc.Validate(context.Background(), "token")

	assert.ErrorIs(t, err, ErrSessionExpired)
	assert.NoError(t, mock.ExpectationsWereMet(), "it still tries to delete the expired row")
}

// A session is still valid at the exact instant it expires and ends just after, for both the 90
// day window and the 1 year cutoff. Create's cleanup (expires_at < now) agrees with this.
func TestValidate_EndsJustAfterItsExpiryInstant(t *testing.T) {
	cases := []struct {
		name               string
		createdAt, expires time.Time
		valid              bool
	}{
		{"idle window, at the instant", testNow.Add(-days(100)), testNow, true},
		{"idle window, just after", testNow.Add(-days(100)), testNow.Add(-time.Nanosecond), false},
		{"1 year cutoff, at the instant", testNow.Add(-SessionAbsoluteTTL), testNow.Add(days(30)), true},
		{"1 year cutoff, just after", testNow.Add(-SessionAbsoluteTTL - time.Nanosecond), testNow.Add(days(30)), false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			svc, mock := setupSessionService(t)
			// Renewed an hour ago, so a valid session needs no UPDATE
			mock.ExpectQuery(exactly(selectSession)).
				WillReturnRows(sessionRow(7, c.createdAt, c.expires, testNow.Add(-time.Hour)))
			if !c.valid {
				mock.ExpectExec(exactly(deleteSession)).WillReturnResult(sqlmock.NewResult(0, 1))
			}

			_, err := svc.Validate(context.Background(), "token")

			if c.valid {
				assert.NoError(t, err)
			} else {
				assert.ErrorIs(t, err, ErrSessionExpired)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

/* -------------------------------------------------------
   FIND
------------------------------------------------------- */

func TestFind_DoesNotRenew(t *testing.T) {
	svc, mock := setupSessionService(t)
	expires := testNow.Add(days(40))

	// Last renewed two days ago, which Validate would renew; Find writes nothing
	mock.ExpectQuery(exactly(selectSession)).
		WithArgs(hashToken("cookie-token")).
		WillReturnRows(sessionRow(7, testNow.Add(-days(50)), expires, testNow.Add(-days(2))))

	got, err := svc.Find(context.Background(), "cookie-token")

	require.NoError(t, err)
	assert.Equal(t, 7, got.UserID)
	assert.Equal(t, expires, got.ExpiresAt)
	assert.False(t, got.Refreshed)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestFind_UnknownAndExpiredAreReportedLikeValidate(t *testing.T) {
	svc, mock := setupSessionService(t)

	mock.ExpectQuery(exactly(selectSession)).WillReturnError(sql.ErrNoRows)
	_, err := svc.Find(context.Background(), "nope")
	assert.ErrorIs(t, err, ErrSessionNotFound)

	mock.ExpectQuery(exactly(selectSession)).
		WillReturnRows(sessionRow(7, testNow.Add(-days(200)), testNow.Add(-time.Second), testNow.Add(-days(90))))
	mock.ExpectExec(exactly(deleteSession)).WithArgs(hashToken("old")).WillReturnResult(sqlmock.NewResult(0, 1))
	_, err = svc.Find(context.Background(), "old")
	assert.ErrorIs(t, err, ErrSessionExpired)

	assert.NoError(t, mock.ExpectationsWereMet())
}

/* -------------------------------------------------------
   DELETE ALL FOR USER
------------------------------------------------------- */

func TestDeleteAllForUser_Success(t *testing.T) {
	svc, mock := setupSessionService(t)

	mock.ExpectExec(exactly(deleteAllSessions)).
		WithArgs(7).
		WillReturnResult(sqlmock.NewResult(0, 3))

	assert.NoError(t, svc.DeleteAllForUser(context.Background(), 7))
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteAllForUser_Error(t *testing.T) {
	svc, mock := setupSessionService(t)

	mock.ExpectExec(exactly(deleteAllSessions)).
		WillReturnError(errors.New("db down"))

	assert.Error(t, svc.DeleteAllForUser(context.Background(), 7))
}
