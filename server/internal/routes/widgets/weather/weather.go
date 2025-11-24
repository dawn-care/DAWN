package weather

import (
	weatherCtrlr "server/internal/controllers/widgets/weather"

	"github.com/go-chi/chi/v5"
)

// RegisterWeatherWidgetRoutes registers all weather widget routes
func RegisterWeatherWidgetRoutes(r chi.Router, controller *weatherCtrlr.WeatherWidgetController) {
	r.Route("/weather", func(r chi.Router) {

		// CRUD operations
		r.Post("/", controller.CreateWeatherWidgetHandler) // Create single weather widget
		// r.Post("/bulk", weatherCtrlr.CreateWeatherWidgetsBulkHandler) // TODO: Also missing bulk handler?
		r.Get("/", controller.GetWeatherWidgetByIDHandler)   // Get by ID (?id=)
		r.Put("/", controller.UpdateWeatherWidgetHandler)    // Update
		r.Delete("/", controller.DeleteWeatherWidgetHandler) // Delete by ID (?id=)

		// Stats
		r.Get("/stats", controller.GetWeatherWidgetStatsHandler)

		// Custom retrievals
		r.Get("/by-carepartner-and-caregiver", controller.GetWeatherWidgetsByCarepartnerAndCaregiverHandler) // ?carepartner_id=&caregiver_id=
		r.Get("/by-day-and-carepartner", controller.GetWeatherByDayAndCarepartnerHandler)                    // ?carepartner_id=&day=&month=&year=

		// Weather overview
		r.Post("/overview", controller.GetWeatherOverviewHandler) // POST body with WeatherWidget info (ZIPCODE etc.)
	})
}
