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
	playback "server/internal/services/playback"
)

func TestPlaybackMigrationIntegration(t *testing.T) {
	dsn := os.Getenv("MIGRATION_TEST_DSN")
	if dsn == "" {
		t.Skip("requires disposable PostgreSQL via MIGRATION_TEST_DSN")
	}
	admin, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	defer admin.Close()
	for _, legacy := range []bool{false, true} {
		t.Run(fmt.Sprintf("existing_database=%v", legacy), func(t *testing.T) {
			ctx := context.Background()
			schema := fmt.Sprintf("playback_test_%d", time.Now().UnixNano())
			_, err := admin.Exec("CREATE SCHEMA " + pq.QuoteIdentifier(schema))
			require.NoError(t, err)
			defer admin.Exec("DROP SCHEMA " + pq.QuoteIdentifier(schema) + " CASCADE")
			db, err := sql.Open("postgres", dsn+" search_path="+schema)
			require.NoError(t, err)
			defer db.Close()
			var before map[string]string
			if legacy {
				initial, err := files.ReadFile("sql/00001_initial_schema.sql")
				require.NoError(t, err)
				_, err = db.Exec(strings.Split(string(initial), "-- +goose Down")[0])
				require.NoError(t, err)
				_, err = db.Exec(`INSERT INTO "Caregiver" (name) VALUES ('Existing user')`)
				require.NoError(t, err)
				before = snapshot(t, db, schema)
				require.NoError(t, Baseline(ctx, db))
			}
			require.NoError(t, Up(ctx, db))
			require.NoError(t, Up(ctx, db), "repeating migrations must be safe")
			if legacy {
				after := snapshot(t, db, schema)
				delete(after, "Playback")
				delete(after, "Playback_id_seq")
				require.Equal(t, before, after, "existing rows and sequences must be preserved")
			}
			var applied int
			require.NoError(t, db.QueryRow(`SELECT count(*) FROM goose_db_version WHERE version_id = 2 AND is_applied`).Scan(&applied))
			require.Equal(t, 1, applied)
			var cp, module int
			require.NoError(t, db.QueryRow(`SELECT carepartner_id, id FROM "Module" ORDER BY id LIMIT 1`).Scan(&cp, &module))
			svc := playback.NewPlaybackService(db)
			p, err := svc.CreatePlayback(ctx, cp, module, "tap", time.Now().Add(time.Hour))
			require.NoError(t, err)
			require.Nil(t, p.Rating)
			for _, status := range []string{"acknowledged", "started", "finished"} {
				p, err = svc.UpdateStatus(ctx, p.ID, cp, status)
				require.NoError(t, err)
				require.Equal(t, status, p.Status)
			}
			require.NotNil(t, p.AcknowledgedAt)
			require.NotNil(t, p.StartedAt)
			require.NotNil(t, p.EndedAt)
			require.False(t, p.StartedAt.Before(*p.AcknowledgedAt))
			require.False(t, p.EndedAt.Before(*p.StartedAt))
			again, err := svc.UpdateStatus(ctx, p.ID, cp, "finished")
			require.NoError(t, err)
			require.WithinDuration(t, *p.EndedAt, *again.EndedAt, time.Microsecond)
			_, err = svc.UpdateStatus(ctx, p.ID, cp, "started")
			require.ErrorIs(t, err, playback.ErrInvalidTransition)
			_, err = svc.UpdateStatus(ctx, p.ID, cp+1000, "finished")
			require.ErrorIs(t, err, playback.ErrNotFound)
			for _, query := range []string{
				`UPDATE "Playback" SET source = 'notes'`,
				`UPDATE "Playback" SET status = 'unknown'`,
				`UPDATE "Playback" SET carepartner_id = -1`,
				`UPDATE "Playback" SET module_id = -1`,
				`UPDATE "Playback" SET expires_at = NULL`,
			} {
				_, err := db.Exec(query)
				require.Error(t, err, query)
			}
		})
	}
}
