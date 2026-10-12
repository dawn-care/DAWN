package scheduled_items

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	model "server/internal/models/scheduled_items"
	svc "server/internal/services/scheduled_items"
)

// Error codes returned in the "code" field of error responses
const (
	CodeInvalidRequest    = "INVALID_REQUEST"
	CodeInvalidRepeat     = "INVALID_REPEAT"
	CodeInvalidDate       = "INVALID_DATE"
	CodeInvalidTime       = "INVALID_TIME"
	CodeNotOn15MinStep    = "NOT_ON_15_MIN_STEP"
	CodeOutsideAwakeHours = "OUTSIDE_AWAKE_HOURS"
	CodeInvalidMedia      = "INVALID_MEDIA"
	CodeInvalidShape      = "INVALID_SHAPE"
	CodeInvalidScope      = "INVALID_SCOPE"
	CodeNotAnOccurrence   = "NOT_AN_OCCURRENCE"
	CodeInvalidRange      = "INVALID_RANGE"
	CodeNotFound          = "NOT_FOUND"
	CodeInternal          = "INTERNAL_ERROR"
)

// ---- Interface ----
type ScheduledItemServiceInterface interface {
	CreateItem(in svc.ItemInput) (*model.ScheduledItem, error)
	GetItem(id int) (*model.ScheduledItem, error)
	ListItems(carepartnerID int) ([]model.ScheduledItem, error)
	Calendar(carepartnerID int, from, to string) ([]model.Occurrence, error)
	UpdateItem(id int, in svc.ItemInput) (*model.ScheduledItem, error)
	DeleteItem(id int, scope, date string) error
}

// ---- Controller ----
type ScheduledItemController struct {
	Service ScheduledItemServiceInterface
}

func NewScheduledItemController(s ScheduledItemServiceInterface) *ScheduledItemController {
	return &ScheduledItemController{Service: s}
}

// ScheduledItemError is the body of every error response from these endpoints
type ScheduledItemError struct {
	Code  string `json:"code"`
	Error string `json:"error"`
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, code, msg string) {
	writeJSON(w, status, ScheduledItemError{Code: code, Error: msg})
}

var errorCodes = []struct {
	err    error
	status int
	code   string
}{
	{svc.ErrNotFound, http.StatusNotFound, CodeNotFound},
	{svc.ErrCarepartnerMissing, http.StatusNotFound, CodeNotFound},
	{svc.ErrInvalidCarepartner, http.StatusBadRequest, CodeInvalidRequest},
	{svc.ErrInvalidRepeat, http.StatusBadRequest, CodeInvalidRepeat},
	{svc.ErrInvalidDate, http.StatusBadRequest, CodeInvalidDate},
	{svc.ErrInvalidTime, http.StatusBadRequest, CodeInvalidTime},
	{svc.ErrNotOn15MinStep, http.StatusBadRequest, CodeNotOn15MinStep},
	{svc.ErrOutsideAwakeHours, http.StatusBadRequest, CodeOutsideAwakeHours},
	{svc.ErrInvalidMedia, http.StatusBadRequest, CodeInvalidMedia},
	{svc.ErrInvalidShape, http.StatusBadRequest, CodeInvalidShape},
	{svc.ErrInvalidScope, http.StatusBadRequest, CodeInvalidScope},
	{svc.ErrNotAnOccurrence, http.StatusBadRequest, CodeNotAnOccurrence},
	{svc.ErrInvalidRange, http.StatusBadRequest, CodeInvalidRange},
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

func itemIDParam(r *http.Request) (int, bool) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	return id, err == nil && id > 0
}

func carepartnerParam(w http.ResponseWriter, r *http.Request) (int, bool) {
	id, err := strconv.Atoi(r.URL.Query().Get("carepartner_id"))
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, CodeInvalidRequest, svc.ErrInvalidCarepartner.Error())
		return 0, false
	}
	return id, true
}

// CreateItemHandler godoc
// @Summary Schedule a video, audio file or photo
// @Description Shapes: video; audio, optionally with background_photo_id; photo with display_seconds (1-600). start_time must be on a 15 minute step, and the item must start at or after wake time and finish by bed time. Overlap is checked by the frontend.
// @Tags scheduled-items
// @Accept json
// @Produce json
// @Param body body scheduled_items.ItemInput true "New scheduled item"
// @Success 201 {object} scheduled_items.ScheduledItem
// @Failure 400 {object} scheduled_items.ScheduledItemError
// @Failure 404 {object} scheduled_items.ScheduledItemError
// @Failure 500 {object} scheduled_items.ScheduledItemError
// @Router /scheduled-items [post]
func (c *ScheduledItemController) CreateItemHandler(w http.ResponseWriter, r *http.Request) {
	var in svc.ItemInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeError(w, http.StatusBadRequest, CodeInvalidRequest, "invalid request body")
		return
	}

	item, err := c.Service.CreateItem(in)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, item)
}

// ListItemsHandler godoc
// @Summary List a carepartner's scheduled items as rules
// @Description Each item once, with its repeat rule and skipped days. Use GET /calendar for dated occurrences.
// @Tags scheduled-items
// @Produce json
// @Param carepartner_id query int true "Carepartner ID"
// @Success 200 {array} scheduled_items.ScheduledItem
// @Failure 400 {object} scheduled_items.ScheduledItemError
// @Failure 404 {object} scheduled_items.ScheduledItemError
// @Failure 500 {object} scheduled_items.ScheduledItemError
// @Router /scheduled-items [get]
func (c *ScheduledItemController) ListItemsHandler(w http.ResponseWriter, r *http.Request) {
	carepartnerID, ok := carepartnerParam(w, r)
	if !ok {
		return
	}

	items, err := c.Service.ListItems(carepartnerID)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, items)
}

// GetItemHandler godoc
// @Summary Get one scheduled item
// @Tags scheduled-items
// @Produce json
// @Param id path int true "Scheduled item ID"
// @Success 200 {object} scheduled_items.ScheduledItem
// @Failure 404 {object} scheduled_items.ScheduledItemError
// @Failure 500 {object} scheduled_items.ScheduledItemError
// @Router /scheduled-items/{id} [get]
func (c *ScheduledItemController) GetItemHandler(w http.ResponseWriter, r *http.Request) {
	id, ok := itemIDParam(r)
	if !ok {
		writeError(w, http.StatusNotFound, CodeNotFound, svc.ErrNotFound.Error())
		return
	}

	item, err := c.Service.GetItem(id)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, item)
}

// UpdateItemHandler godoc
// @Summary Replace a scheduled item
// @Description Same rules as creating one. Days already deleted with scope=occurrence stay deleted.
// @Tags scheduled-items
// @Accept json
// @Produce json
// @Param id path int true "Scheduled item ID"
// @Param body body scheduled_items.ItemInput true "Scheduled item"
// @Success 200 {object} scheduled_items.ScheduledItem
// @Failure 400 {object} scheduled_items.ScheduledItemError
// @Failure 404 {object} scheduled_items.ScheduledItemError
// @Failure 500 {object} scheduled_items.ScheduledItemError
// @Router /scheduled-items/{id} [put]
func (c *ScheduledItemController) UpdateItemHandler(w http.ResponseWriter, r *http.Request) {
	id, ok := itemIDParam(r)
	if !ok {
		writeError(w, http.StatusNotFound, CodeNotFound, svc.ErrNotFound.Error())
		return
	}
	var in svc.ItemInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeError(w, http.StatusBadRequest, CodeInvalidRequest, "invalid request body")
		return
	}

	item, err := c.Service.UpdateItem(id, in)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, item)
}

// DeleteItemHandler godoc
// @Summary Delete a scheduled item, every time or just one day
// @Tags scheduled-items
// @Param id path int true "Scheduled item ID"
// @Param scope query string true "series (every time) or occurrence (just this day)"
// @Param date query string false "YYYY-MM-DD; required for scope=occurrence"
// @Success 204 {string} string "No Content"
// @Failure 400 {object} scheduled_items.ScheduledItemError
// @Failure 404 {object} scheduled_items.ScheduledItemError
// @Failure 500 {object} scheduled_items.ScheduledItemError
// @Router /scheduled-items/{id} [delete]
func (c *ScheduledItemController) DeleteItemHandler(w http.ResponseWriter, r *http.Request) {
	id, ok := itemIDParam(r)
	if !ok {
		writeError(w, http.StatusNotFound, CodeNotFound, svc.ErrNotFound.Error())
		return
	}

	q := r.URL.Query()
	if err := c.Service.DeleteItem(id, q.Get("scope"), q.Get("date")); err != nil {
		writeServiceError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// CalendarHandler godoc
// @Summary Dated occurrences of a carepartner's scheduled items
// @Description Repeating items are expanded on the server, skipped days are left out, and outside_awake_hours reflects the current awake hours. from and to are inclusive and at most 62 days apart.
// @Tags scheduled-items
// @Produce json
// @Param carepartner_id query int true "Carepartner ID"
// @Param from query string true "YYYY-MM-DD"
// @Param to query string true "YYYY-MM-DD"
// @Success 200 {array} scheduled_items.Occurrence
// @Failure 400 {object} scheduled_items.ScheduledItemError
// @Failure 404 {object} scheduled_items.ScheduledItemError
// @Failure 500 {object} scheduled_items.ScheduledItemError
// @Router /calendar [get]
func (c *ScheduledItemController) CalendarHandler(w http.ResponseWriter, r *http.Request) {
	carepartnerID, ok := carepartnerParam(w, r)
	if !ok {
		return
	}

	q := r.URL.Query()
	occurrences, err := c.Service.Calendar(carepartnerID, q.Get("from"), q.Get("to"))
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, occurrences)
}
