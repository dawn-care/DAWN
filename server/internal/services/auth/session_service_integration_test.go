package auth

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"server/internal/migrations"
)

// Runs the real SQL against PostgreSQL. Set MIGRATION_TEST_DSN only for a disposable database.
func TestSessionServiceIntegration(t *testing.T) {
	dsn := os.Getenv("MIGRATION_TEST_DSN")
	if dsn == "" {
		t.Skip("requires a disposable PostgreSQL database via MIGRATION_TEST_DSN")
	}
	admin, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	defer admin.Close()

	schema := fmt.Sprintf("session_svc_test_%d", time.Now().UnixNano())
	_, err = admin.Exec("CREATE SCHEMA " + pq.QuoteIdentifier(schema))
	require.NoError(t, err)
	defer admin.Exec("DROP SCHEMA " + pq.QuoteIdentifier(schema) + " CASCADE")
	db, err := sql.Open("postgres", dsn+" search_path="+schema)
	require.NoError(t, err)
	defer db.Close()
	require.NoError(t, migrations.Up(context.Background(), db))

	ctx := context.Background()
	svc := NewSessionService(db)

	newCaregiver := func(t *testing.T) int {
		t.Helper()
		var id int
		require.NoError(t, db.QueryRow(`INSERT INTO "Caregiver" (name) VALUES ('Session test') RETURNING id`).Scan(&id))
		return id
	}
	// Moves a session back in time, as if it had been created and last used that long ago
	age := func(t *testing.T, token string, created, lastSeen time.Duration, expiresIn time.Duration) {
		t.Helper()
		_, err := db.Exec(`UPDATE "Session" SET created_at = now() - $2::interval, last_seen_at = now() - $3::interval,
			expires_at = now() + $4::interval WHERE token_hash = $1`,
			hashToken(token), created.String(), lastSeen.String(), expiresIn.String())
		require.NoError(t, err)
	}
	row := func(t *testing.T, token string) (expiresAt, lastSeenAt time.Time, found bool) {
		t.Helper()
		err := db.QueryRow(`SELECT expires_at, last_seen_at FROM "Session" WHERE token_hash = $1`, hashToken(token)).Scan(&expiresAt, &lastSeenAt)
		if err == sql.ErrNoRows {
			return time.Time{}, time.Time{}, false
		}
		require.NoError(t, err)
		return expiresAt, lastSeenAt, true
	}
	count := func(t *testing.T, userID int) (n int) {
		t.Helper()
		require.NoError(t, db.QueryRow(`SELECT count(*) FROM "Session" WHERE user_id = $1`, userID).Scan(&n))
		return n
	}

	t.Run("create stores the hash and validate accepts the token", func(t *testing.T) {
		user := newCaregiver(t)
		token, err := svc.Create(ctx, user)
		require.NoError(t, err)

		var raw int
		require.NoError(t, db.QueryRow(`SELECT count(*) FROM "Session" WHERE token_hash = $1`, token).Scan(&raw))
		assert.Equal(t, 0, raw, "the raw token is not stored")
		expires, _, found := row(t, token)
		require.True(t, found, "the hash is stored")
		assert.WithinDuration(t, time.Now().Add(SessionIdleTTL), expires, time.Minute)

		got, err := svc.Validate(ctx, token)
		require.NoError(t, err)
		assert.Equal(t, user, got.UserID)
		assert.False(t, got.Refreshed, "a new session is not renewed straight away")

		_, err = svc.Validate(ctx, "not-a-real-token")
		assert.ErrorIs(t, err, ErrSessionNotFound)
		_, err = svc.Validate(ctx, hashToken(token))
		assert.ErrorIs(t, err, ErrSessionNotFound, "the stored hash is not itself a login")
	})

	t.Run("a session used more than a day ago renews to 90 days", func(t *testing.T) {
		user := newCaregiver(t)
		token, err := svc.Create(ctx, user)
		require.NoError(t, err)
		age(t, token, 40*24*time.Hour, 2*24*time.Hour, 50*24*time.Hour)

		got, err := svc.Validate(ctx, token)
		require.NoError(t, err)
		assert.True(t, got.Refreshed)
		assert.WithinDuration(t, time.Now().Add(SessionIdleTTL), got.ExpiresAt, time.Minute)
		expires, lastSeen, _ := row(t, token)
		assert.WithinDuration(t, time.Now().Add(SessionIdleTTL), expires, time.Minute, "saved in the database")
		assert.WithinDuration(t, time.Now(), lastSeen, time.Minute)

		again, err := svc.Validate(ctx, token)
		require.NoError(t, err)
		assert.False(t, again.Refreshed, "used again straight away, so no second renewal")
	})

	t.Run("renewal never goes past one year after creation", func(t *testing.T) {
		user := newCaregiver(t)
		token, err := svc.Create(ctx, user)
		require.NoError(t, err)
		age(t, token, 300*24*time.Hour, 2*24*time.Hour, 10*24*time.Hour)

		got, err := svc.Validate(ctx, token)
		require.NoError(t, err)
		assert.WithinDuration(t, time.Now().Add(65*24*time.Hour), got.ExpiresAt, time.Minute)
	})

	t.Run("idle for over 90 days ends the session and removes the row", func(t *testing.T) {
		user := newCaregiver(t)
		token, err := svc.Create(ctx, user)
		require.NoError(t, err)
		_, err = db.Exec(`UPDATE "Session" SET expires_at = now() - interval '1 hour' WHERE token_hash = $1`, hashToken(token))
		require.NoError(t, err)

		_, err = svc.Validate(ctx, token)
		assert.ErrorIs(t, err, ErrSessionExpired)
		_, _, found := row(t, token)
		assert.False(t, found)
	})

	t.Run("366 days old ends the session even though it is in use", func(t *testing.T) {
		user := newCaregiver(t)
		token, err := svc.Create(ctx, user)
		require.NoError(t, err)
		age(t, token, 366*24*time.Hour, time.Hour, 80*24*time.Hour)

		_, err = svc.Validate(ctx, token)
		assert.ErrorIs(t, err, ErrSessionExpired)
		_, _, found := row(t, token)
		assert.False(t, found)
	})

	t.Run("delete all ends every session of one caregiver only", func(t *testing.T) {
		user, other := newCaregiver(t), newCaregiver(t)
		first, err := svc.Create(ctx, user)
		require.NoError(t, err)
		second, err := svc.Create(ctx, user)
		require.NoError(t, err)
		otherToken, err := svc.Create(ctx, other)
		require.NoError(t, err)

		require.NoError(t, svc.DeleteAllForUser(ctx, user))

		for _, token := range []string{first, second} {
			_, err := svc.Validate(ctx, token)
			assert.ErrorIs(t, err, ErrSessionNotFound)
		}
		_, err = svc.Validate(ctx, otherToken)
		assert.NoError(t, err)
		assert.Equal(t, 0, count(t, user))
		assert.Equal(t, 1, count(t, other))
	})

	t.Run("create removes only that caregiver's dead sessions", func(t *testing.T) {
		user, other := newCaregiver(t), newCaregiver(t)
		idle, err := svc.Create(ctx, user)
		require.NoError(t, err)
		old, err := svc.Create(ctx, user)
		require.NoError(t, err)
		live, err := svc.Create(ctx, user)
		require.NoError(t, err)
		otherDead, err := svc.Create(ctx, other)
		require.NoError(t, err)
		age(t, idle, 200*24*time.Hour, 100*24*time.Hour, -time.Hour)
		age(t, old, 400*24*time.Hour, time.Hour, 50*24*time.Hour)
		age(t, otherDead, 200*24*time.Hour, 100*24*time.Hour, -time.Hour)

		_, err = svc.Create(ctx, user)
		require.NoError(t, err)

		_, _, found := row(t, idle)
		assert.False(t, found, "idle-expired session is cleaned up")
		_, _, found = row(t, old)
		assert.False(t, found, "session past the cutoff is cleaned up")
		_, _, found = row(t, live)
		assert.True(t, found, "a live session stays")
		_, _, found = row(t, otherDead)
		assert.True(t, found, "another caregiver's sessions are left alone")
	})

	t.Run("deleting the caregiver ends their sessions", func(t *testing.T) {
		user := newCaregiver(t)
		token, err := svc.Create(ctx, user)
		require.NoError(t, err)

		_, err = db.Exec(`DELETE FROM "Caregiver" WHERE id = $1`, user)
		require.NoError(t, err)

		_, err = svc.Validate(ctx, token)
		assert.ErrorIs(t, err, ErrSessionNotFound)
	})

	t.Run("create fails for a caregiver that does not exist", func(t *testing.T) {
		_, err := svc.Create(ctx, -1)
		assert.Error(t, err)
	})
}
