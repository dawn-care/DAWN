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

const columns = `id, carepartner_id, module_id, source, status, requested_at,
	acknowledged_at, started_at, ended_at, expires_at, rating`

func scanPlayback(row *sql.Row) (*playback.Playback, error) {
	p := &playback.Playback{}
	err := row.Scan(&p.ID, &p.CarepartnerID, &p.ModuleID, &p.Source, &p.Status,
		&p.RequestedAt, &p.AcknowledgedAt, &p.StartedAt, &p.EndedAt, &p.ExpiresAt, &p.Rating)
	if err != nil {
		return nil, err
	}
	return p, nil
}

// CreatePlayback is shared by scheduled, staff-requested and tablet-initiated plays.
// The caller must supply an expiry deadline; no arbitrary default is assumed.
func (s *PlaybackService) CreatePlayback(ctx context.Context, carepartnerID, moduleID int, source string, expiresAt time.Time) (*playback.Playback, error) {
	now := time.Now().UTC()
	if carepartnerID <= 0 || moduleID <= 0 || !expiresAt.After(now) {
		return nil, fmt.Errorf("%w: positive IDs and a future expiry are required", ErrInvalidInput)
	}
	switch source {
	case playback.SourceScheduled, playback.SourceStaff, playback.SourceTap:
	default:
		return nil, fmt.Errorf("%w: unknown source", ErrInvalidInput)
	}

	// Check the module's carepartner in the INSERT rather than trusting the request.
	p, err := scanPlayback(s.DB.QueryRowContext(ctx, `
		INSERT INTO "Playback" (carepartner_id, module_id, source, requested_at, expires_at)
		SELECT $1, m.id, $3, $4, $5 FROM "Module" m
		WHERE m.id = $2 AND m.carepartner_id = $1
		RETURNING `+columns, carepartnerID, moduleID, source, now, expiresAt.UTC()))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("%w: module does not belong to this carepartner", ErrInvalidInput)
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
	// Expiry is a deadline to start, not a limit on the length of a playing clip.
	if (status == playback.StatusAcknowledged || status == playback.StatusStarted) && !now.Before(p.ExpiresAt) {
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
