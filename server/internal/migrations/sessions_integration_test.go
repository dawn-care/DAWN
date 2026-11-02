package migrations

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

// Set MIGRATION_TEST_DSN only for a disposable PostgreSQL database.
func TestLongLivedSessionsMigrationIntegration(t *testing.T) {
	dsn := os.Getenv("MIGRATION_TEST_DSN")
	if dsn == "" {
		t.Skip("requires a disposable PostgreSQL database via MIGRATION_TEST_DSN")
	}
	admin, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	defer admin.Close()

	schema := fmt.Sprintf("sessions_test_%d", time.Now().UnixNano())
	_, err = admin.Exec("CREATE SCHEMA " + pq.QuoteIdentifier(schema))
	require.NoError(t, err)
	defer admin.Exec("DROP SCHEMA " + pq.QuoteIdentifier(schema) + " CASCADE")
	db, err := sql.Open("postgres", dsn+" search_path="+schema)
	require.NoError(t, err)
	defer db.Close()
	require.NoError(t, Up(context.Background(), db))

	script, err := files.ReadFile("sql/00008_long_lived_sessions.sql")
	require.NoError(t, err)
	upAndDown := strings.Split(string(script), "-- +goose Down")
	require.Len(t, upAndDown, 2)
	up := strings.SplitN(upAndDown[0], "-- +goose Up", 2)[1]
	down := upAndDown[1]

	t.Run("new shape", func(t *testing.T) {
		requireNewSessionShape(t, db)
	})

	t.Run("deleting a caregiver deletes only their sessions", func(t *testing.T) {
		gone, kept := newCaregiver(t, db), newCaregiver(t, db)
		for _, c := range []struct {
			hash string
			id   int
		}{{"hash-gone-1", gone}, {"hash-gone-2", gone}, {"hash-kept", kept}} {
			_, err := db.Exec(`INSERT INTO "Session" (token_hash, user_id, expires_at) VALUES ($1, $2, now() + interval '90 days')`, c.hash, c.id)
			require.NoError(t, err)
		}
		_, err := db.Exec(`DELETE FROM "Caregiver" WHERE id = $1`, gone)
		require.NoError(t, err)
		require.Equal(t, 0, sessionsOf(t, db, gone))
		require.Equal(t, 1, sessionsOf(t, db, kept))
	})

	t.Run("rejects a session for a caregiver that does not exist", func(t *testing.T) {
		_, err := db.Exec(`INSERT INTO "Session" (token_hash, user_id, expires_at) VALUES ('hash-orphan', -1, now())`)
		require.Error(t, err)
	})

	t.Run("rejects a repeated token hash", func(t *testing.T) {
		id := newCaregiver(t, db)
		_, err := db.Exec(`INSERT INTO "Session" (token_hash, user_id, expires_at) VALUES ('hash-dup', $1, now())`, id)
		require.NoError(t, err)
		_, err = db.Exec(`INSERT INTO "Session" (token_hash, user_id, expires_at) VALUES ('hash-dup', $1, now())`, id)
		require.Error(t, err)
	})

	t.Run("last_seen_at starts at now and created_at is required", func(t *testing.T) {
		id := newCaregiver(t, db)
		var lastSeen time.Time
		require.NoError(t, db.QueryRow(`INSERT INTO "Session" (token_hash, user_id, expires_at) VALUES ('hash-seen', $1, now())
			RETURNING last_seen_at`, id).Scan(&lastSeen))
		require.WithinDuration(t, time.Now(), lastSeen, time.Minute)
		_, err := db.Exec(`INSERT INTO "Session" (token_hash, user_id, created_at, expires_at) VALUES ('hash-null', $1, NULL, now())`, id)
		require.Error(t, err)
	})

	t.Run("down restores the old shape and up deletes the old sessions", func(t *testing.T) {
		_, err := db.Exec(down)
		require.NoError(t, err)
		var oldColumns int
		require.NoError(t, db.QueryRow(`SELECT count(*) FROM information_schema.columns
			WHERE table_schema = $1 AND table_name = 'Session' AND column_name IN ('session_id')`, schema).Scan(&oldColumns))
		require.Equal(t, 1, oldColumns)
		var removed int
		require.NoError(t, db.QueryRow(`SELECT count(*) FROM information_schema.columns
			WHERE table_schema = $1 AND table_name = 'Session' AND column_name IN ('token_hash', 'last_seen_at')`, schema).Scan(&removed))
		require.Equal(t, 0, removed)

		// The old table held raw tokens and allowed sessions for any user_id, even one without a caregiver
		id := newCaregiver(t, db)
		for _, c := range []struct {
			token string
			user  int
		}{{"raw-token-1", id}, {"raw-token-2", id}, {"raw-token-orphan", -1}} {
			_, err := db.Exec(`INSERT INTO "Session" (session_id, user_id, expires_at) VALUES ($1, $2, now() + interval '7 days')`, c.token, c.user)
			require.NoError(t, err)
		}

		_, err = db.Exec(up)
		require.NoError(t, err)
		var left int
		require.NoError(t, db.QueryRow(`SELECT count(*) FROM "Session"`).Scan(&left))
		require.Equal(t, 0, left, "old sessions hold raw tokens, so they are deleted")
		requireNewSessionShape(t, db)
	})
}

func requireNewSessionShape(t *testing.T, db *sql.DB) {
	t.Helper()
	rows, err := db.Query(`SELECT column_name, data_type, is_nullable FROM information_schema.columns
		WHERE table_schema = current_schema() AND table_name = 'Session' ORDER BY column_name`)
	require.NoError(t, err)
	defer rows.Close()
	got := map[string]string{}
	for rows.Next() {
		var name, dataType, nullable string
		require.NoError(t, rows.Scan(&name, &dataType, &nullable))
		got[name] = dataType + "/" + nullable
	}
	require.NoError(t, rows.Err())
	require.Equal(t, map[string]string{
		"id":           "integer/NO",
		"token_hash":   "text/NO",
		"user_id":      "integer/NO",
		"created_at":   "timestamp with time zone/NO",
		"expires_at":   "timestamp with time zone/NO",
		"last_seen_at": "timestamp with time zone/NO",
	}, got)

	var indexes int
	require.NoError(t, db.QueryRow(`SELECT count(*) FROM pg_indexes WHERE schemaname = current_schema()
		AND tablename = 'Session' AND indexname IN ('Session_user_id_idx', 'Session_token_hash_key')`).Scan(&indexes))
	require.Equal(t, 2, indexes, "lookups by user and by token hash need their indexes")
}

func newCaregiver(t *testing.T, db *sql.DB) int {
	t.Helper()
	var id int
	require.NoError(t, db.QueryRow(`INSERT INTO "Caregiver" (name) VALUES ('Session test') RETURNING id`).Scan(&id))
	return id
}

func sessionsOf(t *testing.T, db *sql.DB, userID int) int {
	t.Helper()
	var n int
	require.NoError(t, db.QueryRow(`SELECT count(*) FROM "Session" WHERE user_id = $1`, userID).Scan(&n))
	return n
}
