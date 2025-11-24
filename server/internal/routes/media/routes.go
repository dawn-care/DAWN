package media

import (
	"server/internal/services/media"

	"github.com/go-chi/chi/v5"
)

func RegisterMediaRoutes(r chi.Router, audioSvc *media.AudioService, photoSvc *media.PhotoService, videoSvc *media.VideoService) {
	RegisterAudioRoutes(r, audioSvc)
	RegisterPhotoRoutes(r, photoSvc)
	RegisterVideoRoutes(r, videoSvc)
}
