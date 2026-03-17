package configurations

import (
	configurationCtrl "server/internal/controllers/configurations"

	"github.com/go-chi/chi/v5"
)

// RegisterConfigurationsRoutes registers all configuration-related routes
func RegisterConfigurationsRoutes(r chi.Router, controller *configurationCtrl.ConfigurationController) {
	r.Route("/configurations", func(r chi.Router) {

		// CREATE
		r.Post("/", controller.CreateConfigurationHandler)
		r.Post("/insert", controller.InsertConfigurationsHandler)

		// READ
		r.Get("/{id}", controller.GetConfigurationByIDHandler)
		r.Get("/day", controller.GetConfigurationsByDayAndCarepartnerHandler)
		r.Get("/carepartner", controller.GetCarePartnerConfigurationsHandler)
		r.Get("/stats", controller.GetConfigurationStatsHandler)

		// UPDATE
		r.Put("/", controller.UpdateConfigurationHandler)

		// DELETE
		r.Delete("/{id}", controller.DeleteConfigurationHandler)
		r.Post("/delete", controller.DeleteConfigurationsHandler)
	})
}
