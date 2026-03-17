package media

import (
	"server/internal/controllers/media"

	"github.com/go-chi/chi/v5"
)

// RegisterAudioRoutes wires AudioController handlers to HTTP routes
func RegisterAudioRoutes(r chi.Router, audioCtrl *media.AudioController) {
	r.Route("/audio", func(r chi.Router) {
		// CREATE
		r.Post("/", audioCtrl.CreateAudioHandler)
		r.Post("/upload", audioCtrl.CreateAudioFileHandler)

		// READ
		r.Get("/", audioCtrl.GetAudioByIDHandler)
		r.Get("/file", audioCtrl.GetAudioFileByIDHandler)
		r.Get("/list", audioCtrl.GetAudiosByCarepartnerAndCaregiverHandler)

		// UPDATE
		r.Put("/", audioCtrl.UpdateAudioHandler)
		r.Put("/file", audioCtrl.UpdateAudioFileHandler)

		// DELETE
		r.Delete("/", audioCtrl.DeleteAudioHandler)
		r.Delete("/file", audioCtrl.DeleteAudioFileHandler)
	})
}
