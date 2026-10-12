package screen_widgets

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"sort"

	"github.com/lib/pq"

	mediaModel "server/internal/models/media"
	settingsModel "server/internal/models/screen_settings"
	model "server/internal/models/screen_widgets"
	settings "server/internal/services/screen_settings"
)

var (
	ErrInvalidCarepartner = errors.New("invalid carepartner_id")
	ErrCarepartnerMissing = errors.New("carepartner not found")
	ErrNotFound           = errors.New("screen widget not found")
	ErrInvalidType        = fmt.Errorf("type must be one of %v", WidgetTypes)
	ErrInvalidConfig      = errors.New("invalid config")
	ErrInvalidPhoto       = errors.New("invalid photo")
	ErrUnknownSlot        = errors.New("the current layout has no such region or slot")
	ErrTypeNotAllowed     = errors.New("this widget type can't be placed in that region")
	ErrSlotTaken          = errors.New("that slot already has a widget")
)

type ScreenWidgetService struct {
	DB *sql.DB
}

func NewScreenWidgetService(db *sql.DB) *ScreenWidgetService {
	return &ScreenWidgetService{DB: db}
}

// CreateInput places a new widget in a slot of the carepartner's current preset
type CreateInput struct {
	CarepartnerID int             `json:"carepartner_id"`
	Type          string          `json:"type"`
	Region        string          `json:"region"`
	Slot          *int            `json:"slot"`
	Config        json.RawMessage `json:"config" swaggertype:"object"`
	PhotoMediaID  *int            `json:"photo_media_id"`
}

// UpdateInput replaces a widget's settings and photo; its type and slot don't change
type UpdateInput struct {
	Config       json.RawMessage `json:"config" swaggertype:"object"`
	PhotoMediaID *int            `json:"photo_media_id"`
}

// MoveInput names the slot to move a widget to
type MoveInput struct {
	Region string `json:"region"`
	Slot   *int   `json:"slot"`
}

func placementErr(base error, format string, args ...any) error {
	return fmt.Errorf("%w: %s", base, fmt.Sprintf(format, args...))
}

// isSlotConflict reports whether err is the unique (carepartner, region, slot) violation
func isSlotConflict(err error) bool {
	var pqErr *pq.Error
	return errors.As(err, &pqErr) && pqErr.Code == "23505"
}

// currentPreset returns the layout preset the carepartner's screen uses
func currentPreset(exec interface {
	QueryRow(query string, args ...any) *sql.Row
}, carepartnerID int) (settings.Preset, error) {
	s, err := settingsModel.GetScreenSettings(exec, carepartnerID)
	if err != nil {
		return settings.Preset{}, fmt.Errorf("failed to get screen settings: %w", err)
	}
	if s == nil {
		return settings.Preset{}, ErrCarepartnerMissing
	}
	preset, ok := settings.PresetByName(s.LayoutPreset)
	if !ok {
		return settings.Preset{}, fmt.Errorf("carepartner %d has unknown layout preset %q", carepartnerID, s.LayoutPreset)
	}
	return preset, nil
}

// checkPlacement checks that the preset has the slot and that the region accepts the type
func checkPlacement(preset settings.Preset, widgetType, region string, slot int) error {
	r, ok := preset.Region(region)
	if !ok || slot < 0 || slot >= r.Slots {
		return placementErr(ErrUnknownSlot, "%q has no %s slot %d", preset.Name, region, slot)
	}
	if !r.AcceptsType(widgetType) {
		return placementErr(ErrTypeNotAllowed, "%s accepts %v, not %s", region, r.Accepts, widgetType)
	}
	return nil
}

// checkPhoto checks a widget's photo: required for photo widgets, optional for
// message widgets, not allowed otherwise, and always a photo of the same carepartner
func checkPhoto(tx *sql.Tx, widgetType string, carepartnerID int, photoID *int) error {
	switch widgetType {
	case TypePhoto:
		if photoID == nil {
			return placementErr(ErrInvalidPhoto, "photo widgets need a photo_media_id")
		}
	case TypeMessage:
		if photoID == nil {
			return nil
		}
	default:
		if photoID != nil {
			return placementErr(ErrInvalidPhoto, "only photo and message widgets have a photo")
		}
		return nil
	}

	m, err := model.GetPhoto(tx, *photoID)
	if err != nil {
		return fmt.Errorf("failed to check photo: %w", err)
	}
	if m == nil || m.CarepartnerID != carepartnerID {
		return placementErr(ErrInvalidPhoto, "photo %d not found for this carepartner", *photoID)
	}
	if m.Kind != mediaModel.KindPhoto {
		return placementErr(ErrInvalidPhoto, "media %d is a %s, not a photo", *photoID, m.Kind)
	}
	return nil
}

// commit commits tx, reporting a lost race for a slot as ErrSlotTaken
func commit(tx *sql.Tx) error {
	if err := tx.Commit(); err != nil {
		if isSlotConflict(err) {
			return ErrSlotTaken
		}
		return err
	}
	return nil
}

// markHidden flags widgets whose slot the preset doesn't have
func markHidden(widgets []model.ScreenWidget, preset settings.Preset) {
	for i := range widgets {
		widgets[i].Hidden = !preset.HasSlot(widgets[i].Region, widgets[i].Slot)
	}
}

// ─────────────────────────────────────────────
// CREATE
// ─────────────────────────────────────────────

func (s *ScreenWidgetService) CreateWidget(in CreateInput) (*model.ScreenWidget, error) {
	if in.CarepartnerID <= 0 {
		return nil, ErrInvalidCarepartner
	}
	if !validType(in.Type) {
		return nil, ErrInvalidType
	}
	if in.Region == "" || in.Slot == nil {
		return nil, placementErr(ErrUnknownSlot, "region and slot are required")
	}
	config, err := NormalizeConfig(in.Type, in.Config)
	if err != nil {
		return nil, err
	}

	tx, err := s.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	preset, err := currentPreset(tx, in.CarepartnerID)
	if err != nil {
		return nil, err
	}
	if err := checkPlacement(preset, in.Type, in.Region, *in.Slot); err != nil {
		return nil, err
	}
	if err := checkPhoto(tx, in.Type, in.CarepartnerID, in.PhotoMediaID); err != nil {
		return nil, err
	}
	occupant, err := model.GetScreenWidgetAt(tx, in.CarepartnerID, in.Region, *in.Slot)
	if err != nil {
		return nil, fmt.Errorf("failed to check slot: %w", err)
	}
	if occupant != nil {
		return nil, ErrSlotTaken
	}

	w := &model.ScreenWidget{
		CarepartnerID: in.CarepartnerID,
		Type:          in.Type,
		Region:        in.Region,
		Slot:          *in.Slot,
		Config:        config,
		PhotoMediaID:  in.PhotoMediaID,
	}
	if err := model.CreateScreenWidget(tx, w); err != nil {
		if isSlotConflict(err) {
			return nil, ErrSlotTaken
		}
		return nil, fmt.Errorf("failed to create widget: %w", err)
	}
	if err := commit(tx); err != nil {
		return nil, err
	}
	return w, nil
}

// ─────────────────────────────────────────────
// READ
// ─────────────────────────────────────────────

/*
ListWidgets returns a carepartner's widgets: the ones in the current preset first,
in region and slot order, then hidden ones (slots the current preset doesn't have).
*/
func (s *ScreenWidgetService) ListWidgets(carepartnerID int) ([]model.ScreenWidget, error) {
	if carepartnerID <= 0 {
		return nil, ErrInvalidCarepartner
	}
	preset, err := currentPreset(s.DB, carepartnerID)
	if err != nil {
		return nil, err
	}
	widgets, err := model.ListScreenWidgets(s.DB, carepartnerID)
	if err != nil {
		return nil, fmt.Errorf("failed to list widgets: %w", err)
	}
	markHidden(widgets, preset)

	regionOrder := map[string]int{}
	for i, r := range preset.Regions {
		regionOrder[r.Name] = i
	}
	sort.SliceStable(widgets, func(a, b int) bool {
		wa, wb := widgets[a], widgets[b]
		if wa.Hidden != wb.Hidden {
			return !wa.Hidden
		}
		if wa.Hidden {
			return false // hidden widgets keep the database's region, slot order
		}
		if regionOrder[wa.Region] != regionOrder[wb.Region] {
			return regionOrder[wa.Region] < regionOrder[wb.Region]
		}
		return wa.Slot < wb.Slot
	})
	return widgets, nil
}

func (s *ScreenWidgetService) GetWidget(id int) (*model.ScreenWidget, error) {
	if id <= 0 {
		return nil, ErrNotFound
	}
	w, err := model.GetScreenWidgetByID(s.DB, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get widget: %w", err)
	}
	if w == nil {
		return nil, ErrNotFound
	}
	preset, err := currentPreset(s.DB, w.CarepartnerID)
	if err != nil {
		return nil, err
	}
	w.Hidden = !preset.HasSlot(w.Region, w.Slot)
	return w, nil
}

// ─────────────────────────────────────────────
// UPDATE
// ─────────────────────────────────────────────

// UpdateWidget replaces a widget's settings and photo
func (s *ScreenWidgetService) UpdateWidget(id int, in UpdateInput) (*model.ScreenWidget, error) {
	if id <= 0 {
		return nil, ErrNotFound
	}

	tx, err := s.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	w, err := model.GetScreenWidgetForUpdate(tx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get widget: %w", err)
	}
	if w == nil {
		return nil, ErrNotFound
	}
	config, err := NormalizeConfig(w.Type, in.Config)
	if err != nil {
		return nil, err
	}
	if err := checkPhoto(tx, w.Type, w.CarepartnerID, in.PhotoMediaID); err != nil {
		return nil, err
	}

	w.Config, w.PhotoMediaID = config, in.PhotoMediaID
	if err := model.UpdateScreenWidgetConfig(tx, w); err != nil {
		return nil, fmt.Errorf("failed to update widget: %w", err)
	}
	preset, err := currentPreset(tx, w.CarepartnerID)
	if err != nil {
		return nil, err
	}
	if err := commit(tx); err != nil {
		return nil, err
	}
	w.Hidden = !preset.HasSlot(w.Region, w.Slot)
	return w, nil
}

/*
MoveWidget moves a widget to another slot of the current preset. If that slot has a
widget, the two swap places, as long as the other widget can go in the moving widget's
old slot; otherwise the move fails with ErrSlotTaken.
*/
func (s *ScreenWidgetService) MoveWidget(id int, in MoveInput) (*model.ScreenWidget, error) {
	if id <= 0 {
		return nil, ErrNotFound
	}
	if in.Region == "" || in.Slot == nil {
		return nil, placementErr(ErrUnknownSlot, "region and slot are required")
	}

	tx, err := s.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	w, err := model.GetScreenWidgetForUpdate(tx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get widget: %w", err)
	}
	if w == nil {
		return nil, ErrNotFound
	}
	preset, err := currentPreset(tx, w.CarepartnerID)
	if err != nil {
		return nil, err
	}
	if err := checkPlacement(preset, w.Type, in.Region, *in.Slot); err != nil {
		return nil, err
	}
	if w.Region == in.Region && w.Slot == *in.Slot {
		return w, nil
	}

	occupant, err := model.GetScreenWidgetAt(tx, w.CarepartnerID, in.Region, *in.Slot)
	if err != nil {
		return nil, fmt.Errorf("failed to check slot: %w", err)
	}
	if occupant != nil {
		// Swap only if the other widget can go where this one was
		if checkPlacement(preset, occupant.Type, w.Region, w.Slot) != nil {
			return nil, placementErr(ErrSlotTaken, "the %s widget there can't move to %s slot %d", occupant.Type, w.Region, w.Slot)
		}
		if err := model.MoveScreenWidget(tx, occupant.ID, w.Region, w.Slot); err != nil {
			return nil, fmt.Errorf("failed to move widget: %w", err)
		}
	}
	if err := model.MoveScreenWidget(tx, w.ID, in.Region, *in.Slot); err != nil {
		return nil, fmt.Errorf("failed to move widget: %w", err)
	}
	if err := commit(tx); err != nil {
		return nil, err
	}

	w.Region, w.Slot, w.Hidden = in.Region, *in.Slot, false
	return w, nil
}

// ─────────────────────────────────────────────
// DELETE
// ─────────────────────────────────────────────

func (s *ScreenWidgetService) DeleteWidget(id int) error {
	if id <= 0 {
		return ErrNotFound
	}
	found, err := model.DeleteScreenWidget(s.DB, id)
	if err != nil {
		return fmt.Errorf("failed to delete widget: %w", err)
	}
	if !found {
		return ErrNotFound
	}
	return nil
}
