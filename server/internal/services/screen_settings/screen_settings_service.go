package screen_settings

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"server/internal/models/screen_settings"
)

/*
Preset is a screen layout with a number of widget spots. Spots are 0-based and shared
by every preset: The widgets follow reading order. The
server only knows how many spots a preset has; the frontend and tablet decide where
each spot sits on screen. Any widget type can go in any spot.
*/
type Preset struct {
	Name  string `json:"name"`
	Spots int    `json:"spots"`
}

// Presets are the layouts a carepartner's screen can use. The database doesn't
// check layout_preset, so a new preset only needs adding here.
var Presets = []Preset{
	{Name: "simple", Spots: 2},
	{Name: "balanced", Spots: 4},
	{Name: "grid", Spots: 4},
	{Name: "busier", Spots: 6},
}

// LayoutPresets are the allowed layout_preset values
var LayoutPresets = presetNames()

func presetNames() []string {
	names := make([]string, len(Presets))
	for i, p := range Presets {
		names[i] = p.Name
	}
	return names
}

// PresetByName returns the preset with this name
func PresetByName(name string) (Preset, bool) {
	for _, p := range Presets {
		if p.Name == name {
			return p, true
		}
	}
	return Preset{}, false
}

// HasSpot reports whether the preset has this spot
func (p Preset) HasSpot(spot int) bool {
	return spot >= 0 && spot < p.Spots
}

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
	_, ok := PresetByName(preset)
	return ok
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

// GetPresets returns every layout preset and its spot count
func (s *ScreenSettingsService) GetPresets() []Preset {
	return Presets
}

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
