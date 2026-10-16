package screen_widgets

import (
	ctrl "server/internal/controllers/screen_widgets"

	"github.com/go-chi/chi/v5"
)

// RegisterScreenWidgetRoutes wires ScreenWidgetController handlers to HTTP routes
func RegisterScreenWidgetRoutes(r chi.Router, controller *ctrl.ScreenWidgetController) {
	r.Route("/api/screen-widgets", func(r chi.Router) {
		r.Get("/", controller.ListWidgetsHandler)   // GET /api/screen-widgets?carepartner_id=
		r.Post("/", controller.CreateWidgetHandler) // POST /api/screen-widgets

		r.Get("/{id}", controller.GetWidgetHandler)
		r.Put("/{id}", controller.UpdateWidgetHandler)
		r.Post("/{id}/move", controller.MoveWidgetHandler)
		r.Delete("/{id}", controller.DeleteWidgetHandler)
	})
}
