package media

import (
	"server/internal/controllers/media"

	"github.com/go-chi/chi/v5"
)

// RegisterPhotoRoutes wires PhotoController handlers to HTTP routes
func RegisterPhotoRoutes(r chi.Router, photoCtrl *media.PhotoController) {
	r.Route("/api/photo", func(r chi.Router) {
		// CREATE
		r.Post("/", photoCtrl.CreatePhotoHandler)
		r.Post("/upload", photoCtrl.CreatePhotoFileHandler)

		// READ
		r.Get("/", photoCtrl.GetPhotoByIDHandler)
		r.Get("/file", photoCtrl.GetPhotoFileByIDHandler)
		r.Get("/list", photoCtrl.GetPhotosByCarepartnerAndCaregiverHandler)

		// UPDATE
		r.Put("/", photoCtrl.UpdatePhotoHandler)
		r.Put("/file", photoCtrl.UpdatePhotoFileHandler)

		// DELETE
		r.Delete("/", photoCtrl.DeletePhotoHandler)
		r.Delete("/file", photoCtrl.DeletePhotoFileHandler)
	})
}
