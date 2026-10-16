package screen_widgets

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	model "server/internal/models/screen_widgets"
	svc "server/internal/services/screen_widgets"
)

// Error codes returned in the "code" field of error responses
const (
	CodeInvalidRequest = "INVALID_REQUEST"
	CodeInvalidType    = "INVALID_TYPE"
	CodeInvalidConfig  = "INVALID_CONFIG"
	CodeInvalidPhoto   = "INVALID_PHOTO"
	CodeUnknownSpot    = "UNKNOWN_SPOT"
	CodeSpotTaken      = "SPOT_TAKEN"
	CodeNotFound       = "NOT_FOUND"
	CodeInternal       = "INTERNAL_ERROR"
)

// ---- Interface ----
type ScreenWidgetServiceInterface interface {
	CreateWidget(in svc.CreateInput) (*model.ScreenWidget, error)
	ListWidgets(carepartnerID int) ([]model.ScreenWidget, error)
	GetWidget(id int) (*model.ScreenWidget, error)
	UpdateWidget(id int, in svc.UpdateInput) (*model.ScreenWidget, error)
	MoveWidget(id int, in svc.MoveInput) (*model.ScreenWidget, error)
	DeleteWidget(id int) error
}

// ---- Controller ----
type ScreenWidgetController struct {
	Service ScreenWidgetServiceInterface
}

func NewScreenWidgetController(s ScreenWidgetServiceInterface) *ScreenWidgetController {
	return &ScreenWidgetController{Service: s}
}

// ScreenWidgetError is the body of every error response from these endpoints
type ScreenWidgetError struct {
	Code  string `json:"code"`
	Error string `json:"error"`
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, code, msg string) {
	writeJSON(w, status, ScreenWidgetError{Code: code, Error: msg})
}

func writeServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, svc.ErrNotFound), errors.Is(err, svc.ErrCarepartnerMissing):
		writeError(w, http.StatusNotFound, CodeNotFound, err.Error())
	case errors.Is(err, svc.ErrInvalidCarepartner):
		writeError(w, http.StatusBadRequest, CodeInvalidRequest, err.Error())
	case errors.Is(err, svc.ErrInvalidType):
		writeError(w, http.StatusBadRequest, CodeInvalidType, err.Error())
	case errors.Is(err, svc.ErrInvalidConfig):
		writeError(w, http.StatusBadRequest, CodeInvalidConfig, err.Error())
	case errors.Is(err, svc.ErrInvalidPhoto):
		writeError(w, http.StatusBadRequest, CodeInvalidPhoto, err.Error())
	case errors.Is(err, svc.ErrUnknownSpot):
		writeError(w, http.StatusBadRequest, CodeUnknownSpot, err.Error())
	case errors.Is(err, svc.ErrSpotTaken):
		writeError(w, http.StatusConflict, CodeSpotTaken, err.Error())
	default:
		writeError(w, http.StatusInternalServerError, CodeInternal, err.Error())
	}
}

func widgetIDParam(r *http.Request) (int, bool) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	return id, err == nil && id > 0
}

func decodeBody(w http.ResponseWriter, r *http.Request, v any) bool {
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		writeError(w, http.StatusBadRequest, CodeInvalidRequest, "invalid request body")
		return false
	}
	return true
}

// ListWidgetsHandler godoc
// @Summary List a carepartner's screen widgets
// @Description Widgets in spot order. Widgets whose spot the current layout doesn't have come last with hidden: true; the tablet doesn't show them.
// @Tags screen-widgets
// @Produce json
// @Param carepartner_id query int true "Carepartner ID"
// @Success 200 {array} screen_widgets.ScreenWidget
// @Failure 400 {object} screen_widgets.ScreenWidgetError
// @Failure 404 {object} screen_widgets.ScreenWidgetError
// @Failure 500 {object} screen_widgets.ScreenWidgetError
// @Router /screen-widgets [get]
func (c *ScreenWidgetController) ListWidgetsHandler(w http.ResponseWriter, r *http.Request) {
	carepartnerID, err := strconv.Atoi(r.URL.Query().Get("carepartner_id"))
	if err != nil || carepartnerID <= 0 {
		writeError(w, http.StatusBadRequest, CodeInvalidRequest, svc.ErrInvalidCarepartner.Error())
		return
	}

	widgets, err := c.Service.ListWidgets(carepartnerID)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, widgets)
}

// CreateWidgetHandler godoc
// @Summary Place a new widget in a spot
// @Description The spot (0-based) must exist in the carepartner's current layout (see GET /screen-settings/presets) and be empty. Any widget type can go in any spot. config is validated per type; photo widgets need photo_media_id.
// @Tags screen-widgets
// @Accept json
// @Produce json
// @Param body body screen_widgets.CreateInput true "New widget"
// @Success 201 {object} screen_widgets.ScreenWidget
// @Failure 400 {object} screen_widgets.ScreenWidgetError
// @Failure 404 {object} screen_widgets.ScreenWidgetError
// @Failure 409 {object} screen_widgets.ScreenWidgetError
// @Failure 500 {object} screen_widgets.ScreenWidgetError
// @Router /screen-widgets [post]
func (c *ScreenWidgetController) CreateWidgetHandler(w http.ResponseWriter, r *http.Request) {
	var in svc.CreateInput
	if !decodeBody(w, r, &in) {
		return
	}

	widget, err := c.Service.CreateWidget(in)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, widget)
}

// GetWidgetHandler godoc
// @Summary Get one screen widget
// @Tags screen-widgets
// @Produce json
// @Param id path int true "Widget ID"
// @Success 200 {object} screen_widgets.ScreenWidget
// @Failure 404 {object} screen_widgets.ScreenWidgetError
// @Failure 500 {object} screen_widgets.ScreenWidgetError
// @Router /screen-widgets/{id} [get]
func (c *ScreenWidgetController) GetWidgetHandler(w http.ResponseWriter, r *http.Request) {
	id, ok := widgetIDParam(r)
	if !ok {
		writeError(w, http.StatusNotFound, CodeNotFound, svc.ErrNotFound.Error())
		return
	}

	widget, err := c.Service.GetWidget(id)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, widget)
}

// UpdateWidgetHandler godoc
// @Summary Replace a widget's settings and photo
// @Description The widget's type and spot don't change; use the move endpoint to change its spot.
// @Tags screen-widgets
// @Accept json
// @Produce json
// @Param id path int true "Widget ID"
// @Param body body screen_widgets.UpdateInput true "New settings"
// @Success 200 {object} screen_widgets.ScreenWidget
// @Failure 400 {object} screen_widgets.ScreenWidgetError
// @Failure 404 {object} screen_widgets.ScreenWidgetError
// @Failure 500 {object} screen_widgets.ScreenWidgetError
// @Router /screen-widgets/{id} [put]
func (c *ScreenWidgetController) UpdateWidgetHandler(w http.ResponseWriter, r *http.Request) {
	id, ok := widgetIDParam(r)
	if !ok {
		writeError(w, http.StatusNotFound, CodeNotFound, svc.ErrNotFound.Error())
		return
	}
	var in svc.UpdateInput
	if !decodeBody(w, r, &in) {
		return
	}

	widget, err := c.Service.UpdateWidget(id, in)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, widget)
}

// MoveWidgetHandler godoc
// @Summary Move a widget to another spot
// @Description The target spot must exist in the current layout. If it has a widget, the two swap places; the other widget takes this one's old spot, even if that spot is hidden in the current layout.
// @Tags screen-widgets
// @Accept json
// @Produce json
// @Param id path int true "Widget ID"
// @Param body body screen_widgets.MoveInput true "Target spot"
// @Success 200 {object} screen_widgets.ScreenWidget
// @Failure 400 {object} screen_widgets.ScreenWidgetError
// @Failure 404 {object} screen_widgets.ScreenWidgetError
// @Failure 409 {object} screen_widgets.ScreenWidgetError
// @Failure 500 {object} screen_widgets.ScreenWidgetError
// @Router /screen-widgets/{id}/move [post]
func (c *ScreenWidgetController) MoveWidgetHandler(w http.ResponseWriter, r *http.Request) {
	id, ok := widgetIDParam(r)
	if !ok {
		writeError(w, http.StatusNotFound, CodeNotFound, svc.ErrNotFound.Error())
		return
	}
	var in svc.MoveInput
	if !decodeBody(w, r, &in) {
		return
	}

	widget, err := c.Service.MoveWidget(id, in)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, widget)
}

// DeleteWidgetHandler godoc
// @Summary Remove a widget from the screen
// @Tags screen-widgets
// @Param id path int true "Widget ID"
// @Success 204 {string} string "No Content"
// @Failure 404 {object} screen_widgets.ScreenWidgetError
// @Failure 500 {object} screen_widgets.ScreenWidgetError
// @Router /screen-widgets/{id} [delete]
func (c *ScreenWidgetController) DeleteWidgetHandler(w http.ResponseWriter, r *http.Request) {
	id, ok := widgetIDParam(r)
	if !ok {
		writeError(w, http.StatusNotFound, CodeNotFound, svc.ErrNotFound.Error())
		return
	}

	if err := c.Service.DeleteWidget(id); err != nil {
		writeServiceError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
