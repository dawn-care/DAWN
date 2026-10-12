package screen_settings

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"server/internal/models/screen_settings"
)

// Region is an area of a layout preset with a fixed number of widget slots,
// e.g. the three tiles in the middle row of Balanced
type Region struct {
	Name    string   `json:"name"`
	Slots   int      `json:"slots"`
	Accepts []string `json:"accepts"` // widget types that can be placed here
}

// Preset is a screen layout the caregiver fills slot by slot
type Preset struct {
	Name    string   `json:"name"`
	Regions []Region `json:"regions"`
}

// Presets are the layouts a carepartner's screen can use. The database doesn't
// check layout_preset, so a new preset only needs adding here.
var Presets = []Preset{
	{
		// The carepartner home screen mockup: a large photo frame on the right,
		// three tiles in the middle row and a plate along the bottom
		Name: "balanced",
		Regions: []Region{
			{Name: "photo_frame", Slots: 1, Accepts: []string{"photo"}},
			{Name: "tiles", Slots: 3, Accepts: []string{"reminder", "weather", "sports", "news"}},
			{Name: "plate", Slots: 1, Accepts: []string{"message"}},
		},
	},
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

// Region returns the region with this name
func (p Preset) Region(name string) (Region, bool) {
	for _, r := range p.Regions {
		if r.Name == name {
			return r, true
		}
	}
	return Region{}, false
}

// HasSlot reports whether the preset has this region and slot
func (p Preset) HasSlot(region string, slot int) bool {
	r, ok := p.Region(region)
	return ok && slot >= 0 && slot < r.Slots
}

// AcceptsType reports whether a widget type can be placed in this region
func (r Region) AcceptsType(widgetType string) bool {
	for _, t := range r.Accepts {
		if t == widgetType {
			return true
		}
	}
	return false
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

// GetPresets returns every layout preset and its regions
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
