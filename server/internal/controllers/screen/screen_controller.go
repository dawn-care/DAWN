package screen

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	svc "server/internal/services/screen"
)

// ---- Interface ----
type ScreenServiceInterface interface {
	GetScreen(carepartnerID int) (*svc.Screen, error)
}

// ---- Controller ----
type ScreenController struct {
	Service ScreenServiceInterface
}

func NewScreenController(s ScreenServiceInterface) *ScreenController {
	return &ScreenController{Service: s}
}

// ScreenError is the body of every error response from this endpoint
type ScreenError struct {
	Code  string `json:"code"`
	Error string `json:"error"`
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

// GetScreenHandler godoc
// @Summary Everything the carepartner's tablet shows
// @Description Layout preset; each region's slots in order (null for an empty slot) with widget settings, photo and data; awake hours; and today's and tomorrow's schedule. A widget whose outside lookup fails has data null and error "UPSTREAM_UNAVAILABLE". Weather, sports and news results are cached for 10 minutes. Times are in the carepartner's timezone.
// @Tags screen
// @Produce json
// @Param carepartner_id query int true "Carepartner ID"
// @Success 200 {object} screen.Screen
// @Failure 400 {object} screen.ScreenError
// @Failure 404 {object} screen.ScreenError
// @Failure 500 {object} screen.ScreenError
// @Router /screen [get]
func (c *ScreenController) GetScreenHandler(w http.ResponseWriter, r *http.Request) {
	carepartnerID, err := strconv.Atoi(r.URL.Query().Get("carepartner_id"))
	if err != nil || carepartnerID <= 0 {
		writeJSON(w, http.StatusBadRequest, ScreenError{Code: "INVALID_REQUEST", Error: svc.ErrInvalidCarepartner.Error()})
		return
	}

	screen, err := c.Service.GetScreen(carepartnerID)
	switch {
	case errors.Is(err, svc.ErrCarepartnerMissing):
		writeJSON(w, http.StatusNotFound, ScreenError{Code: "NOT_FOUND", Error: err.Error()})
	case err != nil:
		writeJSON(w, http.StatusInternalServerError, ScreenError{Code: "INTERNAL_ERROR", Error: err.Error()})
	default:
		writeJSON(w, http.StatusOK, screen)
	}
}
