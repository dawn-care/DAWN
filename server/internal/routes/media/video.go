package media

import (
	"server/internal/controllers/media"

	"github.com/go-chi/chi/v5"
)

// RegisterVideoRoutes wires VideoController handlers to HTTP routes
func RegisterVideoRoutes(r chi.Router, videoCtrl *media.VideoController) {
	r.Route("/api/video", func(r chi.Router) {
		// CREATE
		r.Post("/", videoCtrl.CreateVideoHandler)
		r.Post("/upload", videoCtrl.CreateVideoFileHandler)

		// READ
		r.Get("/", videoCtrl.GetVideoByIDHandler)
		r.Get("/file", videoCtrl.GetVideoFileByIDHandler)
		r.Get("/list", videoCtrl.GetVideosByCarepartnerAndCaregiverHandler)

		// UPDATE
		r.Put("/", videoCtrl.UpdateVideoHandler)
		r.Put("/file", videoCtrl.UpdateVideoFileHandler)

		// DELETE
		r.Delete("/", videoCtrl.DeleteVideoHandler)
		r.Delete("/file", videoCtrl.DeleteVideoFileHandler)
	})
}
