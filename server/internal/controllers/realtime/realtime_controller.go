// Package realtime handles the tablet's WebSocket connection and the HTTP
// endpoints that inspect or nudge it.
package realtime

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

	playback "server/internal/models/playback"
	"server/internal/models/user/carepartner"
	hub "server/internal/realtime"
	playbackSvc "server/internal/services/playback"

	"github.com/coder/websocket"
	"github.com/go-chi/chi/v5"
)

type PlaybackServiceInterface interface {
	UpdateStatus(ctx context.Context, id, carepartnerID int, status string) (*playback.Playback, error)
	LiveCommand(ctx context.Context, carepartnerID int) (*playback.Playback, error)
}

type CarepartnerServiceInterface interface {
	GetCarepartnerByID(id int) (*carepartner.Carepartner, error)
}

// HubInterface is implemented by realtime.Hub.
type HubInterface interface {
	Serve(ctx context.Context, conn *websocket.Conn, carepartnerID int, initial []hub.Message, handle hub.Handler)
	Send(carepartnerID int, msg hub.Message) bool
	Connected(carepartnerID int) (bool, time.Time)
}

type RealtimeController struct {
	Playback     PlaybackServiceInterface
	Carepartners CarepartnerServiceInterface
	Hub          HubInterface
	// OriginPatterns lists the browser origins (host[:port]) allowed to open a
	// WebSocket — needed for the tablet's web build in development. The native
	// Android app sends no Origin header and is always allowed.
	OriginPatterns []string
}

func NewRealtimeController(playbacks PlaybackServiceInterface, carepartners CarepartnerServiceInterface, h HubInterface, originPatterns []string) *RealtimeController {
	return &RealtimeController{Playback: playbacks, Carepartners: carepartners, Hub: h, OriginPatterns: originPatterns}
}

// ConnectHandler godoc
// @Summary Open the tablet's WebSocket connection
// @Description Upgrades to a WebSocket for one carepartner's tablet. Any live staff command is sent immediately. See internal/realtime/messages.go for the message format.
// @Tags realtime
// @Param carepartner_id query int true "Carepartner ID"
// @Success 101
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /realtime/tablet [get]
func (c *RealtimeController) ConnectHandler(w http.ResponseWriter, r *http.Request) {
	carepartnerID, ok := authenticateTablet(w, r)
	if !ok {
		return
	}
	cp, err := c.Carepartners.GetCarepartnerByID(carepartnerID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "could not look up carepartner"})
		return
	}
	if cp == nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "carepartner not found"})
		return
	}

	// Resend a command the tablet missed while disconnected. A lookup failure is
	// not fatal: the tablet can still connect and receive new commands.
	var initial []hub.Message
	if live, err := c.Playback.LiveCommand(r.Context(), carepartnerID); err != nil {
		log.Printf("realtime: live command lookup for carepartner %d: %v", carepartnerID, err)
	} else if live != nil && live.MediaID != nil {
		initial = append(initial, hub.PlayCommand(live.ID, *live.MediaID, live.ExpiresAt))
	}

	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{OriginPatterns: c.OriginPatterns})
	if err != nil {
		return // Accept has already written the error response.
	}
	c.Hub.Serve(r.Context(), conn, carepartnerID, initial, c.handleMessage)
}

// authenticateTablet identifies the tablet making the request.
//
// TODO(#318): today the carepartner ID — a guessable integer — is the tablet's
// only credential. Replace this with a check of the device token issued at
// pairing; nothing else in this controller needs to change.
func authenticateTablet(w http.ResponseWriter, r *http.Request) (int, bool) {
	id, err := strconv.Atoi(r.URL.Query().Get("carepartner_id"))
	if err != nil || id <= 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid carepartner ID"})
		return 0, false
	}
	return id, true
}

// handleMessage applies a message from a tablet. Status reports go through the
// same service as the REST endpoint, so the transition rules are identical.
func (c *RealtimeController) handleMessage(ctx context.Context, carepartnerID int, msg hub.Message) []hub.Message {
	if msg.Type != hub.TypeStatus {
		return []hub.Message{{Type: hub.TypeError, Error: "unknown message type"}}
	}
	reply := hub.Message{Type: hub.TypeStatusResult, PlaybackID: msg.PlaybackID}
	p, err := c.Playback.UpdateStatus(ctx, msg.PlaybackID, carepartnerID, msg.Status)
	if err != nil {
		reply.Error = publicError(err)
	} else {
		reply.Status = p.Status
	}
	return []hub.Message{reply}
}

// publicError maps service errors to messages safe to send to a client, never
// SQL or internal detail.
func publicError(err error) string {
	switch {
	case errors.Is(err, playbackSvc.ErrInvalidInput):
		return "invalid playback input"
	case errors.Is(err, playbackSvc.ErrNotFound):
		return "playback not found"
	case errors.Is(err, playbackSvc.ErrInvalidTransition):
		return "invalid playback status transition"
	default:
		return "could not save playback"
	}
}

type ConnectionResponse struct {
	CarepartnerID  int        `json:"carepartner_id"`
	Connected      bool       `json:"connected"`
	ConnectedSince *time.Time `json:"connected_since,omitempty"`
}

// ConnectionHandler godoc
// @Summary Whether a carepartner's tablet is connected
// @Tags realtime
// @Produce json
// @Param id path int true "Carepartner ID"
// @Success 200 {object} ConnectionResponse
// @Failure 400 {object} map[string]string
// @Router /realtime/carepartners/{id}/connection [get]
func (c *RealtimeController) ConnectionHandler(w http.ResponseWriter, r *http.Request) {
	id, ok := carepartnerFromPath(w, r)
	if !ok {
		return
	}
	resp := ConnectionResponse{CarepartnerID: id}
	if connected, since := c.Hub.Connected(id); connected {
		resp.Connected = true
		resp.ConnectedSince = &since
	}
	writeJSON(w, http.StatusOK, resp)
}

// RefreshHandler godoc
// @Summary Tell a carepartner's tablet to re-fetch its content
// @Description Use after changing a carepartner's schedule or media so the tablet picks it up immediately instead of on its next poll.
// @Tags realtime
// @Produce json
// @Param id path int true "Carepartner ID"
// @Success 200 {object} map[string]bool
// @Failure 400 {object} map[string]string
// @Router /realtime/carepartners/{id}/refresh [post]
func (c *RealtimeController) RefreshHandler(w http.ResponseWriter, r *http.Request) {
	id, ok := carepartnerFromPath(w, r)
	if !ok {
		return
	}
	delivered := c.Hub.Send(id, hub.Message{Type: hub.TypeRefresh})
	writeJSON(w, http.StatusOK, map[string]bool{"delivered": delivered})
}

func carepartnerFromPath(w http.ResponseWriter, r *http.Request) (int, bool) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || id <= 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid carepartner ID"})
		return 0, false
	}
	return id, true
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
