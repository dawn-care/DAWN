package playback

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	playback "server/internal/models/playback"
)

const (
	// StaffCommandTTL is how long a tablet has to acknowledge a staff play
	// command. The server sets it so a clip never starts long after the moment a
	// staff member asked for it.
	StaffCommandTTL = 30 * time.Second

	// AcknowledgedGrace is how long past its deadline an acknowledged command may
	// stay unstarted before the sweep marks it failed — for a tablet that accepted
	// a command and then crashed or lost power.
	AcknowledgedGrace = 2 * time.Minute
)

// RequestStaffPlay records a staff member's request to play a media item on a
// carepartner's tablet, with a server-set deadline.
func (s *PlaybackService) RequestStaffPlay(ctx context.Context, carepartnerID, mediaID int) (*playback.Playback, error) {
	return s.CreatePlayback(ctx, carepartnerID, mediaID, nil, playback.SourceStaff, time.Now().UTC().Add(StaffCommandTTL))
}

// LiveCommand returns the most recent staff command a tablet should still act
// on, or nil if there is none. It is resent when a tablet reconnects, so a
// command issued during a brief network drop is not lost. Tablets must ignore a
// command whose playback ID they have already handled. A command whose media was
// deleted since is not resent; it expires like any other unanswered command.
func (s *PlaybackService) LiveCommand(ctx context.Context, carepartnerID int) (*playback.Playback, error) {
	if carepartnerID <= 0 {
		return nil, fmt.Errorf("%w: positive IDs are required", ErrInvalidInput)
	}
	p, err := scanPlayback(s.DB.QueryRowContext(ctx, `SELECT `+columns+`
		FROM "Playback"
		WHERE carepartner_id = $1 AND source = $2 AND status IN ($3, $4) AND expires_at > $5
		  AND media_id IS NOT NULL
		ORDER BY requested_at DESC, id DESC
		LIMIT 1`,
		carepartnerID, playback.SourceStaff, playback.StatusRequested, playback.StatusAcknowledged, time.Now().UTC()))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return p, nil
}

// ExpireStale closes out commands nobody will act on: requested commands whose
// deadline has passed become "expired", and acknowledged commands that never
// started within AcknowledgedGrace of their deadline become "failed". Without
// this, a command sent to an offline tablet would stay "requested" forever.
func (s *PlaybackService) ExpireStale(ctx context.Context) (expired, failed int64, err error) {
	now := time.Now().UTC()

	res, err := s.DB.ExecContext(ctx, `UPDATE "Playback"
		SET status = $1, ended_at = $2
		WHERE status = $3 AND expires_at <= $2`,
		playback.StatusExpired, now, playback.StatusRequested)
	if err != nil {
		return 0, 0, fmt.Errorf("expire requested playbacks: %w", err)
	}
	if expired, err = res.RowsAffected(); err != nil {
		return 0, 0, err
	}

	res, err = s.DB.ExecContext(ctx, `UPDATE "Playback"
		SET status = $1, ended_at = $2
		WHERE status = $3 AND expires_at <= $4`,
		playback.StatusFailed, now, playback.StatusAcknowledged, now.Add(-AcknowledgedGrace))
	if err != nil {
		return expired, 0, fmt.Errorf("fail stalled playbacks: %w", err)
	}
	if failed, err = res.RowsAffected(); err != nil {
		return expired, 0, err
	}
	return expired, failed, nil
}

// RunExpirySweep calls ExpireStale every interval until ctx is cancelled.
func (s *PlaybackService) RunExpirySweep(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			expired, failed, err := s.ExpireStale(ctx)
			if err != nil {
				log.Printf("playback sweep: %v", err)
			} else if expired > 0 || failed > 0 {
				log.Printf("playback sweep: %d expired, %d failed", expired, failed)
			}
		}
	}
}

// StopOutcome says what a stop request did, so the caller knows whether the
// playback is already settled or the tablet still has to act.
type StopOutcome int

const (
	// StopCancelled: the tablet had not accepted the command yet, so nothing was
	// playing. It is recorded as stopped now and will not be resent.
	StopCancelled StopOutcome = iota
	// StopForwarded: the tablet has accepted the command and may be playing. The
	// record is left as it is; the tablet must stop and report "stopped" itself,
	// so the record never says stopped while a clip is still playing.
	StopForwarded
	// StopAlreadyStopped: a repeated stop for a playback already stopped.
	StopAlreadyStopped
)

// RequestStop handles a staff member stopping a playback. Only a command the
// tablet has not yet accepted is recorded as stopped here; anything the tablet
// may be playing waits for the tablet to confirm.
func (s *PlaybackService) RequestStop(ctx context.Context, id, carepartnerID int) (*playback.Playback, StopOutcome, error) {
	if id <= 0 || carepartnerID <= 0 {
		return nil, 0, fmt.Errorf("%w: positive IDs are required", ErrInvalidInput)
	}

	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, 0, err
	}
	defer tx.Rollback()
	p, err := scanPlayback(tx.QueryRowContext(ctx, `SELECT `+columns+`
		FROM "Playback" WHERE id = $1 AND carepartner_id = $2 FOR UPDATE`, id, carepartnerID))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, 0, ErrNotFound
	}
	if err != nil {
		return nil, 0, err
	}

	switch p.Status {
	case playback.StatusRequested:
		now := time.Now().UTC()
		if _, err := tx.ExecContext(ctx, `UPDATE "Playback" SET status = $1, ended_at = $2
			WHERE id = $3 AND carepartner_id = $4`,
			playback.StatusStopped, now, id, carepartnerID); err != nil {
			return nil, 0, err
		}
		if err := tx.Commit(); err != nil {
			return nil, 0, err
		}
		p.Status, p.EndedAt = playback.StatusStopped, &now
		return p, StopCancelled, nil
	case playback.StatusAcknowledged, playback.StatusStarted:
		return p, StopForwarded, nil
	case playback.StatusStopped:
		return p, StopAlreadyStopped, nil
	default:
		// finished, failed or expired: there is nothing left to stop.
		return nil, 0, ErrInvalidTransition
	}
}
