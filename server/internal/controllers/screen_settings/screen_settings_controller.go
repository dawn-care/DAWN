package screen_settings

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	model "server/internal/models/screen_settings"
	svc "server/internal/services/screen_settings"
)

// Error codes returned in the "code" field of error responses
const (
	CodeInvalidRequest  = "INVALID_REQUEST"
	CodeInvalidWakeTime = "INVALID_WAKE_TIME"
	CodeInvalidBedTime  = "INVALID_BED_TIME"
	CodeNotFound        = "NOT_FOUND"
	CodeInternal        = "INTERNAL_ERROR"
)

// ---- Interface ----
type ScreenSettingsServiceInterface interface {
	GetScreenSettings(carepartnerID int) (*model.ScreenSettings, error)
	UpdateScreenSettings(carepartnerID int, in svc.UpdateInput) (*model.ScreenSettings, error)
}

// ---- Controller ----
type ScreenSettingsController struct {
	Service ScreenSettingsServiceInterface
}

func NewScreenSettingsController(s ScreenSettingsServiceInterface) *ScreenSettingsController {
	return &ScreenSettingsController{Service: s}
}

// ScreenSettingsError is the body of every error response from these endpoints
type ScreenSettingsError struct {
	Code  string `json:"code"`
	Error string `json:"error"`
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, code, msg string) {
	writeJSON(w, status, ScreenSettingsError{Code: code, Error: msg})
}

func writeServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, svc.ErrNotFound):
		writeError(w, http.StatusNotFound, CodeNotFound, err.Error())
	case errors.Is(err, svc.ErrInvalidCarepartner):
		writeError(w, http.StatusBadRequest, CodeInvalidRequest, err.Error())
	case errors.Is(err, svc.ErrInvalidWakeTime):
		writeError(w, http.StatusBadRequest, CodeInvalidWakeTime, err.Error())
	case errors.Is(err, svc.ErrInvalidBedTime):
		writeError(w, http.StatusBadRequest, CodeInvalidBedTime, err.Error())
	default:
		writeError(w, http.StatusInternalServerError, CodeInternal, err.Error())
	}
}

func carepartnerIDParam(r *http.Request) (int, bool) {
	id, err := strconv.Atoi(r.URL.Query().Get("carepartner_id"))
	return id, err == nil && id > 0
}

// GetScreenSettingsHandler godoc
// @Summary Get a carepartner's screen settings
// @Description Awake hours ("HH:MM", carepartner's local time). Every carepartner has settings from creation (07:00 to 21:00). Widget placement is in /screen-layout.
// @Tags screen-settings
// @Produce json
// @Param carepartner_id query int true "Carepartner ID"
// @Success 200 {object} screen_settings.ScreenSettings
// @Failure 400 {object} screen_settings.ScreenSettingsError
// @Failure 404 {object} screen_settings.ScreenSettingsError
// @Failure 500 {object} screen_settings.ScreenSettingsError
// @Router /screen-settings [get]
func (c *ScreenSettingsController) GetScreenSettingsHandler(w http.ResponseWriter, r *http.Request) {
	carepartnerID, ok := carepartnerIDParam(r)
	if !ok {
		writeError(w, http.StatusBadRequest, CodeInvalidRequest, svc.ErrInvalidCarepartner.Error())
		return
	}

	settings, err := c.Service.GetScreenSettings(carepartnerID)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, settings)
}

// UpdateScreenSettingsHandler godoc
// @Summary Replace a carepartner's screen settings
// @Description Both fields are required, as "HH:MM". Bed time must be later than wake time on the same day.
// @Tags screen-settings
// @Accept json
// @Produce json
// @Param carepartner_id query int true "Carepartner ID"
// @Param body body screen_settings.UpdateInput true "New settings"
// @Success 200 {object} screen_settings.ScreenSettings
// @Failure 400 {object} screen_settings.ScreenSettingsError
// @Failure 404 {object} screen_settings.ScreenSettingsError
// @Failure 500 {object} screen_settings.ScreenSettingsError
// @Router /screen-settings [put]
func (c *ScreenSettingsController) UpdateScreenSettingsHandler(w http.ResponseWriter, r *http.Request) {
	carepartnerID, ok := carepartnerIDParam(r)
	if !ok {
		writeError(w, http.StatusBadRequest, CodeInvalidRequest, svc.ErrInvalidCarepartner.Error())
		return
	}

	var in svc.UpdateInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeError(w, http.StatusBadRequest, CodeInvalidRequest, "invalid request body")
		return
	}

	settings, err := c.Service.UpdateScreenSettings(carepartnerID, in)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, settings)
}
