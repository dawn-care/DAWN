package media

import (
	"server/internal/controllers/media"

	"github.com/go-chi/chi/v5"
)

// RegisterVideoRoutes wires VideoController handlers to HTTP routes
func RegisterVideoRoutes(r chi.Router, videoCtrl *media.VideoController) {
	r.Route("/video", func(r chi.Router) {
		r.Post("/", videoCtrl.CreateVideoHandler)
		r.Post("/upload", videoCtrl.CreateVideoFileHandler)

		r.Get("/", videoCtrl.GetVideoByIDHandler)                           // e.g., /video?id=1
		r.Get("/file", videoCtrl.GetVideoFileByIDHandler)                   // e.g., /video/file?id=1
		r.Get("/list", videoCtrl.GetVideosByCarepartnerAndCaregiverHandler) // e.g., /video/list?carepartner_id=1&caregiver_id=2
		r.Get("/stats", videoCtrl.GetVideoStatsHandler)

		r.Put("/", videoCtrl.UpdateVideoHandler)
		r.Put("/file", videoCtrl.UpdateVideoFileHandler)

		r.Delete("/", videoCtrl.DeleteVideoHandler)
		r.Delete("/file", videoCtrl.DeleteVideoFileHandler)
	})
}
