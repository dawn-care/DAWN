package tablet_events

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"slices"

	model "server/internal/models/tablet_events"
	screenWidgets "server/internal/services/screen_widgets"

	"github.com/lib/pq"
)

// ErrInvalidInput means the batch can never be stored as sent. The controller
// answers 400, and the tablet drops the batch instead of retrying it.
var ErrInvalidInput = errors.New("invalid touch batch")

// MaxBatch is the most events one request may carry; the tablet sends at 50
// (BATCH_LIMIT in carepartner/src/telemetry/touchLog.ts).
const MaxBatch = 50

// widgetID matches a positive ScreenWidget id written as a string, as the CHECK on
// "TabletTouch".widget_id does
var widgetID = regexp.MustCompile(`^[1-9][0-9]{0,9}$`)

type TouchService struct {
	DB *sql.DB
}

func NewTouchService(db *sql.DB) *TouchService {
	return &TouchService{DB: db}
}

func invalid(format string, args ...any) error {
	return fmt.Errorf("%w: %s", ErrInvalidInput, fmt.Sprintf(format, args...))
}

/*
ValidateTouch checks one event against the same rules as the CHECK constraints on
"TabletTouch", so a bad batch is refused before any query runs. Widget fields only
for a widget, media_kind only for media, and none of them for an empty cell or when
there is no target.
*/
func ValidateTouch(e model.TouchEvent) error {
	if !slices.Contains(model.Screens, e.Screen) {
		return invalid("unknown screen")
	}
	if !slices.Contains(model.Gestures, e.Gesture) {
		return invalid("unknown gesture")
	}
	if e.Count < 1 {
		return invalid("count must be at least 1")
	}
	if e.FirstAt.IsZero() || e.LastAt.IsZero() {
		return invalid("first_at and last_at are required")
	}
	if e.FirstAt.After(e.LastAt) {
		return invalid("first_at is after last_at")
	}

	target := ""
	if e.TargetType != nil {
		target = *e.TargetType
		if !slices.Contains(model.TargetTypes, target) {
			return invalid("unknown target_type")
		}
	}
	hasWidget := e.WidgetType != nil || e.WidgetID != nil
	switch target {
	case model.TargetWidget:
		if e.WidgetType == nil || !slices.Contains(screenWidgets.WidgetTypes, *e.WidgetType) {
			return invalid("a widget touch needs a known widget_type")
		}
		if e.WidgetID == nil || !widgetID.MatchString(*e.WidgetID) {
			return invalid("a widget touch needs a widget_id")
		}
		if e.MediaKind != nil {
			return invalid("media_kind is only for media")
		}
	case model.TargetMedia:
		if e.MediaKind == nil || !slices.Contains(model.MediaKinds, *e.MediaKind) {
			return invalid("a media touch needs a known media_kind")
		}
		if hasWidget {
			return invalid("widget_type and widget_id are only for a widget")
		}
	default: // empty, or nothing to name
		if hasWidget || e.MediaKind != nil {
			return invalid("widget and media fields are only for a widget or media")
		}
	}
	return nil
}

/*
RecordTouches stores one batch from a carepartner's tablet in one transaction:
every event is saved, or none. Every event is checked first, so one bad event
refuses the whole batch with ErrInvalidInput.

An event already stored (same carepartner, screen, gesture and first_at) is skipped,
so a batch the tablet resends after a lost response isn't saved twice.

A carepartner that doesn't exist is ErrInvalidInput too: retrying can't fix it, and
a 5xx would make an unpaired tablet retry forever.
*/
func (s *TouchService) RecordTouches(ctx context.Context, carepartnerID int, events []model.TouchEvent) error {
	if carepartnerID <= 0 {
		return invalid("a positive carepartner_id is required")
	}
	if len(events) == 0 || len(events) > MaxBatch {
		return invalid("a batch has 1 to %d events", MaxBatch)
	}
	for i, e := range events {
		if err := ValidateTouch(e); err != nil {
			return fmt.Errorf("event %d: %w", i, err)
		}
	}

	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("record touches: %w", err)
	}
	defer tx.Rollback()
	for _, e := range events {
		_, err := tx.ExecContext(ctx, `
			INSERT INTO "TabletTouch" (carepartner_id, screen, target_type, widget_type, widget_id,
			                           media_kind, gesture, count, first_at, last_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
			ON CONFLICT ON CONSTRAINT "TabletTouch_event_key" DO NOTHING`,
			carepartnerID, e.Screen, e.TargetType, e.WidgetType, e.WidgetID,
			e.MediaKind, e.Gesture, e.Count, e.FirstAt.UTC(), e.LastAt.UTC())
		if err != nil {
			return insertError(err)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("record touches: %w", err)
	}
	return nil
}

/*
insertError turns a refusal by the table into ErrInvalidInput: a missing carepartner
(foreign key, 23503) or a row the CHECK constraints refuse (23514), which means
ValidateTouch has fallen behind the migration. Neither succeeds on a retry. Anything
else, like a lost connection, stays a server error so the tablet retries.
*/
func insertError(err error) error {
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		switch pqErr.Code {
		case "23503":
			return invalid("unknown carepartner")
		case "23514":
			return invalid("refused by the table's checks")
		}
	}
	return fmt.Errorf("record touches: %w", err)
}
