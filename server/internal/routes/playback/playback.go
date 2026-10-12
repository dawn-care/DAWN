package playback

import (
	playbackCtrl "server/internal/controllers/playback"

	"github.com/go-chi/chi/v5"
)

func RegisterPlaybackRoutes(r chi.Router, controller *playbackCtrl.PlaybackController) {
	r.Post("/api/playbacks", controller.CreatePlaybackHandler)
	r.Patch("/api/playbacks/{id}/status", controller.UpdateStatusHandler)
}
