package screen_settings

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"server/internal/models/screen_settings"
)

// LayoutPresets are the allowed layout_preset values. The database doesn't check
// them, so a new preset only needs adding here.
var LayoutPresets = []string{"balanced"}

// Allowed awake hours, inclusive. The database enforces the same ranges.
const (
	EarliestWake = "06:00"
	LatestWake   = "10:00"
	EarliestBed  = "19:00"
	LatestBed    = "22:00"
)

var (
	ErrInvalidCarepartner = errors.New("invalid carepartner_id")
	ErrNotFound           = errors.New("screen settings not found")
	ErrInvalidPreset      = fmt.Errorf("layout_preset must be one of %v", LayoutPresets)
	ErrInvalidWakeTime    = fmt.Errorf("wake_time must be HH:MM between %s and %s", EarliestWake, LatestWake)
	ErrInvalidBedTime     = fmt.Errorf("bed_time must be HH:MM between %s and %s", EarliestBed, LatestBed)
)

type ScreenSettingsService struct {
	DB *sql.DB
}

func NewScreenSettingsService(db *sql.DB) *ScreenSettingsService {
	return &ScreenSettingsService{DB: db}
}

// UpdateInput replaces every setting; all fields are required
type UpdateInput struct {
	LayoutPreset string `json:"layout_preset"`
	WakeTime     string `json:"wake_time"`
	BedTime      string `json:"bed_time"`
}

func validPreset(preset string) bool {
	for _, p := range LayoutPresets {
		if p == preset {
			return true
		}
	}
	return false
}

// normalizeTime parses "HH:MM" and returns it zero-padded if it falls in [earliest, latest]
func normalizeTime(value, earliest, latest string) (string, bool) {
	t, err := time.Parse("15:04", value)
	if err != nil {
		return "", false
	}
	normalized := t.Format("15:04")
	// Zero-padded HH:MM strings compare in time order
	return normalized, normalized >= earliest && normalized <= latest
}

// ─────────────────────────────────────────────
// READ
// ─────────────────────────────────────────────

func (s *ScreenSettingsService) GetScreenSettings(carepartnerID int) (*screen_settings.ScreenSettings, error) {
	if carepartnerID <= 0 {
		return nil, ErrInvalidCarepartner
	}
	settings, err := screen_settings.GetScreenSettings(s.DB, carepartnerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get screen settings: %w", err)
	}
	if settings == nil {
		return nil, ErrNotFound
	}
	return settings, nil
}

// ─────────────────────────────────────────────
// UPDATE
// ─────────────────────────────────────────────

func (s *ScreenSettingsService) UpdateScreenSettings(carepartnerID int, in UpdateInput) (*screen_settings.ScreenSettings, error) {
	if carepartnerID <= 0 {
		return nil, ErrInvalidCarepartner
	}
	if !validPreset(in.LayoutPreset) {
		return nil, ErrInvalidPreset
	}
	wake, ok := normalizeTime(in.WakeTime, EarliestWake, LatestWake)
	if !ok {
		return nil, ErrInvalidWakeTime
	}
	bed, ok := normalizeTime(in.BedTime, EarliestBed, LatestBed)
	if !ok {
		return nil, ErrInvalidBedTime
	}

	settings := &screen_settings.ScreenSettings{
		CarepartnerID: carepartnerID,
		LayoutPreset:  in.LayoutPreset,
		WakeTime:      wake,
		BedTime:       bed,
	}
	found, err := screen_settings.UpdateScreenSettings(s.DB, settings)
	if err != nil {
		return nil, fmt.Errorf("failed to update screen settings: %w", err)
	}
	if !found {
		return nil, ErrNotFound
	}
	return settings, nil
}
