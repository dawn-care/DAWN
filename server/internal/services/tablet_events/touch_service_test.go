package tablet_events

import (
	"context"
	"encoding/json"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/lib/pq"
	"github.com/stretchr/testify/require"

	model "server/internal/models/tablet_events"
	"server/internal/utils"
)

var qInsert = regexp.QuoteMeta(`INSERT INTO "TabletTouch"`)

func str(s string) *string { return &s }

var (
	first = time.Date(2026, 10, 18, 14, 0, 0, 0, time.UTC)
	last  = first.Add(2800 * time.Millisecond)
)

func widgetTouch() model.TouchEvent {
	return model.TouchEvent{Screen: "always_on", TargetType: str("widget"), WidgetType: str("weather"), WidgetID: str("12"),
		Gesture: "tap", Count: 3, FirstAt: first, LastAt: last}
}

func emptyTouch() model.TouchEvent {
	return model.TouchEvent{Screen: "default", TargetType: str("empty"), Gesture: "long_press", Count: 1, FirstAt: first, LastAt: first}
}

func mediaTouch() model.TouchEvent {
	return model.TouchEvent{Screen: "playing", TargetType: str("media"), MediaKind: str("video"), Gesture: "swipe", Count: 1, FirstAt: first, LastAt: first}
}

func nightTouch() model.TouchEvent {
	return model.TouchEvent{Screen: "night", Gesture: "tap", Count: 2, FirstAt: first, LastAt: last}
}

func newService(t *testing.T) (*TouchService, sqlmock.Sqlmock) {
	db, mock := utils.SetupMockDB(t)
	t.Cleanup(func() {
		require.NoError(t, mock.ExpectationsWereMet())
		db.Close()
	})
	return NewTouchService(db), mock
}

// The tablet's own body (touchLog.ts toWire) decodes into TouchEvent, nulls included
func TestTouchEventDecodesTheTabletsBody(t *testing.T) {
	var e model.TouchEvent
	require.NoError(t, json.Unmarshal([]byte(`{"screen":"always_on","target_type":"widget","widget_type":"weather",
		"widget_id":"12","media_kind":null,"gesture":"tap","count":3,
		"first_at":"2026-10-18T14:00:00.000Z","last_at":"2026-10-18T14:00:02.800Z"}`), &e))
	require.Equal(t, widgetTouch(), e)
	require.NoError(t, ValidateTouch(e))

	var night model.TouchEvent
	require.NoError(t, json.Unmarshal([]byte(`{"screen":"night","target_type":null,"widget_type":null,"widget_id":null,
		"media_kind":null,"gesture":"tap","count":2,"first_at":"2026-10-18T14:00:00Z","last_at":"2026-10-18T14:00:02.800Z"}`), &night))
	require.Equal(t, nightTouch(), night)
}

func TestValidateTouch_Valid(t *testing.T) {
	for name, e := range map[string]model.TouchEvent{
		"widget": widgetTouch(), "empty": emptyTouch(), "media": mediaTouch(), "no target": nightTouch(),
	} {
		require.NoError(t, ValidateTouch(e), name)
	}
}

func TestValidateTouch_Invalid(t *testing.T) {
	cases := map[string]func(e *model.TouchEvent){
		"pair screen":               func(e *model.TouchEvent) { e.Screen = "pair" },
		"no screen":                 func(e *model.TouchEvent) { e.Screen = "" },
		"unknown gesture":           func(e *model.TouchEvent) { e.Gesture = "pinch" },
		"no gesture":                func(e *model.TouchEvent) { e.Gesture = "" },
		"count 0":                   func(e *model.TouchEvent) { e.Count = 0 },
		"negative count":            func(e *model.TouchEvent) { e.Count = -1 },
		"no first_at":               func(e *model.TouchEvent) { e.FirstAt = time.Time{} },
		"no last_at":                func(e *model.TouchEvent) { e.LastAt = time.Time{} },
		"first_at after last_at":    func(e *model.TouchEvent) { e.FirstAt = e.LastAt.Add(time.Millisecond) },
		"unknown target_type":       func(e *model.TouchEvent) { e.TargetType = str("button") },
		"empty-string target_type":  func(e *model.TouchEvent) { e.TargetType = str("") },
		"unknown widget_type":       func(e *model.TouchEvent) { e.WidgetType = str("clock") },
		"no widget_type":            func(e *model.TouchEvent) { e.WidgetType = nil },
		"no widget_id":              func(e *model.TouchEvent) { e.WidgetID = nil },
		"widget_id not a number":    func(e *model.TouchEvent) { e.WidgetID = str("abc") },
		"widget_id zero":            func(e *model.TouchEvent) { e.WidgetID = str("0") },
		"widget_id leading zero":    func(e *model.TouchEvent) { e.WidgetID = str("012") },
		"widget_id too long":        func(e *model.TouchEvent) { e.WidgetID = str("12345678901") },
		"widget with media_kind":    func(e *model.TouchEvent) { e.MediaKind = str("video") },
		"media without media_kind":  func(e *model.TouchEvent) { *e = mediaTouch(); e.MediaKind = nil },
		"unknown media_kind":        func(e *model.TouchEvent) { *e = mediaTouch(); e.MediaKind = str("gif") },
		"media with widget_type":    func(e *model.TouchEvent) { *e = mediaTouch(); e.WidgetType = str("weather") },
		"media with widget_id":      func(e *model.TouchEvent) { *e = mediaTouch(); e.WidgetID = str("12") },
		"empty with widget_type":    func(e *model.TouchEvent) { *e = emptyTouch(); e.WidgetType = str("weather") },
		"empty with media_kind":     func(e *model.TouchEvent) { *e = emptyTouch(); e.MediaKind = str("audio") },
		"no target with widget_id":  func(e *model.TouchEvent) { *e = nightTouch(); e.WidgetID = str("12") },
		"no target with media_kind": func(e *model.TouchEvent) { *e = nightTouch(); e.MediaKind = str("audio") },
	}
	for name, change := range cases {
		t.Run(name, func(t *testing.T) {
			e := widgetTouch()
			change(&e)
			require.ErrorIs(t, ValidateTouch(e), ErrInvalidInput)
		})
	}
}

func TestRecordTouches_StoresTheBatchInOneTransaction(t *testing.T) {
	s, mock := newService(t)
	// Times are stored in UTC whatever zone they were sent in
	night := nightTouch()
	night.FirstAt = night.FirstAt.In(time.FixedZone("EDT", -4*3600))

	mock.ExpectBegin()
	mock.ExpectExec(qInsert).WithArgs(7, "always_on", str("widget"), str("weather"), str("12"), nil, "tap", 3, first, last).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(qInsert).WithArgs(7, "playing", str("media"), nil, nil, str("video"), "swipe", 1, first, first).
		WillReturnResult(sqlmock.NewResult(2, 1))
	mock.ExpectExec(qInsert).WithArgs(7, "night", nil, nil, nil, nil, "tap", 2, first, last).
		WillReturnResult(sqlmock.NewResult(3, 1))
	mock.ExpectCommit()

	require.NoError(t, s.RecordTouches(context.Background(), 7, []model.TouchEvent{widgetTouch(), mediaTouch(), night}))
}

// A resent event hits the unique key and is skipped (0 rows); the batch still succeeds
func TestRecordTouches_ResentEventIsSkipped(t *testing.T) {
	s, mock := newService(t)
	mock.ExpectBegin()
	mock.ExpectExec(qInsert + `(?s).*ON CONFLICT ON CONSTRAINT "TabletTouch_event_key" DO NOTHING`).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit()

	require.NoError(t, s.RecordTouches(context.Background(), 7, []model.TouchEvent{widgetTouch()}))
}

func TestRecordTouches_FullBatch(t *testing.T) {
	s, mock := newService(t)
	events := make([]model.TouchEvent, MaxBatch)
	mock.ExpectBegin()
	for i := range events {
		events[i] = emptyTouch()
		mock.ExpectExec(qInsert).WillReturnResult(sqlmock.NewResult(int64(i+1), 1))
	}
	mock.ExpectCommit()

	require.NoError(t, s.RecordTouches(context.Background(), 7, events))
}

// Bad input is refused before any query runs: no transaction is started
func TestRecordTouches_InvalidBatchTouchesNoDatabase(t *testing.T) {
	bad := widgetTouch()
	bad.Count = 0
	tooMany := make([]model.TouchEvent, MaxBatch+1)
	for i := range tooMany {
		tooMany[i] = emptyTouch()
	}
	cases := map[string]struct {
		carepartnerID int
		events        []model.TouchEvent
	}{
		"carepartner 0":            {0, []model.TouchEvent{widgetTouch()}},
		"negative carepartner":     {-1, []model.TouchEvent{widgetTouch()}},
		"no events":                {7, nil},
		"empty events":             {7, []model.TouchEvent{}},
		"51 events":                {7, tooMany},
		"one bad event among good": {7, []model.TouchEvent{widgetTouch(), mediaTouch(), bad}},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			s, _ := newService(t) // no expectations: any query fails ExpectationsWereMet
			require.ErrorIs(t, s.RecordTouches(context.Background(), c.carepartnerID, c.events), ErrInvalidInput)
		})
	}
}

// Refusals by the table can't succeed on a retry, so they are ErrInvalidInput (400)
// and nothing is committed
func TestRecordTouches_TableRefusalIsInvalidInput(t *testing.T) {
	for name, code := range map[string]pq.ErrorCode{"unknown carepartner": "23503", "check constraint": "23514"} {
		t.Run(name, func(t *testing.T) {
			s, mock := newService(t)
			mock.ExpectBegin()
			mock.ExpectExec(qInsert).WillReturnResult(sqlmock.NewResult(1, 1))
			mock.ExpectExec(qInsert).WillReturnError(&pq.Error{Code: code})
			mock.ExpectRollback()

			require.ErrorIs(t, s.RecordTouches(context.Background(), 999, []model.TouchEvent{widgetTouch(), emptyTouch()}), ErrInvalidInput)
		})
	}
}

// Anything else is a server error, so the tablet retries; it must not look like bad input
func TestRecordTouches_ServerErrors(t *testing.T) {
	lost := errors.New("connection lost")

	t.Run("begin", func(t *testing.T) {
		s, mock := newService(t)
		mock.ExpectBegin().WillReturnError(lost)
		err := s.RecordTouches(context.Background(), 7, []model.TouchEvent{widgetTouch()})
		require.ErrorIs(t, err, lost)
		require.NotErrorIs(t, err, ErrInvalidInput)
	})
	t.Run("insert", func(t *testing.T) {
		s, mock := newService(t)
		mock.ExpectBegin()
		mock.ExpectExec(qInsert).WillReturnError(lost)
		mock.ExpectRollback()
		err := s.RecordTouches(context.Background(), 7, []model.TouchEvent{widgetTouch()})
		require.ErrorIs(t, err, lost)
		require.NotErrorIs(t, err, ErrInvalidInput)
	})
	t.Run("other database error", func(t *testing.T) {
		s, mock := newService(t)
		mock.ExpectBegin()
		mock.ExpectExec(qInsert).WillReturnError(&pq.Error{Code: "57P01"}) // admin shutdown
		mock.ExpectRollback()
		require.NotErrorIs(t, s.RecordTouches(context.Background(), 7, []model.TouchEvent{widgetTouch()}), ErrInvalidInput)
	})
	t.Run("commit", func(t *testing.T) {
		s, mock := newService(t)
		mock.ExpectBegin()
		mock.ExpectExec(qInsert).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit().WillReturnError(lost)
		err := s.RecordTouches(context.Background(), 7, []model.TouchEvent{widgetTouch()})
		require.ErrorIs(t, err, lost)
		require.NotErrorIs(t, err, ErrInvalidInput)
	})
}
