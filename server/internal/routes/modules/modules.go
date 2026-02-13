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

		// GET
		r.Get("/", controller.GetMetadataHandler)
		r.Get("/{id}", controller.GetModuleByIDHandler)
		r.Get("/metadata/search", controller.GetMetadataHandler)
		r.Get("/metadata/all", controller.GetAllMetadataHandler)
		r.Get("/", controller.GetModulesByCaregiverAndCarepartnerHandler)

		// DELETE
		r.Delete("/delete", controller.DeleteModuleHandler)
	})
}
