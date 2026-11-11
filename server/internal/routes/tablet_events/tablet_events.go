package tablet_events

import (
	tabletEventsCtrl "server/internal/controllers/tablet_events"

	"github.com/go-chi/chi/v5"
)

func RegisterTabletEventRoutes(r chi.Router, touches *tabletEventsCtrl.TouchController) {
	// The resident's touches, sent in batches by the tablet
	r.Post("/api/tablet-events/touches", touches.RecordTouchesHandler)
}
