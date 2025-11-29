package modules

import (
	moduleCtrl "server/internal/controllers/modules"

	"github.com/go-chi/chi/v5"
)

// RegisterModulesRoutes registers all module-related routes
func RegisterModulesRoutes(r chi.Router, controller *moduleCtrl.ModuleController) {
	r.Route("/modules", func(r chi.Router) {

		// CREATE
		r.Post("/", controller.CreateModuleHandler)

		// READ
		r.Get("/{id}", controller.GetModuleByIDHandler)

		// Search module metadata by caregiver/carepartner/module_type
		r.Get("/metadata/search", controller.GetMetadataHandler)
	})
}
