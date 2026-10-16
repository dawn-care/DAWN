package playback

import (
	playbackCtrl "server/internal/controllers/playback"

	"github.com/go-chi/chi/v5"
)

func RegisterPlaybackRoutes(r chi.Router, controller *playbackCtrl.PlaybackController) {
	// Tablet-reported plays and status changes.
	r.Post("/api/playbacks", controller.CreatePlaybackHandler)
	r.Patch("/api/playbacks/{id}/status", controller.UpdateStatusHandler)

	// Staff commands, pushed to the tablet over its WebSocket connection.
	r.Post("/api/playbacks/play", controller.PlayHandler)
	r.Post("/api/playbacks/{id}/stop", controller.StopHandler)
}
