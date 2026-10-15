package playback

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"time"

	playback "server/internal/models/playback"
	"server/internal/realtime"
	playbackSvc "server/internal/services/playback"

	"github.com/go-chi/chi/v5"
)

type PlaybackServiceInterface interface {
	CreatePlayback(context.Context, int, int, *int, string, time.Time) (*playback.Playback, error)
	UpdateStatus(context.Context, int, int, string) (*playback.Playback, error)
	RequestStaffPlay(context.Context, int, int) (*playback.Playback, error)
	RequestStop(context.Context, int, int) (*playback.Playback, playbackSvc.StopOutcome, error)
}

// Notifier pushes a message to a carepartner's tablet, reporting whether it was
// delivered to an open connection. Implemented by realtime.Hub.
type Notifier interface {
	Send(carepartnerID int, msg realtime.Message) bool
}

type PlaybackController struct {
	Service  PlaybackServiceInterface
	Notifier Notifier
}

func NewPlaybackController(service PlaybackServiceInterface, notifier Notifier) *PlaybackController {
	return &PlaybackController{Service: service, Notifier: notifier}
}

// CreatePlaybackRequest names the media to play. scheduled_item_id is required when
// source is "scheduled" and must be left out otherwise.
type CreatePlaybackRequest struct {
	CarepartnerID   int       `json:"carepartner_id"`
	MediaID         int       `json:"media_id"`
	ScheduledItemID *int      `json:"scheduled_item_id"`
	Source          string    `json:"source"`
	ExpiresAt       time.Time `json:"expires_at"`
}

type UpdateStatusRequest struct {
	CarepartnerID int    `json:"carepartner_id"`
	Status        string `json:"status"`
}

// CreatePlaybackHandler godoc
// @Summary Record a playback request
// @Tags playback
// @Accept json
// @Produce json
// @Param playback body CreatePlaybackRequest true "Playback request; expires_at is RFC3339"
// @Success 201 {object} playback.Playback
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /playbacks [post]
func (c *PlaybackController) CreatePlaybackHandler(w http.ResponseWriter, r *http.Request) {
	var req CreatePlaybackRequest
	if !decodeRequest(w, r, &req) {
		return
	}
	p, err := c.Service.CreatePlayback(r.Context(), req.CarepartnerID, req.MediaID, req.ScheduledItemID, req.Source, req.ExpiresAt)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, p)
}

// UpdateStatusHandler godoc
// @Summary Report a playback status change
// @Tags playback
// @Accept json
// @Produce json
// @Param id path int true "Playback ID"
// @Param status body UpdateStatusRequest true "Status report"
// @Success 200 {object} playback.Playback
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /playbacks/{id}/status [patch]
func (c *PlaybackController) UpdateStatusHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || id <= 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid playback ID"})
		return
	}
	var req UpdateStatusRequest
	if !decodeRequest(w, r, &req) {
		return
	}
	p, err := c.Service.UpdateStatus(r.Context(), id, req.CarepartnerID, req.Status)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, p)
}

// Accept only the declared fields and exactly one JSON value. In particular,
// free-text notes, error messages and client-supplied lifecycle timestamps are rejected.
func decodeRequest(w http.ResponseWriter, r *http.Request, dst any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 4096)
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

func writeServiceError(w http.ResponseWriter, err error) {
	status, message := http.StatusInternalServerError, "could not save playback"
	switch {
	case errors.Is(err, playbackSvc.ErrInvalidInput):
		status, message = http.StatusBadRequest, "invalid playback input"
	case errors.Is(err, playbackSvc.ErrNotFound):
		status, message = http.StatusNotFound, "playback not found"
	case errors.Is(err, playbackSvc.ErrInvalidTransition):
		status, message = http.StatusConflict, "invalid playback status transition"
	}
	// Never expose SQL errors or echo arbitrary request text.
	writeJSON(w, status, map[string]string{"error": message})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
