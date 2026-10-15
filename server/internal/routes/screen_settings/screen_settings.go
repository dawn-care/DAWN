package screen_settings

import (
	ctrl "server/internal/controllers/screen_settings"

	"github.com/go-chi/chi/v5"
)

// RegisterScreenSettingsRoutes wires ScreenSettingsController handlers to HTTP routes
func RegisterScreenSettingsRoutes(r chi.Router, controller *ctrl.ScreenSettingsController) {
	r.Route("/api/screen-settings", func(r chi.Router) {
		r.Get("/presets", controller.GetPresetsHandler)    // GET /api/screen-settings/presets
		r.Get("/", controller.GetScreenSettingsHandler)    // GET /api/screen-settings?carepartner_id=
		r.Put("/", controller.UpdateScreenSettingsHandler) // PUT /api/screen-settings?carepartner_id=
	})
}
