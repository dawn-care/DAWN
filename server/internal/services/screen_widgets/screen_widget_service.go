package screen_widgets

import (
	"database/sql"
	"errors"
	"fmt"

	mediaModel "server/internal/models/media"
	settingsModel "server/internal/models/screen_settings"
	model "server/internal/models/screen_widgets"
)

var (
	ErrInvalidCarepartner = errors.New("invalid carepartner_id")
	ErrCarepartnerMissing = errors.New("carepartner not found")
	ErrNotFound           = errors.New("screen widget not found")
	ErrInvalidRequest     = errors.New("invalid request")
	ErrInvalidType        = fmt.Errorf("type must be one of %v", WidgetTypes)
	ErrInvalidConfig      = errors.New("invalid config")
	ErrInvalidPhoto       = errors.New("invalid photo")
)

type ScreenWidgetService struct {
	DB *sql.DB
}

func NewScreenWidgetService(db *sql.DB) *ScreenWidgetService {
	return &ScreenWidgetService{DB: db}
}

func wrap(base error, format string, args ...any) error {
	return fmt.Errorf("%w: %s", base, fmt.Sprintf(format, args...))
}

// checkPhoto checks a widget's photo: required for photo widgets, optional for
// message widgets, not allowed otherwise, and always a photo of the same carepartner
func checkPhoto(tx *sql.Tx, widgetType string, carepartnerID int, photoID *int) error {
	switch widgetType {
	case TypePhoto:
		if photoID == nil {
			return wrap(ErrInvalidPhoto, "photo widgets need a photo_media_id")
		}
	case TypeMessage:
		if photoID == nil {
			return nil
		}
	default:
		if photoID != nil {
			return wrap(ErrInvalidPhoto, "only photo and message widgets have a photo")
		}
		return nil
	}

	m, err := model.GetPhoto(tx, *photoID)
	if err != nil {
		return fmt.Errorf("failed to check photo: %w", err)
	}
	if m == nil || m.CarepartnerID != carepartnerID {
		return wrap(ErrInvalidPhoto, "photo %d not found for this carepartner", *photoID)
	}
	if m.Kind != mediaModel.KindPhoto {
		return wrap(ErrInvalidPhoto, "media %d is a %s, not a photo", *photoID, m.Kind)
	}
	return nil
}

/*
checkWidget checks one widget of a layout and returns it ready to store. It checks, in
order: type, ID (one of the carepartner's widgets, listed once, same type), corners,
config, then photo. stored holds the carepartner's current widgets by ID; seen collects
the IDs already in the layout.
*/
func checkWidget(tx *sql.Tx, carepartnerID int, in model.WidgetInput, stored map[int]model.ScreenWidget, seen map[int]bool) (model.ScreenWidget, error) {
	if !validType(in.Type) {
		return model.ScreenWidget{}, ErrInvalidType
	}
	w := model.ScreenWidget{CarepartnerID: carepartnerID, Type: in.Type, PhotoMediaID: in.PhotoMediaID}
	if in.ID != nil {
		existing, ok := stored[*in.ID]
		if !ok {
			return w, wrap(ErrNotFound, "widget %d isn't one of this carepartner's widgets", *in.ID)
		}
		if seen[*in.ID] {
			return w, wrap(ErrInvalidRequest, "widget %d is listed more than once", *in.ID)
		}
		if existing.Type != in.Type {
			return w, wrap(ErrInvalidType, "widget %d is a %s widget; send a new widget without an id to change the type", *in.ID, existing.Type)
		}
		seen[*in.ID] = true
		w.ID = *in.ID
	}
	if err := checkArea(in.TopLeft, in.BottomRight); err != nil {
		return w, err
	}
	w.TopLeft, w.BottomRight = *in.TopLeft, *in.BottomRight

	config, err := NormalizeConfig(in.Type, in.Config)
	if err != nil {
		return w, err
	}
	w.Config = config
	if err := checkPhoto(tx, in.Type, carepartnerID, in.PhotoMediaID); err != nil {
		return w, err
	}
	return w, nil
}

// ─────────────────────────────────────────────
// READ
// ─────────────────────────────────────────────

// GetLayout returns a carepartner's widgets in reading order
func (s *ScreenWidgetService) GetLayout(carepartnerID int) (*model.Layout, error) {
	if carepartnerID <= 0 {
		return nil, ErrInvalidCarepartner
	}
	settings, err := settingsModel.GetScreenSettings(s.DB, carepartnerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get screen settings: %w", err)
	}
	if settings == nil {
		return nil, ErrCarepartnerMissing
	}
	widgets, err := model.ListScreenWidgets(s.DB, carepartnerID)
	if err != nil {
		return nil, fmt.Errorf("failed to list widgets: %w", err)
	}
	return &model.Layout{Widgets: widgets}, nil
}

// ─────────────────────────────────────────────
// SAVE
// ─────────────────────────────────────────────

/*
SaveLayout replaces a carepartner's whole layout in one transaction: widgets with an ID
are updated, widgets without one are created, and the carepartner's other widgets are
deleted. Nothing is saved unless every widget is valid and no two share a cell. Errors
name a widget by its index in the list, e.g. "widgets[2]". The saved widgets come back
in the order they were sent, so new widgets' IDs can be matched up by position.
*/
func (s *ScreenWidgetService) SaveLayout(carepartnerID int, in model.LayoutInput) (*model.Layout, error) {
	if carepartnerID <= 0 {
		return nil, ErrInvalidCarepartner
	}
	if in.Widgets == nil {
		return nil, wrap(ErrInvalidRequest, "widgets is required; send [] to clear the screen")
	}
	inputs := *in.Widgets

	tx, err := s.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// One save at a time per carepartner
	found, err := settingsModel.LockScreenSettings(tx, carepartnerID)
	if err != nil {
		return nil, fmt.Errorf("failed to lock screen settings: %w", err)
	}
	if !found {
		return nil, ErrCarepartnerMissing
	}
	current, err := model.ListScreenWidgetsForUpdate(tx, carepartnerID)
	if err != nil {
		return nil, fmt.Errorf("failed to list widgets: %w", err)
	}
	stored := make(map[int]model.ScreenWidget, len(current))
	for _, w := range current {
		stored[w.ID] = w
	}

	widgets := make([]model.ScreenWidget, len(inputs))
	areas := make([]model.GridArea, len(inputs))
	seen := map[int]bool{}
	for i, input := range inputs {
		w, err := checkWidget(tx, carepartnerID, input, stored, seen)
		if err != nil {
			return nil, fmt.Errorf("widgets[%d]: %w", i, err)
		}
		widgets[i], areas[i] = w, w.Area()
	}
	if a, b, overlap := findOverlap(areas); overlap {
		return nil, wrap(ErrOverlap, "widgets[%d] and widgets[%d] cover the same cell", a, b)
	}

	keep := []int{}
	for _, w := range widgets {
		if w.ID != 0 {
			keep = append(keep, w.ID)
		}
	}
	if err := model.DeleteScreenWidgetsExcept(tx, carepartnerID, keep); err != nil {
		return nil, fmt.Errorf("failed to delete widgets: %w", err)
	}
	for i := range widgets {
		if widgets[i].ID == 0 {
			err = model.CreateScreenWidget(tx, &widgets[i])
		} else {
			err = model.UpdateScreenWidget(tx, &widgets[i])
		}
		if err != nil {
			return nil, fmt.Errorf("failed to save widgets[%d]: %w", i, err)
		}
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return &model.Layout{Widgets: widgets}, nil
}
