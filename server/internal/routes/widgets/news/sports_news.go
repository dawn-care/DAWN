package news

import (
	newsCtrlr "server/internal/controllers/widgets/news"

	"github.com/go-chi/chi/v5"
)

// RegisterSportsNewsWidgetRoutes registers all sports news widget routes
func RegisterSportsNewsWidgetRoutes(r chi.Router, controller *newsCtrlr.SportsNewsWidgetController) {
	r.Route("/sports", func(r chi.Router) {

		// CRUD operations
		r.Post("/upload", controller.CreateSportsNewsWidgetHandler) // Create single sports news widget
		// r.Post("/bulk", newsService.CreateSportsNewsWidgetsBulk) // TODO: Missing bulk create handler (unless it is supposed to use the news one?)
		r.Get("/", controller.GetSportsNewsWidgetByIDHandler)         // Get by ID (?id=)
		r.Put("/update", controller.UpdateSportsNewsWidgetHandler)    // Update
		r.Delete("/delete", controller.DeleteSportsNewsWidgetHandler) // Delete by ID (?id=)

		// Stats
		r.Get("/stats", controller.GetSportsNewsWidgetStatsHandler)

		// Custom retrievals
		r.Get("/by-carepartner-and-caregiver", controller.GetSportsNewsWidgetsByCarepartnerAndCaregiverHandler) // ?carepartner_id=&caregiver_id=
		r.Get("/by-day-and-carepartner", controller.GetSportsNewsByDayAndCarepartnerHandler)                    // ?carepartner_id=&day=&month=&year=

		// Team-specific game endpoint
		r.Get("/team-game", controller.GetTeamGameHandler) // ?league=&team=
	})
}
