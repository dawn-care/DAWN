package playback

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	playback "server/internal/models/playback"
)

var (
	ErrInvalidInput      = errors.New("invalid playback input")
	ErrNotFound          = errors.New("playback not found")
	ErrInvalidTransition = errors.New("invalid playback status transition")
)

type PlaybackService struct {
	DB *sql.DB
}

func NewPlaybackService(db *sql.DB) *PlaybackService {
	return &PlaybackService{DB: db}
}

const columns = `id, carepartner_id, media_id, media_kind, media_display_name, scheduled_item_id,
	source, status, requested_at, acknowledged_at, started_at, ended_at, expires_at, rating`

func scanPlayback(row *sql.Row) (*playback.Playback, error) {
	p := &playback.Playback{}
	err := row.Scan(&p.ID, &p.CarepartnerID, &p.MediaID, &p.MediaKind, &p.MediaDisplayName,
		&p.ScheduledItemID, &p.Source, &p.Status,
		&p.RequestedAt, &p.AcknowledgedAt, &p.StartedAt, &p.EndedAt, &p.ExpiresAt, &p.Rating)
	if err != nil {
		return nil, err
	}
	return p, nil
}

// CreatePlayback is shared by scheduled, staff-requested and tablet-initiated plays.
// A scheduled play names its scheduled item; tap and staff plays don't have one.
// The caller must supply an expiry deadline; no arbitrary default is assumed.
func (s *PlaybackService) CreatePlayback(ctx context.Context, carepartnerID, mediaID int, scheduledItemID *int, source string, expiresAt time.Time) (*playback.Playback, error) {
	now := time.Now().UTC()
	if carepartnerID <= 0 || mediaID <= 0 || !expiresAt.After(now) {
		return nil, fmt.Errorf("%w: positive IDs and a future expiry are required", ErrInvalidInput)
	}
	switch source {
	case playback.SourceScheduled:
		if scheduledItemID == nil || *scheduledItemID <= 0 {
			return nil, fmt.Errorf("%w: scheduled plays need a scheduled_item_id", ErrInvalidInput)
		}
	case playback.SourceStaff, playback.SourceTap:
		if scheduledItemID != nil {
			return nil, fmt.Errorf("%w: only scheduled plays have a scheduled_item_id", ErrInvalidInput)
		}
	default:
		return nil, fmt.Errorf("%w: unknown source", ErrInvalidInput)
	}

	// Check in the INSERT, rather than trusting the request, that the media belongs to the
	// carepartner and that a scheduled item belongs to them and plays this media.
	// The media's kind and name are copied so the record outlives the media.
	p, err := scanPlayback(s.DB.QueryRowContext(ctx, `
		INSERT INTO "Playback" (carepartner_id, media_id, media_kind, media_display_name,
		                        scheduled_item_id, source, requested_at, expires_at)
		SELECT $1, m.id, m.kind, m.display_name, $3::int, $4, $5, $6 FROM "Media" m
		WHERE m.id = $2 AND m.carepartner_id = $1
		  AND ($3::int IS NULL OR EXISTS (
		    SELECT 1 FROM "ScheduledItem" si
		    WHERE si.id = $3::int AND si.carepartner_id = $1 AND si.media_id = $2))
		RETURNING `+columns, carepartnerID, mediaID, scheduledItemID, source, now, expiresAt.UTC()))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("%w: media or scheduled item does not belong to this carepartner", ErrInvalidInput)
	}
	if err != nil {
		return nil, fmt.Errorf("create playback: %w", err)
	}
	return p, nil
}

// UpdateStatus serializes reports for a playback so simultaneous requests cannot
// overwrite a terminal outcome. Carepartner scoping is not a substitute for
// authenticating the tablet in the HTTP layer.
func (s *PlaybackService) UpdateStatus(ctx context.Context, id, carepartnerID int, status string) (*playback.Playback, error) {
	if id <= 0 || carepartnerID <= 0 {
		return nil, fmt.Errorf("%w: positive IDs are required", ErrInvalidInput)
	}
	switch status {
	case playback.StatusRequested, playback.StatusAcknowledged, playback.StatusStarted,
		playback.StatusFinished, playback.StatusStopped, playback.StatusFailed, playback.StatusExpired:
	default:
		return nil, fmt.Errorf("%w: unknown status", ErrInvalidInput)
	}

	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	p, err := scanPlayback(tx.QueryRowContext(ctx, `SELECT `+columns+`
		FROM "Playback" WHERE id = $1 AND carepartner_id = $2 FOR UPDATE`, id, carepartnerID))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	// Retrying a report must not move its timestamp or change an ended playback.
	if p.Status == status {
		return p, tx.Commit()
	}
	if !canTransition(p.Status, status) {
		return nil, ErrInvalidTransition
	}
	now := time.Now().UTC()
	if status == playback.StatusExpired && now.Before(p.ExpiresAt) {
		return nil, fmt.Errorf("%w: expiry deadline has not passed", ErrInvalidTransition)
	}
	// Expiry is a deadline for the tablet to accept a command, not a limit on how
	// long a clip may take to start or play. Acknowledging only moves a playback to
	// "acknowledged"; it does not mark it finished. A tablet that acknowledges in
	// time has committed to playing, so a later "started" (for example after a slow
	// video buffers) is recorded as it happened rather than rewritten to "expired".
	// If the tablet then fails to play, it reports "failed"; if it never reports
	// anything, ExpireStale (commands.go) marks it "failed" AcknowledgedGrace after
	// the deadline, so an acknowledged playback cannot stay looking healthy.
	if status == playback.StatusAcknowledged && !now.Before(p.ExpiresAt) {
		status = playback.StatusExpired
	}
	p.Status = status
	switch status {
	case playback.StatusAcknowledged:
		p.AcknowledgedAt = &now
	case playback.StatusStarted:
		p.StartedAt = &now
	default:
		p.EndedAt = &now
	}
	_, err = tx.ExecContext(ctx, `UPDATE "Playback"
		SET status = $1, acknowledged_at = $2, started_at = $3, ended_at = $4
		WHERE id = $5 AND carepartner_id = $6`,
		p.Status, p.AcknowledgedAt, p.StartedAt, p.EndedAt, id, carepartnerID)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return p, nil
}

func canTransition(from, to string) bool {
	switch from {
	case playback.StatusRequested:
		return to == playback.StatusAcknowledged || to == playback.StatusStopped ||
			to == playback.StatusFailed || to == playback.StatusExpired
	case playback.StatusAcknowledged:
		return to == playback.StatusStarted || to == playback.StatusStopped ||
			to == playback.StatusFailed || to == playback.StatusExpired
	case playback.StatusStarted:
		return to == playback.StatusFinished || to == playback.StatusStopped || to == playback.StatusFailed
	default:
		return false
	}
}
