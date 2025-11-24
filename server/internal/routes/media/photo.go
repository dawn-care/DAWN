package media

import (
	"encoding/json"
	"net/http"
	"strconv"

	models "server/internal/models/media"
	"server/internal/services/media"

	"github.com/go-chi/chi/v5"
)

func RegisterPhotoRoutes(r chi.Router, photoSvc *media.PhotoService) {
	r.Route("/photo", func(r chi.Router) {

		r.Post("/", func(w http.ResponseWriter, r *http.Request) {
			var p models.PhotoMedia
			if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			if err := photoSvc.CreatePhoto(&p); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			json.NewEncoder(w).Encode(p)
		})

		r.Post("/upload", func(w http.ResponseWriter, r *http.Request) {
			file, header, err := r.FormFile("file")
			if err != nil {
				http.Error(w, "file is required", http.StatusBadRequest)
				return
			}
			uri := r.FormValue("uri")
			if err := photoSvc.CreatePhotoFile(file, header, uri); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			json.NewEncoder(w).Encode(map[string]string{"uri": uri})
		})

		r.Get("/{id}", func(w http.ResponseWriter, r *http.Request) {
			id, _ := strconv.Atoi(chi.URLParam(r, "id"))
			p, err := photoSvc.GetPhotoByID(id)
			if err != nil {
				http.Error(w, err.Error(), http.StatusNotFound)
				return
			}
			json.NewEncoder(w).Encode(p)
		})

		r.Delete("/{id}", func(w http.ResponseWriter, r *http.Request) {
			id, _ := strconv.Atoi(chi.URLParam(r, "id"))
			if err := photoSvc.DeletePhoto(id); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		})
	})
}
