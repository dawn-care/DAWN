package screen_widgets

import (
	ctrl "server/internal/controllers/screen_widgets"

	"github.com/go-chi/chi/v5"
)

// RegisterScreenWidgetRoutes wires ScreenWidgetController handlers to HTTP routes
func RegisterScreenWidgetRoutes(r chi.Router, controller *ctrl.ScreenWidgetController) {
	r.Route("/api/screen-layout", func(r chi.Router) {
		r.Get("/", controller.GetLayoutHandler)  // GET /api/screen-layout?carepartner_id=
		r.Put("/", controller.SaveLayoutHandler) // PUT /api/screen-layout?carepartner_id=
	})
}
