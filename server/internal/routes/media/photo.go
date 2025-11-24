package media

import (
	"server/internal/controllers/media"

	"github.com/go-chi/chi/v5"
)

// RegisterPhotoRoutes wires PhotoController handlers to HTTP routes
func RegisterPhotoRoutes(r chi.Router, photoCtrl *media.PhotoController) {
	r.Route("/photo", func(r chi.Router) {
		r.Post("/", photoCtrl.CreatePhotoHandler)
		r.Post("/upload", photoCtrl.CreatePhotoFileHandler)

		r.Get("/", photoCtrl.GetPhotoByIDHandler)                           // e.g., /photo?id=1
		r.Get("/file", photoCtrl.GetPhotoFileByIDHandler)                   // e.g., /photo/file?id=1
		r.Get("/list", photoCtrl.GetPhotosByCarepartnerAndCaregiverHandler) // e.g., /photo/list?carepartner_id=1&caregiver_id=2
		r.Get("/stats", photoCtrl.GetPhotoStatsHandler)

		r.Put("/", photoCtrl.UpdatePhotoHandler)
		r.Put("/file", photoCtrl.UpdatePhotoFileHandler)

		r.Delete("/", photoCtrl.DeletePhotoHandler)
		r.Delete("/file", photoCtrl.DeletePhotoFileHandler)
	})
}
