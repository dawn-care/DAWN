package media

import (
	"server/internal/controllers/media"

	"github.com/go-chi/chi/v5"
)

func RegisterAudioRoutes(r chi.Router, audioCtrl *media.AudioController) {
	// Audio routes
	r.Route("/audio", func(r chi.Router) {
		r.Post("/", audioCtrl.CreateAudioHandler)
		r.Post("/upload", audioCtrl.CreateAudioFileHandler)
		r.Get("/", audioCtrl.GetAudioByIDHandler)
		r.Get("/file", audioCtrl.GetAudioFileByIDHandler)
		// r.Get("/list", audioCtrl.GetAudiosByCarepartnerAndCaregiverHandler)
		r.Put("/", audioCtrl.UpdateAudioHandler)
		r.Put("/file", audioCtrl.UpdateAudioFileHandler)
		r.Delete("/", audioCtrl.DeleteAudioHandler)
		r.Delete("/file", audioCtrl.DeleteAudioFileHandler)
		r.Get("/search", audioCtrl.SearchAudiosHandler)
		r.Get("/stats", audioCtrl.GetMediaStatsHandler)
	})
}
