package middleware

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"server/internal/migrations"
	"server/internal/services/auth"

	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Runs the middleware with the real session service and SQL.
// Set MIGRATION_TEST_DSN only for a disposable PostgreSQL database.
func TestRequireAuthIntegration(t *testing.T) {
	dsn := os.Getenv("MIGRATION_TEST_DSN")
	if dsn == "" {
		t.Skip("requires a disposable PostgreSQL database via MIGRATION_TEST_DSN")
	}
	admin, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	defer admin.Close()

	schema := fmt.Sprintf("middleware_test_%d", time.Now().UnixNano())
	_, err = admin.Exec("CREATE SCHEMA " + pq.QuoteIdentifier(schema))
	require.NoError(t, err)
	defer admin.Exec("DROP SCHEMA " + pq.QuoteIdentifier(schema) + " CASCADE")
	db, err := sql.Open("postgres", dsn+" search_path="+schema)
	require.NoError(t, err)
	defer db.Close()
	require.NoError(t, migrations.Up(context.Background(), db))

	sessions := auth.NewSessionService(db)
	ctx := context.Background()
	var caregiver int
	require.NoError(t, db.QueryRow(`INSERT INTO "Caregiver" (name) VALUES ('Middleware test') RETURNING id`).Scan(&caregiver))

	// age moves a session back in time, as if it was created and last used that long ago
	age := func(t *testing.T, token string, created, lastSeen, expiresIn time.Duration) {
		t.Helper()
		_, err := db.Exec(`UPDATE "Session" SET created_at = now() - $2::interval, last_seen_at = now() - $3::interval,
			expires_at = now() + $4::interval WHERE token_hash = $1`,
			hashOf(token), created.String(), lastSeen.String(), expiresIn.String())
		require.NoError(t, err)
	}
	newSession := func(t *testing.T) string {
		t.Helper()
		token, err := sessions.Create(ctx, caregiver)
		require.NoError(t, err)
		return token
	}

	t.Run("a session used this week stays as it is", func(t *testing.T) {
		token := newSession(t)

		w, reached, userID := serve(sessions, sessionCookieFor(token))

		assert.Equal(t, http.StatusOK, w.Code)
		assert.True(t, reached)
		assert.Equal(t, caregiver, userID)
		assert.Empty(t, w.Result().Cookies())
	})

	t.Run("a session unused for two days renews and the browser gets 90 more days", func(t *testing.T) {
		token := newSession(t)
		age(t, token, 40*24*time.Hour, 2*24*time.Hour, 50*24*time.Hour)

		w, reached, _ := serve(sessions, sessionCookieFor(token))

		assert.True(t, reached)
		cookies := w.Result().Cookies()
		require.Len(t, cookies, 1)
		assert.Equal(t, token, cookies[0].Value)
		assert.InDelta(t, auth.SessionIdleTTL.Seconds(), cookies[0].MaxAge, 60)

		again, _, _ := serve(sessions, sessionCookieFor(token))
		assert.Empty(t, again.Result().Cookies(), "used again straight away, so no second cookie")
	})

	t.Run("near the one year cutoff the cookie is shortened to match", func(t *testing.T) {
		token := newSession(t)
		age(t, token, 300*24*time.Hour, 2*24*time.Hour, 10*24*time.Hour)

		w, reached, _ := serve(sessions, sessionCookieFor(token))

		assert.True(t, reached)
		cookies := w.Result().Cookies()
		require.Len(t, cookies, 1)
		assert.InDelta(t, (65 * 24 * time.Hour).Seconds(), cookies[0].MaxAge, 60)
	})

	t.Run("an expired session is denied and its cookie deleted", func(t *testing.T) {
		token := newSession(t)
		age(t, token, 200*24*time.Hour, 100*24*time.Hour, -time.Hour)

		w, reached, _ := serve(sessions, sessionCookieFor(token))

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.False(t, reached)
		cookies := w.Result().Cookies()
		require.Len(t, cookies, 1)
		assert.Less(t, cookies[0].MaxAge, 0)
	})

	t.Run("a session past the one year cutoff is denied even though it is in use", func(t *testing.T) {
		token := newSession(t)
		age(t, token, 366*24*time.Hour, time.Hour, 80*24*time.Hour)

		w, reached, _ := serve(sessions, sessionCookieFor(token))

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.False(t, reached)
	})

	t.Run("a made-up token is denied", func(t *testing.T) {
		w, reached, _ := serve(sessions, sessionCookieFor("made-up"))

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.False(t, reached)
	})

	t.Run("a database that is down is a 500, not a logout", func(t *testing.T) {
		token := newSession(t)
		closed, err := sql.Open("postgres", dsn+" search_path="+schema)
		require.NoError(t, err)
		require.NoError(t, closed.Close())

		w, reached, _ := serve(auth.NewSessionService(closed), sessionCookieFor(token))

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.False(t, reached)
		assert.Empty(t, w.Result().Cookies())
	})
}

// hashOf is how the session service stores a token: the hex SHA-256 of it
func hashOf(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
