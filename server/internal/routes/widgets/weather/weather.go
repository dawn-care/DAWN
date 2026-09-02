package weather

import (
	weatherCtrlr "server/internal/controllers/widgets/weather"

	"github.com/go-chi/chi/v5"
)

// RegisterWeatherWidgetRoutes registers all weather widget routes
func RegisterWeatherWidgetRoutes(r chi.Router, controller *weatherCtrlr.WeatherWidgetController) {
	r.Route("/api/weather", func(r chi.Router) {

		// CRUD operations
		r.Post("/upload", controller.CreateWeatherWidgetHandler) // Create single weather widget
		// r.Post("/bulk", weatherCtrlr.CreateWeatherWidgetsBulkHandler) // TODO: Also missing bulk handler?
		r.Get("/", controller.GetWeatherWidgetByIDHandler)         // Get by ID (?id=)
		r.Put("/update", controller.UpdateWeatherWidgetHandler)    // Update
		r.Delete("/delete", controller.DeleteWeatherWidgetHandler) // Delete by ID (?id=)

		// Stats
		r.Get("/stats", controller.GetWeatherWidgetStatsHandler)

		// Custom retrievals
		r.Get("/by-carepartner-and-caregiver", controller.GetWeatherWidgetsByCarepartnerAndCaregiverHandler) // Get by caregiver and carepartner ?carepartner_id=&caregiver_id=
		r.Get("/by-day-and-carepartner", controller.GetWeatherByDayAndCarepartnerHandler)                    // Get by carepartner and day ?carepartner_id=&day=&month=&year=

		// Weather overview
		r.Post("/overview", controller.GetWeatherOverviewHandler) // POST body with WeatherWidget info (ZIPCODE etc.)
	})
}
