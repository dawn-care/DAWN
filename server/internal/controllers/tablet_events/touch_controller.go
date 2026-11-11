package tablet_events

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"

	// Imported under its own package name: swag (v1.16) can't find a model type through
	// any other alias, and RecordTouchesRequest would be empty in the Swagger docs
	tablet_events "server/internal/models/tablet_events"
	touchSvc "server/internal/services/tablet_events"
)

type TouchServiceInterface interface {
	RecordTouches(context.Context, int, []tablet_events.TouchEvent) error
}

type TouchController struct {
	Service TouchServiceInterface
}

func NewTouchController(service TouchServiceInterface) *TouchController {
	return &TouchController{Service: service}
}

// RecordTouchesRequest is one batch from the tablet's touch telemetry
// (carepartner/src/telemetry/touchLog.ts), at most 50 events
type RecordTouchesRequest struct {
	CarepartnerID int                        `json:"carepartner_id"`
	Events        []tablet_events.TouchEvent `json:"events"`
}

// maxBodyBytes fits a full batch (50 events are about 15 KB) with room to spare
const maxBodyBytes = 64 << 10

// RecordTouchesHandler godoc
// @Summary Store a batch of the resident's touches
// @Description The tablet sends the touches it recorded about once a minute, at most 50 events per request; quick repeats on the same thing are one event with a count. The whole batch is stored, or none of it. Unknown fields are rejected. 400 means the batch can never be stored and the tablet drops it; 5xx means it should retry, and a resent batch is not stored twice.
// @Tags tablet-events
// @Accept json
// @Param batch body RecordTouchesRequest true "Touch events; first_at and last_at are RFC 3339"
// @Success 204
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /tablet-events/touches [post]
func (c *TouchController) RecordTouchesHandler(w http.ResponseWriter, r *http.Request) {
	var req RecordTouchesRequest
	if !decodeRequest(w, r, &req) {
		return
	}
	err := c.Service.RecordTouches(r.Context(), req.CarepartnerID, req.Events)
	switch {
	case err == nil:
		w.WriteHeader(http.StatusNoContent)
	case errors.Is(err, touchSvc.ErrInvalidInput):
		// The reason is the service's own wording, never request text
		log.Printf("tablet touches: refused a batch from carepartner %d: %v", req.CarepartnerID, err)
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
	default:
		// Never expose SQL errors
		log.Printf("tablet touches: could not store a batch from carepartner %d: %v", req.CarepartnerID, err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "could not store touches"})
	}
}

// decodeRequest accepts only the declared fields, at any depth, and exactly one
// JSON value, so a batch with positions or text in it is refused
func decodeRequest(w http.ResponseWriter, r *http.Request, dst any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
	defer r.Body.Close()
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dst); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return false
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "expected a single JSON object"})
		return false
	}
	return true
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
