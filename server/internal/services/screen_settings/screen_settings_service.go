package screen_settings

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"server/internal/models/screen_settings"
)

var (
	ErrInvalidCarepartner = errors.New("invalid carepartner_id")
	ErrNotFound           = errors.New("screen settings not found")
	ErrInvalidWakeTime    = errors.New("wake_time must be HH:MM")
	// Awake hours are within one day, so bed time can't be at or before wake time.
	// The database checks the same.
	ErrInvalidBedTime = errors.New("bed_time must be HH:MM, later than wake_time")
)

type ScreenSettingsService struct {
	DB *sql.DB
}

func NewScreenSettingsService(db *sql.DB) *ScreenSettingsService {
	return &ScreenSettingsService{DB: db}
}

// UpdateInput replaces every setting; all fields are required
type UpdateInput struct {
	WakeTime string `json:"wake_time"`
	BedTime  string `json:"bed_time"`
}

// normalizeTime parses "HH:MM" and returns it zero-padded
func normalizeTime(value string) (string, bool) {
	t, err := time.Parse("15:04", value)
	if err != nil {
		return "", false
	}
	return t.Format("15:04"), true
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
	wake, ok := normalizeTime(in.WakeTime)
	if !ok {
		return nil, ErrInvalidWakeTime
	}
	bed, ok := normalizeTime(in.BedTime)
	// Zero-padded HH:MM strings compare in time order
	if !ok || bed <= wake {
		return nil, ErrInvalidBedTime
	}

	settings := &screen_settings.ScreenSettings{
		CarepartnerID: carepartnerID,
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
