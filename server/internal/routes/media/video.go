package media

import (
	"encoding/json"
	"net/http"
	"strconv"

	models "server/internal/models/media"
	"server/internal/services/media"

	"github.com/go-chi/chi/v5"
)

func RegisterVideoRoutes(r chi.Router, videoSvc *media.VideoService) {
	r.Route("/video", func(r chi.Router) {

		r.Post("/", func(w http.ResponseWriter, r *http.Request) {
			var v models.VideoMedia
			if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			if err := videoSvc.CreateVideo(&v); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			json.NewEncoder(w).Encode(v)
		})

		r.Post("/upload", func(w http.ResponseWriter, r *http.Request) {
			file, header, err := r.FormFile("file")
			if err != nil {
				http.Error(w, "file is required", http.StatusBadRequest)
				return
			}
			uri := r.FormValue("uri")
			if err := videoSvc.CreateVideoFile(file, header, uri); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			json.NewEncoder(w).Encode(map[string]string{"uri": uri})
		})

		r.Get("/{id}", func(w http.ResponseWriter, r *http.Request) {
			id, _ := strconv.Atoi(chi.URLParam(r, "id"))
			v, err := videoSvc.GetVideoByID(id)
			if err != nil {
				http.Error(w, err.Error(), http.StatusNotFound)
				return
			}
			json.NewEncoder(w).Encode(v)
		})

		r.Delete("/{id}", func(w http.ResponseWriter, r *http.Request) {
			id, _ := strconv.Atoi(chi.URLParam(r, "id"))
			if err := videoSvc.DeleteVideo(id); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		})
	})
}
