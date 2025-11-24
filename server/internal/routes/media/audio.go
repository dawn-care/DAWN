package media

import (
	"encoding/json"
	"net/http"
	"strconv"

	models "server/internal/models/media"
	"server/internal/services/media"

	"github.com/go-chi/chi/v5"
)

func RegisterAudioRoutes(r chi.Router, audioSvc *media.AudioService) {
	r.Route("/audio", func(r chi.Router) {
		r.Post("/", func(w http.ResponseWriter, r *http.Request) {
			var a models.AudioMedia
			if err := json.NewDecoder(r.Body).Decode(&a); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			if err := audioSvc.CreateAudio(&a); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			json.NewEncoder(w).Encode(a)
		})

		r.Post("/upload", func(w http.ResponseWriter, r *http.Request) {
			file, header, err := r.FormFile("file")
			if err != nil {
				http.Error(w, "file is required", http.StatusBadRequest)
				return
			}
			uri := r.FormValue("uri")
			if err := audioSvc.CreateAudioFile(file, header, uri); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			json.NewEncoder(w).Encode(map[string]string{"uri": uri})
		})

		r.Get("/{id}", func(w http.ResponseWriter, r *http.Request) {
			id, _ := strconv.Atoi(chi.URLParam(r, "id"))
			a, err := audioSvc.GetAudioByID(id)
			if err != nil {
				http.Error(w, err.Error(), http.StatusNotFound)
				return
			}
			json.NewEncoder(w).Encode(a)
		})

		r.Delete("/{id}", func(w http.ResponseWriter, r *http.Request) {
			id, _ := strconv.Atoi(chi.URLParam(r, "id"))
			if err := audioSvc.DeleteAudio(id); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		})
	})
}
