package scheduled_items

import (
	ctrl "server/internal/controllers/scheduled_items"

	"github.com/go-chi/chi/v5"
)

// RegisterScheduledItemRoutes wires ScheduledItemController handlers to HTTP routes
func RegisterScheduledItemRoutes(r chi.Router, controller *ctrl.ScheduledItemController) {
	r.Route("/api/scheduled-items", func(r chi.Router) {
		r.Get("/", controller.ListItemsHandler)   // GET /api/scheduled-items?carepartner_id=
		r.Post("/", controller.CreateItemHandler) // POST /api/scheduled-items

		r.Get("/{id}", controller.GetItemHandler)
		r.Put("/{id}", controller.UpdateItemHandler)
		r.Delete("/{id}", controller.DeleteItemHandler) // ?scope=series | ?scope=occurrence&date=YYYY-MM-DD
	})

	r.Get("/api/calendar", controller.CalendarHandler) // ?carepartner_id=&from=YYYY-MM-DD&to=YYYY-MM-DD
}
