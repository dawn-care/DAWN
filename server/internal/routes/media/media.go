package media

import (
	"server/internal/controllers/media"

	"github.com/go-chi/chi/v5"
)

// RegisterMediaRoutes wires MediaController handlers to HTTP routes
func RegisterMediaRoutes(r chi.Router, mediaCtrl *media.MediaController) {
	r.Route("/api/media", func(r chi.Router) {
		r.Post("/", mediaCtrl.UploadMediaHandler)
		r.Get("/", mediaCtrl.ListMediaHandler)

		r.Get("/{id}", mediaCtrl.GetMediaHandler)
		r.Get("/{id}/file", mediaCtrl.GetMediaFileHandler)
		r.Patch("/{id}", mediaCtrl.RenameMediaHandler)
		r.Put("/{id}/tags", mediaCtrl.SetMediaTagsHandler)
		r.Delete("/{id}", mediaCtrl.DeleteMediaHandler)
	})
}
