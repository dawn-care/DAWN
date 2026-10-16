package screen_widgets

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	model "server/internal/models/screen_widgets"
	svc "server/internal/services/screen_widgets"
)

// Error codes returned in the "code" field of error responses
const (
	CodeInvalidRequest  = "INVALID_REQUEST"
	CodeInvalidType     = "INVALID_TYPE"
	CodeInvalidConfig   = "INVALID_CONFIG"
	CodeInvalidPhoto    = "INVALID_PHOTO"
	CodeInvalidPosition = "INVALID_POSITION"
	CodeOverlap         = "WIDGETS_OVERLAP"
	CodeNotFound        = "NOT_FOUND"
	CodeInternal        = "INTERNAL_ERROR"
)

// ---- Interface ----
type ScreenWidgetServiceInterface interface {
	GetLayout(carepartnerID int) (*model.Layout, error)
	SaveLayout(carepartnerID int, in model.LayoutInput) (*model.Layout, error)
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

var errorCodes = []struct {
	err    error
	status int
	code   string
}{
	{svc.ErrNotFound, http.StatusNotFound, CodeNotFound},
	{svc.ErrCarepartnerMissing, http.StatusNotFound, CodeNotFound},
	{svc.ErrInvalidCarepartner, http.StatusBadRequest, CodeInvalidRequest},
	{svc.ErrInvalidRequest, http.StatusBadRequest, CodeInvalidRequest},
	{svc.ErrInvalidType, http.StatusBadRequest, CodeInvalidType},
	{svc.ErrInvalidPosition, http.StatusBadRequest, CodeInvalidPosition},
	{svc.ErrInvalidConfig, http.StatusBadRequest, CodeInvalidConfig},
	{svc.ErrInvalidPhoto, http.StatusBadRequest, CodeInvalidPhoto},
	{svc.ErrOverlap, http.StatusBadRequest, CodeOverlap},
}

func writeServiceError(w http.ResponseWriter, err error) {
	for _, e := range errorCodes {
		if errors.Is(err, e.err) {
			writeError(w, e.status, e.code, err.Error())
			return
		}
	}
	writeError(w, http.StatusInternalServerError, CodeInternal, err.Error())
}

func carepartnerParam(w http.ResponseWriter, r *http.Request) (int, bool) {
	id, err := strconv.Atoi(r.URL.Query().Get("carepartner_id"))
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, CodeInvalidRequest, svc.ErrInvalidCarepartner.Error())
		return 0, false
	}
	return id, true
}

// GetLayoutHandler godoc
// @Summary Get a carepartner's screen layout
// @Description Every widget on the carepartner's screen, in reading order (top row first, then left to right). top_left and bottom_right are the cells at the widget's corners, both inclusive, on the 4x3 grid (columns 0-3, rows 0-2).
// @Tags screen-layout
// @Produce json
// @Param carepartner_id query int true "Carepartner ID"
// @Success 200 {object} screen_widgets.Layout
// @Failure 400 {object} screen_widgets.ScreenWidgetError
// @Failure 404 {object} screen_widgets.ScreenWidgetError
// @Failure 500 {object} screen_widgets.ScreenWidgetError
// @Router /screen-layout [get]
func (c *ScreenWidgetController) GetLayoutHandler(w http.ResponseWriter, r *http.Request) {
	carepartnerID, ok := carepartnerParam(w, r)
	if !ok {
		return
	}

	layout, err := c.Service.GetLayout(carepartnerID)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, layout)
}

// SaveLayoutHandler godoc
// @Summary Replace a carepartner's screen layout
// @Description Send every widget on the screen. A widget with an id updates that widget (its type can't change); one without an id is created; the carepartner's other widgets are deleted. widgets is required: send [] to clear the screen. Every widget must lie inside the 4x3 grid and no two may share a cell. Nothing is saved if any widget is invalid; errors name the widget by its index, e.g. "widgets[2]". The saved widgets come back in the order they were sent.
// @Tags screen-layout
// @Accept json
// @Produce json
// @Param carepartner_id query int true "Carepartner ID"
// @Param body body screen_widgets.LayoutInput true "Every widget on the screen"
// @Success 200 {object} screen_widgets.Layout
// @Failure 400 {object} screen_widgets.ScreenWidgetError
// @Failure 404 {object} screen_widgets.ScreenWidgetError
// @Failure 500 {object} screen_widgets.ScreenWidgetError
// @Router /screen-layout [put]
func (c *ScreenWidgetController) SaveLayoutHandler(w http.ResponseWriter, r *http.Request) {
	carepartnerID, ok := carepartnerParam(w, r)
	if !ok {
		return
	}
	var in model.LayoutInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeError(w, http.StatusBadRequest, CodeInvalidRequest, "invalid request body")
		return
	}

	layout, err := c.Service.SaveLayout(carepartnerID, in)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, layout)
}
