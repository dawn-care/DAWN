package playback

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

// timeBetween matches a time argument inside [from, to], for deadlines the
// service computes from the current time.
type timeBetween struct{ from, to time.Time }

func (m timeBetween) Match(v driver.Value) bool {
	t, ok := v.(time.Time)
	return ok && !t.Before(m.from) && !t.After(m.to)
}

func TestRequestStaffPlaySetsServerDeadline(t *testing.T) {
	s, mock := setup(t)
	before := time.Now().UTC()
	mock.ExpectQuery(`INSERT INTO "Playback".*WHERE m.id = \$2 AND m.carepartner_id = \$1`).
		WithArgs(2, 3, "staff", sqlmock.AnyArg(),
			timeBetween{before.Add(StaffCommandTTL), time.Now().UTC().Add(StaffCommandTTL + time.Second)}).
		WillReturnRows(playbackRows("requested", before.Add(StaffCommandTTL)))

	p, err := s.RequestStaffPlay(context.Background(), 2, 3)
	require.NoError(t, err)
	require.Equal(t, "requested", p.Status)
}

func TestRequestStaffPlayRejectsInvalidIDs(t *testing.T) {
	s, _ := setup(t)
	_, err := s.RequestStaffPlay(context.Background(), 0, 3)
	require.ErrorIs(t, err, ErrInvalidInput)
}

func TestLiveCommand(t *testing.T) {
	t.Run("returns the most recent unexpired staff command", func(t *testing.T) {
		s, mock := setup(t)
		mock.ExpectQuery(`SELECT .* FROM "Playback".*source = \$2 AND status IN \(\$3, \$4\) AND expires_at > \$5.*ORDER BY requested_at DESC`).
			WithArgs(2, "staff", "requested", "acknowledged", sqlmock.AnyArg()).
			WillReturnRows(playbackRows("requested", time.Now().Add(time.Minute)))

		p, err := s.LiveCommand(context.Background(), 2)
		require.NoError(t, err)
		require.NotNil(t, p)
		require.Equal(t, 1, p.ID)
	})

	t.Run("nil when nothing is live", func(t *testing.T) {
		s, mock := setup(t)
		mock.ExpectQuery(`SELECT .* FROM "Playback"`).WillReturnRows(sqlmock.NewRows(nil))

		p, err := s.LiveCommand(context.Background(), 2)
		require.NoError(t, err)
		require.Nil(t, p)
	})

	t.Run("rejects invalid IDs without querying", func(t *testing.T) {
		s, _ := setup(t)
		_, err := s.LiveCommand(context.Background(), 0)
		require.ErrorIs(t, err, ErrInvalidInput)
	})

	t.Run("database errors are returned", func(t *testing.T) {
		s, mock := setup(t)
		cause := errors.New("database unavailable")
		mock.ExpectQuery(`SELECT .* FROM "Playback"`).WillReturnError(cause)

		_, err := s.LiveCommand(context.Background(), 2)
		require.ErrorIs(t, err, cause)
	})
}

func TestExpireStale(t *testing.T) {
	t.Run("expires unacknowledged commands and fails stalled ones", func(t *testing.T) {
		s, mock := setup(t)
		now := time.Now().UTC()
		mock.ExpectExec(`UPDATE "Playback"\s+SET status = \$1, ended_at = \$2\s+WHERE status = \$3 AND expires_at <= \$2`).
			WithArgs("expired", sqlmock.AnyArg(), "requested").
			WillReturnResult(sqlmock.NewResult(0, 2))
		mock.ExpectExec(`UPDATE "Playback"\s+SET status = \$1, ended_at = \$2\s+WHERE status = \$3 AND expires_at <= \$4`).
			// Stalled means acknowledged and still unstarted AcknowledgedGrace past the deadline.
			WithArgs("failed", sqlmock.AnyArg(), "acknowledged",
				timeBetween{now.Add(-AcknowledgedGrace - time.Second), time.Now().UTC().Add(-AcknowledgedGrace + time.Second)}).
			WillReturnResult(sqlmock.NewResult(0, 1))

		expired, failed, err := s.ExpireStale(context.Background())
		require.NoError(t, err)
		require.EqualValues(t, 2, expired)
		require.EqualValues(t, 1, failed)
	})

	t.Run("stops at the first database error", func(t *testing.T) {
		s, mock := setup(t)
		cause := errors.New("database unavailable")
		mock.ExpectExec(`UPDATE "Playback"`).WillReturnError(cause)

		_, _, err := s.ExpireStale(context.Background())
		require.ErrorIs(t, err, cause)
	})
}

func TestRunExpirySweepStopsWithContext(t *testing.T) {
	s, _ := setup(t) // no expectations: the sweep must not run once cancelled

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	done := make(chan struct{})
	go func() {
		s.RunExpirySweep(ctx, time.Hour)
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("RunExpirySweep did not return after its context was cancelled")
	}
}

func TestRequestStop(t *testing.T) {
	for _, test := range []struct {
		from       string
		outcome    StopOutcome
		wantStatus string
		err        error
	}{
		// Not accepted by the tablet yet, so nothing is playing: cancel it now.
		{from: "requested", outcome: StopCancelled, wantStatus: "stopped"},
		// The tablet may be playing: leave the record for the tablet to update.
		{from: "acknowledged", outcome: StopForwarded, wantStatus: "acknowledged"},
		{from: "started", outcome: StopForwarded, wantStatus: "started"},
		{from: "stopped", outcome: StopAlreadyStopped, wantStatus: "stopped"},
		{from: "finished", err: ErrInvalidTransition},
		{from: "failed", err: ErrInvalidTransition},
		{from: "expired", err: ErrInvalidTransition},
	} {
		t.Run(test.from, func(t *testing.T) {
			s, mock := setup(t)
			mock.ExpectBegin()
			mock.ExpectQuery(`SELECT .* FROM "Playback" WHERE id = \$1 AND carepartner_id = \$2 FOR UPDATE`).
				WithArgs(1, 2).WillReturnRows(playbackRows(test.from, time.Now().Add(time.Minute)))
			if test.outcome == StopCancelled && test.err == nil {
				mock.ExpectExec(`UPDATE "Playback" SET status = \$1, ended_at = \$2`).
					WithArgs("stopped", sqlmock.AnyArg(), 1, 2).WillReturnResult(sqlmock.NewResult(0, 1))
				mock.ExpectCommit()
			} else {
				// Nothing is written, so the transaction is only rolled back.
				mock.ExpectRollback()
			}

			p, outcome, err := s.RequestStop(context.Background(), 1, 2)
			if test.err != nil {
				require.ErrorIs(t, err, test.err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, test.outcome, outcome)
			require.Equal(t, test.wantStatus, p.Status)
			if test.outcome == StopCancelled {
				require.NotNil(t, p.EndedAt)
			}
		})
	}
}

func TestRequestStopErrors(t *testing.T) {
	t.Run("invalid IDs are rejected without querying", func(t *testing.T) {
		s, _ := setup(t)
		_, _, err := s.RequestStop(context.Background(), 0, 2)
		require.ErrorIs(t, err, ErrInvalidInput)
	})

	t.Run("unknown or another carepartner's playback is not found", func(t *testing.T) {
		s, mock := setup(t)
		mock.ExpectBegin()
		mock.ExpectQuery(`SELECT .* FOR UPDATE`).WithArgs(1, 2).WillReturnError(sql.ErrNoRows)
		mock.ExpectRollback()

		_, _, err := s.RequestStop(context.Background(), 1, 2)
		require.ErrorIs(t, err, ErrNotFound)
	})

	t.Run("a failed write leaves the playback unstopped", func(t *testing.T) {
		s, mock := setup(t)
		cause := errors.New("database unavailable")
		mock.ExpectBegin()
		mock.ExpectQuery(`SELECT .* FOR UPDATE`).WithArgs(1, 2).WillReturnRows(playbackRows("requested", time.Now().Add(time.Minute)))
		mock.ExpectExec(`UPDATE "Playback"`).WillReturnError(cause)
		mock.ExpectRollback()

		_, _, err := s.RequestStop(context.Background(), 1, 2)
		require.ErrorIs(t, err, cause)
	})
}
