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
func TestEmailLoginCodesMigrationIntegration(t *testing.T) {
	dsn := os.Getenv("MIGRATION_TEST_DSN")
	if dsn == "" {
		t.Skip("requires a disposable PostgreSQL database via MIGRATION_TEST_DSN")
	}
	admin, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	defer admin.Close()

	schema := fmt.Sprintf("email_login_test_%d", time.Now().UnixNano())
	_, err = admin.Exec("CREATE SCHEMA " + pq.QuoteIdentifier(schema))
	require.NoError(t, err)
	defer admin.Exec("DROP SCHEMA " + pq.QuoteIdentifier(schema) + " CASCADE")
	db, err := sql.Open("postgres", dsn+" search_path="+schema)
	require.NoError(t, err)
	defer db.Close()
	require.NoError(t, Up(context.Background(), db))

	script, err := files.ReadFile("sql/00010_email_login_codes.sql")
	require.NoError(t, err)
	upAndDown := strings.Split(string(script), "-- +goose Down")
	require.Len(t, upAndDown, 2)
	up := strings.SplitN(upAndDown[0], "-- +goose Up", 2)[1]
	down := upAndDown[1]

	t.Run("new shape", func(t *testing.T) {
		requireEmailLoginShape(t, db)
	})

	t.Run("the demo caregiver's email moved out of google_id", func(t *testing.T) {
		var email, googleID sql.NullString
		require.NoError(t, db.QueryRow(`SELECT email, google_id FROM "Caregiver" WHERE name = 'Jane Smith'`).Scan(&email, &googleID))
		require.Equal(t, "test.caregiver@example.com", email.String)
		require.False(t, googleID.Valid)
	})

	t.Run("email is optional and a Google caregiver keeps its google_id", func(t *testing.T) {
		_, err := db.Exec(`INSERT INTO "Caregiver" (name, google_id) VALUES ('No email one', 'g-1'), ('No email two', 'g-2')`)
		require.NoError(t, err)
	})

	t.Run("email is unique ignoring case", func(t *testing.T) {
		_, err := db.Exec(`INSERT INTO "Caregiver" (name, email) VALUES ('Unique', 'Person@Example.com')`)
		require.NoError(t, err)
		_, err = db.Exec(`INSERT INTO "Caregiver" (name, email) VALUES ('Duplicate', 'person@example.COM')`)
		require.Error(t, err)
	})

	t.Run("a login code starts unused with no attempts", func(t *testing.T) {
		var attempts int
		var usedAt sql.NullTime
		var createdAt time.Time
		require.NoError(t, db.QueryRow(`INSERT INTO "LoginCode" (email, code_hash, expires_at)
			VALUES ('person@example.com', 'hash', now() + interval '10 minutes')
			RETURNING attempts, used_at, created_at`).Scan(&attempts, &usedAt, &createdAt))
		require.Equal(t, 0, attempts)
		require.False(t, usedAt.Valid)
		require.WithinDuration(t, time.Now(), createdAt, time.Minute)
		_, err := db.Exec(`INSERT INTO "LoginCode" (email, code_hash) VALUES ('person@example.com', 'hash')`)
		require.Error(t, err, "expires_at is required")
	})

	t.Run("a request log row needs an email and an ip hash", func(t *testing.T) {
		_, err := db.Exec(`INSERT INTO "LoginCodeRequest" (email, ip_hash) VALUES ('person@example.com', 'iphash')`)
		require.NoError(t, err)
		_, err = db.Exec(`INSERT INTO "LoginCodeRequest" (email) VALUES ('person@example.com')`)
		require.Error(t, err)
	})

	t.Run("down removes the new shape, keeps the demo email in google_id, and up restores it", func(t *testing.T) {
		_, err := db.Exec(down)
		require.NoError(t, err)

		var columns, tables int
		require.NoError(t, db.QueryRow(`SELECT count(*) FROM information_schema.columns
			WHERE table_schema = $1 AND table_name = 'Caregiver' AND column_name = 'email'`, schema).Scan(&columns))
		require.Equal(t, 0, columns)
		require.NoError(t, db.QueryRow(`SELECT count(*) FROM information_schema.tables
			WHERE table_schema = $1 AND table_name IN ('LoginCode', 'LoginCodeRequest')`, schema).Scan(&tables))
		require.Equal(t, 0, tables)
		var googleID string
		require.NoError(t, db.QueryRow(`SELECT google_id FROM "Caregiver" WHERE name = 'Jane Smith'`).Scan(&googleID))
		require.Equal(t, "test.caregiver@example.com", googleID)

		_, err = db.Exec(up)
		require.NoError(t, err)
		requireEmailLoginShape(t, db)
		var email string
		require.NoError(t, db.QueryRow(`SELECT email FROM "Caregiver" WHERE name = 'Jane Smith'`).Scan(&email))
		require.Equal(t, "test.caregiver@example.com", email)
	})
}

func requireEmailLoginShape(t *testing.T, db *sql.DB) {
	t.Helper()
	columnsOf := func(table string) map[string]string {
		rows, err := db.Query(`SELECT column_name, data_type, is_nullable FROM information_schema.columns
			WHERE table_schema = current_schema() AND table_name = $1`, table)
		require.NoError(t, err)
		defer rows.Close()
		got := map[string]string{}
		for rows.Next() {
			var name, dataType, nullable string
			require.NoError(t, rows.Scan(&name, &dataType, &nullable))
			got[name] = dataType + "/" + nullable
		}
		require.NoError(t, rows.Err())
		return got
	}

	require.Equal(t, map[string]string{
		"id":        "integer/NO",
		"name":      "text/YES",
		"role":      "text/YES",
		"google_id": "text/YES",
		"email":     "text/YES",
	}, columnsOf("Caregiver"))
	require.Equal(t, map[string]string{
		"id":         "integer/NO",
		"email":      "text/NO",
		"code_hash":  "text/NO",
		"attempts":   "integer/NO",
		"created_at": "timestamp with time zone/NO",
		"expires_at": "timestamp with time zone/NO",
		"used_at":    "timestamp with time zone/YES",
	}, columnsOf("LoginCode"))
	require.Equal(t, map[string]string{
		"id":         "integer/NO",
		"email":      "text/NO",
		"ip_hash":    "text/NO",
		"created_at": "timestamp with time zone/NO",
	}, columnsOf("LoginCodeRequest"))

	var indexes int
	require.NoError(t, db.QueryRow(`SELECT count(*) FROM pg_indexes WHERE schemaname = current_schema()
		AND indexname IN ('Caregiver_email_lower_key', 'LoginCode_email_created_at_idx',
			'LoginCodeRequest_email_created_at_idx', 'LoginCodeRequest_ip_hash_created_at_idx')`).Scan(&indexes))
	require.Equal(t, 4, indexes, "the code lookup and both rate limit counts need their indexes")
}
