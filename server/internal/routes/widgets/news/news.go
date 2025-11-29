package news

import (
	newsCtrlr "server/internal/controllers/widgets/news"

	"github.com/go-chi/chi/v5"
)

// RegisterNewsWidgetRoutes registers all news widget routes
func RegisterNewsWidgetRoutes(r chi.Router, controller *newsCtrlr.NewsWidgetController) {
	r.Route("/news", func(r chi.Router) {

		// CRUD operations
		r.Post("/", controller.CreateNewsWidgetHandler)     // Create single news widget
		r.Post("/bulk", controller.CreateNewsWidgetHandler) // TODO: add bulk handler?
		r.Get("/", controller.GetNewsWidgetByIDHandler)     // Get by ID (query param ?id=)
		r.Put("/", controller.UpdateNewsWidgetHandler)      // Update widget
		r.Delete("/", controller.DeleteNewsWidgetHandler)   // Delete by ID (query param ?id=)

		// Stats
		r.Get("/stats", controller.GetNewsWidgetStatsHandler)

		// Custom retrievals
		r.Get("/by-carepartner-and-caregiver", controller.GetNewsWidgetsByCarepartnerAndCaregiverHandler) // ?carepartner_id=&caregiver_id=
		r.Get("/by-day-and-carepartner", controller.GetNewsByDayAndCarepartnerHandler)                    // ?carepartner_id=&day=&month=&year=

		// External APIs
		r.Get("/good-news", controller.GetGoodNewsHandler)
		r.Get("/joke", controller.GetJokeHandler)
	})
}
