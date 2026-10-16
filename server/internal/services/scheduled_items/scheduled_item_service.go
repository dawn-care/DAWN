package scheduled_items

import (
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/lib/pq"

	mediaModel "server/internal/models/media"
	model "server/internal/models/scheduled_items"
	settingsModel "server/internal/models/screen_settings"
)

// MaxDisplaySeconds is the longest a timed photo can stay on screen (10 minutes)
const MaxDisplaySeconds = 600

// Delete scopes
const (
	ScopeSeries     = "series"     // delete every time
	ScopeOccurrence = "occurrence" // delete just this day
)

var (
	ErrInvalidCarepartner = errors.New("invalid carepartner_id")
	ErrCarepartnerMissing = errors.New("carepartner not found")
	ErrNotFound           = errors.New("scheduled item not found")
	ErrInvalidRepeat      = errors.New("repeat must be once, daily, weekly or monthly")
	ErrInvalidDate        = errors.New("invalid date")
	ErrInvalidTime        = errors.New("start_time must be HH:MM")
	ErrNotOn15MinStep     = errors.New("start_time must be on a 15 minute step (:00, :15, :30 or :45)")
	ErrOutsideAwakeHours  = errors.New("the item must start at or after wake time and finish by bed time")
	ErrInvalidMedia       = errors.New("media not found for this carepartner")
	ErrInvalidShape       = errors.New("invalid shape")
	ErrInvalidScope       = errors.New("scope must be series or occurrence")
	ErrNotAnOccurrence    = errors.New("the item doesn't occur on that date")
	ErrInvalidRange       = fmt.Errorf("from and to must be dates with from <= to, at most %d days apart", MaxCalendarDays)
)

type ScheduledItemService struct {
	DB *sql.DB
}

func NewScheduledItemService(db *sql.DB) *ScheduledItemService {
	return &ScheduledItemService{DB: db}
}

/*
ItemInput creates or replaces a scheduled item. The allowed shapes are:
video (no background, no display_seconds), audio (optional background_photo_id),
and photo (display_seconds 1-600, no background). Dates are "YYYY-MM-DD" and
start_time is "HH:MM" on a 15 minute step, in the carepartner's local time.
*/
type ItemInput struct {
	CarepartnerID     int     `json:"carepartner_id"` // ignored on update
	MediaID           int     `json:"media_id"`
	BackgroundPhotoID *int    `json:"background_photo_id"`
	DisplaySeconds    *int    `json:"display_seconds"`
	StartDate         string  `json:"start_date"`
	StartTime         string  `json:"start_time"`
	Repeat            string  `json:"repeat"`
	EndDate           *string `json:"end_date"`
}

func wrap(base error, format string, args ...any) error {
	return fmt.Errorf("%w: %s", base, fmt.Sprintf(format, args...))
}

// minutes converts "HH:MM" to minutes after midnight
func minutes(hhmm string) int {
	t, err := time.Parse("15:04", hhmm)
	if err != nil {
		return 0
	}
	return t.Hour()*60 + t.Minute()
}

// outsideAwakeHours reports whether an item starting at start and lasting duration seconds
// starts before wake or finishes after bed
func outsideAwakeHours(start string, durationSeconds int, wake, bed string) bool {
	startSec := minutes(start) * 60
	return startSec < minutes(wake)*60 || startSec+durationSeconds > minutes(bed)*60
}

// awakeHours returns the carepartner's wake and bed times
func (s *ScheduledItemService) awakeHours(carepartnerID int) (string, string, error) {
	settings, err := settingsModel.GetScreenSettings(s.DB, carepartnerID)
	if err != nil {
		return "", "", fmt.Errorf("failed to get screen settings: %w", err)
	}
	if settings == nil {
		return "", "", ErrCarepartnerMissing
	}
	return settings.WakeTime, settings.BedTime, nil
}

// carepartnerMedia returns a media item if it belongs to the carepartner
func (s *ScheduledItemService) carepartnerMedia(id, carepartnerID int, field string) (*mediaModel.Media, error) {
	m, err := mediaModel.GetMediaByID(s.DB, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get media: %w", err)
	}
	if m == nil || m.CarepartnerID != carepartnerID {
		return nil, wrap(ErrInvalidMedia, "%s %d", field, id)
	}
	return m, nil
}

/*
validate checks an input for a carepartner and returns the item to store, with its date
and time normalized. It checks, in order: repeat, dates, time and 15 minute step, media
and shape, then awake hours (which need the media's length).
*/
func (s *ScheduledItemService) validate(carepartnerID int, in ItemInput) (*model.ScheduledItem, error) {
	switch in.Repeat {
	case model.RepeatOnce, model.RepeatDaily, model.RepeatWeekly, model.RepeatMonthly:
	default:
		return nil, ErrInvalidRepeat
	}

	start, err := parseDate(in.StartDate)
	if err != nil {
		return nil, wrap(ErrInvalidDate, "start_date must be YYYY-MM-DD")
	}
	var endDate *string
	if in.EndDate != nil && *in.EndDate != "" {
		if in.Repeat == model.RepeatOnce {
			return nil, wrap(ErrInvalidDate, "items that happen once have no end_date")
		}
		end, err := parseDate(*in.EndDate)
		if err != nil {
			return nil, wrap(ErrInvalidDate, "end_date must be YYYY-MM-DD")
		}
		if end.Before(start) {
			return nil, wrap(ErrInvalidDate, "end_date is before start_date")
		}
		e := end.Format(DateLayout)
		endDate = &e
	}

	t, err := time.Parse("15:04", in.StartTime)
	if err != nil {
		return nil, ErrInvalidTime
	}
	if t.Minute()%15 != 0 {
		return nil, ErrNotOn15MinStep
	}

	m, err := s.carepartnerMedia(in.MediaID, carepartnerID, "media")
	if err != nil {
		return nil, err
	}
	switch m.Kind {
	case mediaModel.KindVideo:
		if in.BackgroundPhotoID != nil || in.DisplaySeconds != nil {
			return nil, wrap(ErrInvalidShape, "a video has no background photo or display_seconds")
		}
	case mediaModel.KindAudio:
		if in.DisplaySeconds != nil {
			return nil, wrap(ErrInvalidShape, "audio plays for its own length; display_seconds is only for photos")
		}
		if in.BackgroundPhotoID != nil {
			bg, err := s.carepartnerMedia(*in.BackgroundPhotoID, carepartnerID, "background photo")
			if err != nil {
				return nil, err
			}
			if bg.Kind != mediaModel.KindPhoto {
				return nil, wrap(ErrInvalidShape, "background_photo_id is a %s, not a photo", bg.Kind)
			}
		}
	case mediaModel.KindPhoto:
		if in.BackgroundPhotoID != nil {
			return nil, wrap(ErrInvalidShape, "a photo has no background photo")
		}
		if in.DisplaySeconds == nil || *in.DisplaySeconds < 1 || *in.DisplaySeconds > MaxDisplaySeconds {
			return nil, wrap(ErrInvalidShape, "a photo needs display_seconds from 1 to %d", MaxDisplaySeconds)
		}
	}

	item := &model.ScheduledItem{
		CarepartnerID:     carepartnerID,
		MediaID:           in.MediaID,
		BackgroundPhotoID: in.BackgroundPhotoID,
		DisplaySeconds:    in.DisplaySeconds,
		StartDate:         start.Format(DateLayout),
		StartTime:         t.Format("15:04"),
		Repeat:            in.Repeat,
		EndDate:           endDate,
	}
	item.Media.DurationSeconds = m.DurationSeconds

	wake, bed, err := s.awakeHours(carepartnerID)
	if err != nil {
		return nil, err
	}
	if outsideAwakeHours(item.StartTime, item.DurationSeconds(), wake, bed) {
		return nil, wrap(ErrOutsideAwakeHours, "awake hours are %s to %s", wake, bed)
	}
	return item, nil
}

// isMissingMedia reports whether err is a foreign key violation, e.g. media deleted meanwhile
func isMissingMedia(err error) bool {
	var pqErr *pq.Error
	return errors.As(err, &pqErr) && pqErr.Code == "23503"
}

// ─────────────────────────────────────────────
// CREATE
// ─────────────────────────────────────────────

func (s *ScheduledItemService) CreateItem(in ItemInput) (*model.ScheduledItem, error) {
	if in.CarepartnerID <= 0 {
		return nil, ErrInvalidCarepartner
	}
	item, err := s.validate(in.CarepartnerID, in)
	if err != nil {
		return nil, err
	}
	if err := model.CreateScheduledItem(s.DB, item); err != nil {
		if isMissingMedia(err) {
			return nil, ErrInvalidMedia
		}
		return nil, fmt.Errorf("failed to create scheduled item: %w", err)
	}
	return s.GetItem(item.ID)
}

// ─────────────────────────────────────────────
// READ
// ─────────────────────────────────────────────

func (s *ScheduledItemService) GetItem(id int) (*model.ScheduledItem, error) {
	if id <= 0 {
		return nil, ErrNotFound
	}
	item, err := model.GetScheduledItemByID(s.DB, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get scheduled item: %w", err)
	}
	if item == nil {
		return nil, ErrNotFound
	}
	wake, bed, err := s.awakeHours(item.CarepartnerID)
	if err != nil {
		return nil, err
	}
	item.OutsideAwakeHours = outsideAwakeHours(item.StartTime, item.DurationSeconds(), wake, bed)
	return item, nil
}

// ListItems returns a carepartner's scheduled items as rules (not expanded), by start time
func (s *ScheduledItemService) ListItems(carepartnerID int) ([]model.ScheduledItem, error) {
	if carepartnerID <= 0 {
		return nil, ErrInvalidCarepartner
	}
	wake, bed, err := s.awakeHours(carepartnerID)
	if err != nil {
		return nil, err
	}
	items, err := model.ListScheduledItems(s.DB, carepartnerID, "", "")
	if err != nil {
		return nil, fmt.Errorf("failed to list scheduled items: %w", err)
	}
	for i := range items {
		items[i].OutsideAwakeHours = outsideAwakeHours(items[i].StartTime, items[i].DurationSeconds(), wake, bed)
	}
	return items, nil
}

/*
Calendar expands a carepartner's items into dated occurrences from from to to (inclusive,
"YYYY-MM-DD"), leaving out skipped days. outside_awake_hours is computed against the
current awake hours, so changing them never requires editing stored items.
*/
func (s *ScheduledItemService) Calendar(carepartnerID int, from, to string) ([]model.Occurrence, error) {
	if carepartnerID <= 0 {
		return nil, ErrInvalidCarepartner
	}
	fromDate, err1 := parseDate(from)
	toDate, err2 := parseDate(to)
	if err1 != nil || err2 != nil || toDate.Before(fromDate) || toDate.Sub(fromDate) >= MaxCalendarDays*24*time.Hour {
		return nil, ErrInvalidRange
	}
	wake, bed, err := s.awakeHours(carepartnerID)
	if err != nil {
		return nil, err
	}
	items, err := model.ListScheduledItems(s.DB, carepartnerID, fromDate.Format(DateLayout), toDate.Format(DateLayout))
	if err != nil {
		return nil, fmt.Errorf("failed to list scheduled items: %w", err)
	}

	occurrences := []model.Occurrence{}
	for _, item := range items {
		start, err := parseDate(item.StartDate)
		if err != nil {
			return nil, err
		}
		var end *time.Time
		if item.EndDate != nil {
			e, err := parseDate(*item.EndDate)
			if err != nil {
				return nil, err
			}
			end = &e
		}
		duration := item.DurationSeconds()
		outside := outsideAwakeHours(item.StartTime, duration, wake, bed)
		for _, date := range OccurrenceDates(start, item.Repeat, end, item.SkipDates, fromDate, toDate) {
			occurrences = append(occurrences, model.Occurrence{
				ScheduledItemID:   item.ID,
				Date:              date,
				Time:              item.StartTime,
				Repeat:            item.Repeat,
				Shape:             item.Shape,
				Media:             item.Media,
				BackgroundPhoto:   item.BackgroundPhoto,
				DisplaySeconds:    item.DisplaySeconds,
				DurationSeconds:   duration,
				OutsideAwakeHours: outside,
			})
		}
	}
	sort.SliceStable(occurrences, func(a, b int) bool {
		oa, ob := occurrences[a], occurrences[b]
		if oa.Date != ob.Date {
			return oa.Date < ob.Date
		}
		if oa.Time != ob.Time {
			return oa.Time < ob.Time
		}
		return oa.ScheduledItemID < ob.ScheduledItemID
	})
	return occurrences, nil
}

// ─────────────────────────────────────────────
// UPDATE
// ─────────────────────────────────────────────

// UpdateItem replaces an item's media, timing and repeat. Skipped days are kept.
func (s *ScheduledItemService) UpdateItem(id int, in ItemInput) (*model.ScheduledItem, error) {
	if id <= 0 {
		return nil, ErrNotFound
	}
	existing, err := model.GetScheduledItemByID(s.DB, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get scheduled item: %w", err)
	}
	if existing == nil {
		return nil, ErrNotFound
	}
	item, err := s.validate(existing.CarepartnerID, in)
	if err != nil {
		return nil, err
	}
	item.ID = id
	found, err := model.UpdateScheduledItem(s.DB, item)
	if err != nil {
		if isMissingMedia(err) {
			return nil, ErrInvalidMedia
		}
		return nil, fmt.Errorf("failed to update scheduled item: %w", err)
	}
	if !found {
		return nil, ErrNotFound
	}
	return s.GetItem(id)
}

// ─────────────────────────────────────────────
// DELETE
// ─────────────────────────────────────────────

/*
DeleteItem deletes every time (scope "series") or just one day (scope "occurrence" with
date). Deleting the only day of an item that happens once deletes the item.
*/
func (s *ScheduledItemService) DeleteItem(id int, scope, date string) error {
	if id <= 0 {
		return ErrNotFound
	}
	switch scope {
	case ScopeSeries:
		found, err := model.DeleteScheduledItem(s.DB, id)
		if err != nil {
			return fmt.Errorf("failed to delete scheduled item: %w", err)
		}
		if !found {
			return ErrNotFound
		}
		return nil

	case ScopeOccurrence:
		day, err := parseDate(date)
		if err != nil {
			return wrap(ErrInvalidDate, "date must be YYYY-MM-DD")
		}
		item, err := model.GetScheduledItemByID(s.DB, id)
		if err != nil {
			return fmt.Errorf("failed to get scheduled item: %w", err)
		}
		if item == nil {
			return ErrNotFound
		}
		start, _ := parseDate(item.StartDate)
		var end *time.Time
		if item.EndDate != nil {
			e, _ := parseDate(*item.EndDate)
			end = &e
		}
		if !OccursOn(start, item.Repeat, end, day) {
			return wrap(ErrNotAnOccurrence, "%s", day.Format(DateLayout))
		}
		if item.Repeat == model.RepeatOnce {
			_, err = model.DeleteScheduledItem(s.DB, id)
		} else {
			err = model.CreateSkip(s.DB, id, day.Format(DateLayout))
		}
		if err != nil {
			return fmt.Errorf("failed to delete occurrence: %w", err)
		}
		return nil

	default:
		return ErrInvalidScope
	}
}
