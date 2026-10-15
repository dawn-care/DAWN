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
			var carepartnersBefore string
			if legacy {
				initial, err := files.ReadFile("sql/00001_initial_schema.sql")
				require.NoError(t, err)
				_, err = db.Exec(strings.Split(string(initial), "-- +goose Down")[0])
				require.NoError(t, err)
				_, err = db.Exec(`INSERT INTO "Caregiver" (name) VALUES ('Existing user')`)
				require.NoError(t, err)
				before = snapshot(t, db, schema)
				carepartnersBefore = carepartnerCore(t, db)
				require.NoError(t, Baseline(ctx, db))
			}
			require.NoError(t, Up(ctx, db))
			require.NoError(t, Up(ctx, db), "repeating migrations must be safe")
			if legacy {
				// Migration 3 drops the old content tables and reshapes Carepartner, so compare
				// the tables it keeps as they are, and carepartners by identity, name and caregiver.
				after := snapshot(t, db, schema)
				for _, kept := range []string{"Caregiver", "Caregiver_id_seq", "Carepartner_id_seq",
					"Session", "Session_id_seq", "DevicePairing", "DevicePairing_id_seq"} {
					require.Equal(t, before[kept], after[kept], "existing %s must be preserved", kept)
				}
				require.Equal(t, carepartnersBefore, carepartnerCore(t, db), "existing carepartners must be preserved")
			}
			var applied int
			require.NoError(t, db.QueryRow(`SELECT count(*) FROM goose_db_version WHERE version_id = 2 AND is_applied`).Scan(&applied))
			require.Equal(t, 1, applied)
			var cp, media, item int
			require.NoError(t, db.QueryRow(`SELECT id FROM "Carepartner" ORDER BY id LIMIT 1`).Scan(&cp))
			require.NoError(t, db.QueryRow(`INSERT INTO "Media" (carepartner_id, kind, display_name, object_key, mime_type, size_bytes, duration_seconds)
				VALUES ($1, 'video', 'Clip', 'playback-test', 'video/mp4', 1, 60) RETURNING id`, cp).Scan(&media))
			require.NoError(t, db.QueryRow(`INSERT INTO "ScheduledItem" (carepartner_id, media_id, start_date, start_time, repeat)
				VALUES ($1, $2, CURRENT_DATE, '10:00', 'daily') RETURNING id`, cp, media).Scan(&item))
			svc := playback.NewPlaybackService(db)
			scheduled, err := svc.CreatePlayback(ctx, cp, media, &item, "scheduled", time.Now().Add(time.Hour))
			require.NoError(t, err)
			require.Equal(t, item, *scheduled.ScheduledItemID)
			_, err = svc.CreatePlayback(ctx, cp+1000, media, nil, "tap", time.Now().Add(time.Hour))
			require.ErrorIs(t, err, playback.ErrInvalidInput, "media of another carepartner")
			p, err := svc.CreatePlayback(ctx, cp, media, nil, "tap", time.Now().Add(time.Hour))
			require.NoError(t, err)
			require.Equal(t, media, *p.MediaID)
			require.Equal(t, "video", p.MediaKind)
			require.Equal(t, "Clip", p.MediaDisplayName)
			require.Nil(t, p.ScheduledItemID)
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
				`UPDATE "Playback" SET media_id = -1`,
				`UPDATE "Playback" SET expires_at = NULL`,
			} {
				_, err := db.Exec(query)
				require.Error(t, err, query)
			}
			// Deleting the media keeps its playbacks and what they played. It also deletes
			// the scheduled item that played it, which clears the scheduled play's link.
			_, err = db.Exec(`DELETE FROM "Media" WHERE id = $1`, media)
			require.NoError(t, err)
			var kept, unlinked int
			require.NoError(t, db.QueryRow(`SELECT count(*), count(*) FILTER (WHERE media_id IS NULL AND scheduled_item_id IS NULL
				AND media_kind = 'video' AND media_display_name = 'Clip') FROM "Playback" WHERE id IN ($1, $2)`,
				p.ID, scheduled.ID).Scan(&kept, &unlinked))
			require.Equal(t, 2, kept)
			require.Equal(t, 2, unlinked)
		})
	}
}

// carepartnerCore lists carepartners by the fields every migration keeps
func carepartnerCore(t *testing.T, db *sql.DB) string {
	t.Helper()
	var core string
	require.NoError(t, db.QueryRow(`SELECT COALESCE(jsonb_agg(jsonb_build_array(id, name, NULLIF(caregiver_id, 0)) ORDER BY id), '[]')::text
		FROM "Carepartner"`).Scan(&core))
	return core
}
