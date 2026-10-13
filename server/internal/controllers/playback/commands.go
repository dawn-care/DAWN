package playback

import (
	"net/http"
	"strconv"

	playback "server/internal/models/playback"
	"server/internal/realtime"

	"github.com/go-chi/chi/v5"
)

type PlayRequest struct {
	CarepartnerID int `json:"carepartner_id"`
	ModuleID      int `json:"module_id"`
}

type StopRequest struct {
	CarepartnerID int `json:"carepartner_id"`
}

// CommandResponse returns the stored playback and whether the command reached an
// open tablet connection. An undelivered play command is resent if the tablet
// reconnects before it expires.
type CommandResponse struct {
	Playback  *playback.Playback `json:"playback"`
	Delivered bool               `json:"delivered"`
}

// PlayHandler godoc
// @Summary Play a module on a carepartner's tablet (staff)
// @Description Records a staff playback with a server-set 30 second deadline and pushes it to the tablet.
// @Tags playback
// @Accept json
// @Produce json
// @Param play body PlayRequest true "Carepartner and module to play"
// @Success 201 {object} CommandResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /playbacks/play [post]
func (c *PlaybackController) PlayHandler(w http.ResponseWriter, r *http.Request) {
	var req PlayRequest
	if !decodeRequest(w, r, &req) {
		return
	}
	p, err := c.Service.RequestStaffPlay(r.Context(), req.CarepartnerID, req.ModuleID)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	delivered := c.Notifier.Send(p.CarepartnerID, realtime.PlayCommand(p.ID, p.ModuleID, p.ExpiresAt))
	writeJSON(w, http.StatusCreated, CommandResponse{Playback: p, Delivered: delivered})
}

// StopHandler godoc
// @Summary Stop a playback on a carepartner's tablet (staff)
// @Description Marks the playback stopped and tells the tablet to return to its idle screen.
// @Tags playback
// @Accept json
// @Produce json
// @Param id path int true "Playback ID"
// @Param stop body StopRequest true "Carepartner the playback belongs to"
// @Success 200 {object} CommandResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /playbacks/{id}/stop [post]
func (c *PlaybackController) StopHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || id <= 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid playback ID"})
		return
	}
	var req StopRequest
	if !decodeRequest(w, r, &req) {
		return
	}
	p, err := c.Service.UpdateStatus(r.Context(), id, req.CarepartnerID, playback.StatusStopped)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	delivered := c.Notifier.Send(p.CarepartnerID, realtime.Message{Type: realtime.TypeStop, PlaybackID: p.ID})
	writeJSON(w, http.StatusOK, CommandResponse{Playback: p, Delivered: delivered})
}
