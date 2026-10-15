package screen_widgets

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/lib/pq"

	"server/internal/db"
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
	ErrUnknownSpot        = errors.New("the current layout has no such spot")
	ErrSpotTaken          = errors.New("that spot already has a widget")
)

type ScreenWidgetService struct {
	DB *sql.DB
}

func NewScreenWidgetService(db *sql.DB) *ScreenWidgetService {
	return &ScreenWidgetService{DB: db}
}

// CreateInput places a new widget in a spot of the carepartner's current preset
type CreateInput struct {
	CarepartnerID int             `json:"carepartner_id"`
	Type          string          `json:"type"`
	Spot          *int            `json:"spot"`
	Config        json.RawMessage `json:"config" swaggertype:"object"`
	PhotoMediaID  *int            `json:"photo_media_id"`
}

// UpdateInput replaces a widget's settings and photo; its type and spot don't change
type UpdateInput struct {
	Config       json.RawMessage `json:"config" swaggertype:"object"`
	PhotoMediaID *int            `json:"photo_media_id"`
}

// MoveInput names the spot to move a widget to
type MoveInput struct {
	Spot *int `json:"spot"`
}

func placementErr(base error, format string, args ...any) error {
	return fmt.Errorf("%w: %s", base, fmt.Sprintf(format, args...))
}

// isSpotConflict reports whether err is the unique (carepartner, spot) violation
func isSpotConflict(err error) bool {
	var pqErr *pq.Error
	return errors.As(err, &pqErr) && pqErr.Code == "23505"
}

// currentPreset returns the layout preset the carepartner's screen uses
func currentPreset(exec db.QueryRower, carepartnerID int) (settings.Preset, error) {
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

// checkPlacement checks that the preset has the spot; any widget type can go in any spot
func checkPlacement(preset settings.Preset, spot int) error {
	if !preset.HasSpot(spot) {
		return placementErr(ErrUnknownSpot, "%q has spots 0 to %d, not %d", preset.Name, preset.Spots-1, spot)
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

// commit commits tx, reporting a lost race for a spot as ErrSpotTaken
func commit(tx *sql.Tx) error {
	if err := tx.Commit(); err != nil {
		if isSpotConflict(err) {
			return ErrSpotTaken
		}
		return err
	}
	return nil
}

// markHidden flags widgets whose spot the preset doesn't have
func markHidden(widgets []model.ScreenWidget, preset settings.Preset) {
	for i := range widgets {
		widgets[i].Hidden = !preset.HasSpot(widgets[i].Spot)
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
	if in.Spot == nil {
		return nil, placementErr(ErrUnknownSpot, "spot is required")
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
	if err := checkPlacement(preset, *in.Spot); err != nil {
		return nil, err
	}
	if err := checkPhoto(tx, in.Type, in.CarepartnerID, in.PhotoMediaID); err != nil {
		return nil, err
	}
	occupant, err := model.GetScreenWidgetAt(tx, in.CarepartnerID, *in.Spot)
	if err != nil {
		return nil, fmt.Errorf("failed to check spot: %w", err)
	}
	if occupant != nil {
		return nil, ErrSpotTaken
	}

	w := &model.ScreenWidget{
		CarepartnerID: in.CarepartnerID,
		Type:          in.Type,
		Spot:          *in.Spot,
		Config:        config,
		PhotoMediaID:  in.PhotoMediaID,
	}
	if err := model.CreateScreenWidget(tx, w); err != nil {
		if isSpotConflict(err) {
			return nil, ErrSpotTaken
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
ListWidgets returns a carepartner's widgets in spot order. Widgets in the current preset
come first, since hidden ones sit in spots past the preset's last.
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
	w.Hidden = !preset.HasSpot(w.Spot)
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
	w.Hidden = !preset.HasSpot(w.Spot)
	return w, nil
}

/*
MoveWidget moves a widget to another spot of the current preset. If that spot has a
widget, the two swap places; the other widget takes the moving widget's old spot, even
if that spot is hidden in the current preset.
*/
func (s *ScreenWidgetService) MoveWidget(id int, in MoveInput) (*model.ScreenWidget, error) {
	if id <= 0 {
		return nil, ErrNotFound
	}
	if in.Spot == nil {
		return nil, placementErr(ErrUnknownSpot, "spot is required")
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
	if err := checkPlacement(preset, *in.Spot); err != nil {
		return nil, err
	}
	if w.Spot == *in.Spot {
		w.Hidden = false
		return w, nil
	}

	occupant, err := model.GetScreenWidgetAt(tx, w.CarepartnerID, *in.Spot)
	if err != nil {
		return nil, fmt.Errorf("failed to check spot: %w", err)
	}
	if occupant != nil {
		if err := model.MoveScreenWidget(tx, occupant.ID, w.Spot); err != nil {
			return nil, fmt.Errorf("failed to move widget: %w", err)
		}
	}
	if err := model.MoveScreenWidget(tx, w.ID, *in.Spot); err != nil {
		return nil, fmt.Errorf("failed to move widget: %w", err)
	}
	if err := commit(tx); err != nil {
		return nil, err
	}

	w.Spot, w.Hidden = *in.Spot, false
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
