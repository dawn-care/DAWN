package playback

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
	model "server/internal/models/playback"
)

func setup(t *testing.T) (*PlaybackService, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, mock.ExpectationsWereMet()); db.Close() })
	return NewPlaybackService(db), mock
}

func playbackRows(status string, expiry time.Time) *sqlmock.Rows {
	start := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	var ack, began, ended any
	if status != "requested" {
		ack = start.Add(time.Second)
	}
	if status == "started" || status == "finished" {
		began = start.Add(2 * time.Second)
	}
	if status == "finished" || status == "stopped" || status == "failed" || status == "expired" {
		ended = start.Add(3 * time.Second)
	}
	return sqlmock.NewRows(strings.Fields(strings.ReplaceAll(columns, ",", ""))).AddRow(1, 2, 3, "video", "Clip", nil, "tap", status, start, ack, began, ended, expiry, nil)
}

func intp(i int) *int { return &i }

func TestCreatePlayback(t *testing.T) {
	for _, source := range []string{"scheduled", "staff", "tap"} {
		t.Run(source, func(t *testing.T) {
			s, mock := setup(t)
			expiry := time.Now().UTC().Add(time.Hour)
			// Only scheduled plays name a scheduled item
			var item *int
			if source == "scheduled" {
				item = intp(5)
			}
			mock.ExpectQuery(`INSERT INTO "Playback".*FROM "Media" m.*WHERE m.id = \$2 AND m.carepartner_id = \$1.*"ScheduledItem"`).
				WithArgs(2, 3, item, source, sqlmock.AnyArg(), expiry).WillReturnRows(playbackRows("requested", expiry))
			p, err := s.CreatePlayback(context.Background(), 2, 3, item, source, expiry)
			require.NoError(t, err)
			require.Equal(t, "requested", p.Status)
			require.Nil(t, p.AcknowledgedAt)
			require.Nil(t, p.StartedAt)
			require.Nil(t, p.EndedAt)
			require.Nil(t, p.Rating)
		})
	}
}

func TestCreateRejectsInvalidInput(t *testing.T) {
	for _, test := range []struct {
		cp, media int
		item      *int
		source    string
		expiry    time.Time
	}{
		{0, 3, nil, "tap", time.Now().Add(time.Hour)}, {2, 0, nil, "tap", time.Now().Add(time.Hour)},
		{2, 3, nil, "arbitrary text", time.Now().Add(time.Hour)}, {2, 3, nil, "tap", time.Time{}},
		{2, 3, nil, "tap", time.Now().Add(-time.Hour)},
		// scheduled plays need their scheduled item; tap and staff plays can't have one
		{2, 3, nil, "scheduled", time.Now().Add(time.Hour)}, {2, 3, intp(0), "scheduled", time.Now().Add(time.Hour)},
		{2, 3, intp(5), "tap", time.Now().Add(time.Hour)}, {2, 3, intp(5), "staff", time.Now().Add(time.Hour)},
	} {
		s, _ := setup(t)
		_, err := s.CreatePlayback(context.Background(), test.cp, test.media, test.item, test.source, test.expiry)
		require.ErrorIs(t, err, ErrInvalidInput)
	}
}

func TestCreateDatabaseErrors(t *testing.T) {
	for _, cause := range []error{sql.ErrNoRows, errors.New("database unavailable")} {
		s, mock := setup(t)
		mock.ExpectQuery(`INSERT INTO "Playback"`).WillReturnError(cause)
		p, err := s.CreatePlayback(context.Background(), 2, 3, nil, "tap", time.Now().Add(time.Hour))
		require.Nil(t, p)
		if errors.Is(cause, sql.ErrNoRows) {
			require.ErrorIs(t, err, ErrInvalidInput)
		} else {
			require.ErrorIs(t, err, cause)
		}
	}
}

func TestStatusTransitions(t *testing.T) {
	statuses := []string{"requested", "acknowledged", "started", "finished", "stopped", "failed", "expired"}
	allowed := map[string][]string{
		"requested":    {"acknowledged", "stopped", "failed", "expired"},
		"acknowledged": {"started", "stopped", "failed", "expired"},
		"started":      {"finished", "stopped", "failed"},
	}
	for _, from := range statuses {
		for _, to := range statuses {
			t.Run(from+"/"+to, func(t *testing.T) {
				s, mock := setup(t)
				valid := from == to
				for _, next := range allowed[from] {
					if next == to {
						valid = true
					}
				}
				expiry := time.Now().UTC().Add(time.Hour)
				if to == "expired" {
					expiry = time.Now().UTC().Add(-time.Hour)
				}
				mock.ExpectBegin()
				mock.ExpectQuery(`SELECT .* FROM "Playback" WHERE id = \$1 AND carepartner_id = \$2 FOR UPDATE`).
					WithArgs(1, 2).WillReturnRows(playbackRows(from, expiry))
				if valid {
					if from != to {
						mock.ExpectExec(`UPDATE "Playback"`).WithArgs(to, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), 1, 2).WillReturnResult(sqlmock.NewResult(0, 1))
					}
					mock.ExpectCommit()
				} else {
					mock.ExpectRollback()
				}
				p, err := s.UpdateStatus(context.Background(), 1, 2, to)
				if !valid {
					require.ErrorIs(t, err, ErrInvalidTransition)
					require.Nil(t, p)
					return
				}
				require.NoError(t, err)
				require.Equal(t, to, p.Status)
				if from == to {
					expected, scanErr := scanPlaybackForTest(from, expiry)
					require.NoError(t, scanErr)
					require.Equal(t, expected, p, "a retry must preserve all fields")
				} else {
					switch to {
					case "acknowledged":
						require.NotNil(t, p.AcknowledgedAt)
						require.Nil(t, p.StartedAt)
						require.Nil(t, p.EndedAt)
					case "started":
						require.NotNil(t, p.AcknowledgedAt)
						require.NotNil(t, p.StartedAt)
						require.Nil(t, p.EndedAt)
					default:
						require.NotNil(t, p.EndedAt)
					}
				}
			})
		}
	}
}

func scanPlaybackForTest(status string, expiry time.Time) (*model.Playback, error) {
	db, mock, err := sqlmock.New()
	if err != nil {
		return nil, err
	}
	defer db.Close()
	mock.ExpectQuery("SELECT").WillReturnRows(playbackRows(status, expiry))
	return scanPlayback(db.QueryRow("SELECT"))
}

func TestExpiryDeadline(t *testing.T) {
	for _, test := range []struct {
		from, to, want string
		past           bool
	}{
		{"requested", "acknowledged", "expired", true},
		// Acknowledged in time, so a late start is recorded as it happened.
		{"acknowledged", "started", "started", true},
		{"started", "finished", "finished", true},
		{"requested", "expired", "", false},
	} {
		t.Run(test.from+"/"+test.to, func(t *testing.T) {
			s, mock := setup(t)
			expiry := time.Now().UTC().Add(time.Hour)
			if test.past {
				expiry = time.Now().UTC().Add(-time.Hour)
			}
			mock.ExpectBegin()
			mock.ExpectQuery("SELECT").WillReturnRows(playbackRows(test.from, expiry))
			if test.want == "" {
				mock.ExpectRollback()
			} else {
				mock.ExpectExec(`UPDATE "Playback"`).WithArgs(test.want, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), 1, 2).WillReturnResult(sqlmock.NewResult(0, 1))
				mock.ExpectCommit()
			}
			p, err := s.UpdateStatus(context.Background(), 1, 2, test.to)
			if test.want == "" {
				require.ErrorIs(t, err, ErrInvalidTransition)
				return
			}
			require.NoError(t, err)
			require.Equal(t, test.want, p.Status)
			if test.want == "started" {
				require.NotNil(t, p.StartedAt)
				require.Nil(t, p.EndedAt)
			} else {
				require.NotNil(t, p.EndedAt)
			}
		})
	}
}

func TestUpdateValidationAndFailures(t *testing.T) {
	for _, test := range []struct {
		id, cp int
		status string
	}{{0, 2, "started"}, {1, 0, "started"}, {1, 2, "notes"}} {
		s, _ := setup(t)
		_, err := s.UpdateStatus(context.Background(), test.id, test.cp, test.status)
		require.ErrorIs(t, err, ErrInvalidInput)
	}
	for _, stage := range []string{"begin", "missing", "select", "update", "commit"} {
		t.Run(stage, func(t *testing.T) {
			s, mock := setup(t)
			cause := errors.New("database failure")
			begin := mock.ExpectBegin()
			if stage == "begin" {
				begin.WillReturnError(cause)
			} else {
				query := mock.ExpectQuery("SELECT").WithArgs(1, 2)
				switch stage {
				case "missing":
					query.WillReturnError(sql.ErrNoRows)
				case "select":
					query.WillReturnError(cause)
				default:
					query.WillReturnRows(playbackRows("requested", time.Now().Add(time.Hour)))
					update := mock.ExpectExec(`UPDATE "Playback"`)
					if stage == "update" {
						update.WillReturnError(cause)
					} else {
						update.WillReturnResult(sqlmock.NewResult(0, 1))
					}
				}
				if stage == "commit" {
					mock.ExpectCommit().WillReturnError(cause)
				} else {
					mock.ExpectRollback()
				}
			}
			p, err := s.UpdateStatus(context.Background(), 1, 2, "acknowledged")
			require.Nil(t, p)
			if stage == "missing" {
				require.ErrorIs(t, err, ErrNotFound)
			} else {
				require.ErrorIs(t, err, cause)
			}
		})
	}
}

func TestListPlayed(t *testing.T) {
	s, mock := setup(t)
	expiry := time.Now().UTC().Add(time.Hour)
	rows := playbackRows("finished", expiry)
	rows.AddRow(4, 2, nil, "audio", "Deleted song", nil, "scheduled", "stopped",
		expiry, expiry, expiry, expiry, expiry, nil)
	mock.ExpectQuery(`SELECT .* FROM "Playback"\s+WHERE carepartner_id = \$1 AND started_at IS NOT NULL\s+ORDER BY started_at DESC, id DESC\s+LIMIT \$2`).
		WithArgs(2, 10).WillReturnRows(rows)
	played, err := s.ListPlayed(context.Background(), 2, 10)
	require.NoError(t, err)
	require.Len(t, played, 2)
	require.Equal(t, "finished", played[0].Status)
	// A playback whose media was deleted still says what played
	require.Nil(t, played[1].MediaID)
	require.Equal(t, "Deleted song", played[1].MediaDisplayName)
}

func TestListPlayedEmpty(t *testing.T) {
	s, mock := setup(t)
	mock.ExpectQuery(`SELECT .* FROM "Playback"`).WithArgs(2, 1).
		WillReturnRows(sqlmock.NewRows(strings.Fields(strings.ReplaceAll(columns, ",", ""))))
	played, err := s.ListPlayed(context.Background(), 2, 1)
	require.NoError(t, err)
	require.NotNil(t, played)
	require.Empty(t, played)
}

func TestListPlayedValidationAndFailures(t *testing.T) {
	for _, test := range []struct{ cp, limit int }{{0, 1}, {2, 0}, {2, MaxListLimit + 1}} {
		s, _ := setup(t)
		_, err := s.ListPlayed(context.Background(), test.cp, test.limit)
		require.ErrorIs(t, err, ErrInvalidInput)
	}
	t.Run("query", func(t *testing.T) {
		s, mock := setup(t)
		cause := errors.New("database failure")
		mock.ExpectQuery(`SELECT .* FROM "Playback"`).WillReturnError(cause)
		_, err := s.ListPlayed(context.Background(), 2, 1)
		require.ErrorIs(t, err, cause)
	})
	t.Run("rows", func(t *testing.T) {
		s, mock := setup(t)
		cause := errors.New("connection lost")
		mock.ExpectQuery(`SELECT .* FROM "Playback"`).
			WillReturnRows(playbackRows("finished", time.Now().Add(time.Hour)).RowError(0, cause))
		_, err := s.ListPlayed(context.Background(), 2, 1)
		require.ErrorIs(t, err, cause)
	})
}

// ratedRows is one playback that started at started and, unless ended is nil, ended then
func ratedRows(started time.Time, ended *time.Time, rating any) *sqlmock.Rows {
	var end any
	status := "started"
	if ended != nil {
		end, status = *ended, "finished"
	}
	return sqlmock.NewRows(strings.Fields(strings.ReplaceAll(columns, ",", ""))).
		AddRow(1, 2, 3, "video", "Clip", nil, "staff", status, started, started, started, end, started.Add(30*time.Second), rating)
}

func expectRatingRead(mock sqlmock.Sqlmock, rows *sqlmock.Rows) {
	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT .* FROM "Playback" WHERE id = \$1 AND carepartner_id = \$2 FOR UPDATE`).WithArgs(1, 2).WillReturnRows(rows)
}

func TestRatePlayback(t *testing.T) {
	for name, test := range map[string]struct {
		rating string
		stored int
	}{
		"helped a lot": {"helped_a_lot", 2}, "helped": {"helped", 1}, "no change": {"no_change", 0},
		"made it worse": {"worse", -1}, "made it much worse": {"much_worse", -2},
	} {
		t.Run(name, func(t *testing.T) {
			s, mock := setup(t)
			ended := time.Now().UTC().Add(-time.Minute)
			expectRatingRead(mock, ratedRows(ended.Add(-30*time.Second), &ended, nil))
			mock.ExpectExec(`UPDATE "Playback" SET rating = \$1 WHERE id = \$2 AND carepartner_id = \$3`).
				WithArgs(test.stored, 1, 2).WillReturnResult(sqlmock.NewResult(0, 1))
			mock.ExpectCommit()

			p, err := s.RatePlayback(context.Background(), 1, 2, test.rating)
			require.NoError(t, err)
			require.Equal(t, test.stored, *p.Rating)
		})
	}
}

func TestRatePlaybackReplacesEarlierRating(t *testing.T) {
	s, mock := setup(t)
	ended := time.Now().UTC().Add(-2 * time.Minute)
	// A mis-tap on "worse" fixed to "helped" inside the window
	expectRatingRead(mock, ratedRows(ended.Add(-time.Minute), &ended, -1))
	mock.ExpectExec(`UPDATE "Playback" SET rating`).WithArgs(1, 1, 2).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	p, err := s.RatePlayback(context.Background(), 1, 2, "helped")
	require.NoError(t, err)
	require.Equal(t, 1, *p.Rating)
}

func TestRatePlaybackWindow(t *testing.T) {
	now := time.Now().UTC()
	for name, test := range map[string]struct {
		started time.Time
		ended   *time.Time
		want    error
	}{
		"ended over 5 minutes ago":            {now.Add(-7 * time.Minute), timePtr(now.Add(-6 * time.Minute)), ErrRatingClosed},
		"never ended, started 16 minutes ago": {now.Add(-16 * time.Minute), nil, ErrRatingClosed},
		"never ended, started 10 minutes ago": {now.Add(-10 * time.Minute), nil, nil},
	} {
		t.Run(name, func(t *testing.T) {
			s, mock := setup(t)
			expectRatingRead(mock, ratedRows(test.started, test.ended, nil))
			if test.want == nil {
				mock.ExpectExec(`UPDATE "Playback" SET rating`).WillReturnResult(sqlmock.NewResult(0, 1))
				mock.ExpectCommit()
			} else {
				mock.ExpectRollback()
			}
			_, err := s.RatePlayback(context.Background(), 1, 2, "helped")
			if test.want == nil {
				require.NoError(t, err)
			} else {
				require.ErrorIs(t, err, test.want)
			}
		})
	}
}

func TestRatePlaybackRefusals(t *testing.T) {
	for _, test := range []struct {
		id, cp int
		rating string
	}{{0, 2, "helped"}, {1, 0, "helped"}, {1, 2, "great"}, {1, 2, ""}} {
		s, _ := setup(t)
		_, err := s.RatePlayback(context.Background(), test.id, test.cp, test.rating)
		require.ErrorIs(t, err, ErrInvalidInput)
	}

	t.Run("not found", func(t *testing.T) {
		s, mock := setup(t)
		mock.ExpectBegin()
		mock.ExpectQuery(`FOR UPDATE`).WithArgs(1, 2).WillReturnError(sql.ErrNoRows)
		mock.ExpectRollback()
		_, err := s.RatePlayback(context.Background(), 1, 2, "helped")
		require.ErrorIs(t, err, ErrNotFound)
	})
	t.Run("never reached the screen", func(t *testing.T) {
		s, mock := setup(t)
		expectRatingRead(mock, playbackRows("expired", time.Now().Add(-time.Hour)))
		mock.ExpectRollback()
		_, err := s.RatePlayback(context.Background(), 1, 2, "helped")
		require.ErrorIs(t, err, ErrNotRateable)
	})
}

func timePtr(t time.Time) *time.Time { return &t }

func TestGetPlayback(t *testing.T) {
	s, mock := setup(t)
	mock.ExpectQuery(`SELECT .* FROM "Playback" WHERE id = \$1 AND carepartner_id = \$2`).WithArgs(1, 2).
		WillReturnRows(playbackRows("finished", time.Now().Add(time.Hour)))
	p, err := s.GetPlayback(context.Background(), 1, 2)
	require.NoError(t, err)
	require.Equal(t, "finished", p.Status)
	require.NotNil(t, p.RatingClosesAt())
}

func TestGetPlaybackRefusals(t *testing.T) {
	s, _ := setup(t)
	_, err := s.GetPlayback(context.Background(), 0, 2)
	require.ErrorIs(t, err, ErrInvalidInput)
	_, err = s.GetPlayback(context.Background(), 1, 0)
	require.ErrorIs(t, err, ErrInvalidInput)

	s, mock := setup(t)
	mock.ExpectQuery(`FROM "Playback"`).WithArgs(1, 2).WillReturnError(sql.ErrNoRows)
	_, err = s.GetPlayback(context.Background(), 1, 2)
	require.ErrorIs(t, err, ErrNotFound)
}
