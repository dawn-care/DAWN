package routes

import (
	moduleCtrl "server/internal/controllers/modules"
	db "server/internal/db"
	moduleSvc "server/internal/services/modules"

	"github.com/go-chi/chi/v5"
)

func RegisterModuleRoutes(r chi.Router) {
	r.Route("/modules", func(r chi.Router) {
		svc := moduleSvc.NewModuleService(db.GetDB())
		ctrl := moduleCtrl.NewModuleController(svc)

		// CREATE
		r.Post("/", ctrl.CreateModuleHandler)

		// READ
		r.Get("/{id}", ctrl.GetModuleByIDHandler)

		// Search for metadata list by caregiver/carepartner/module_type
		// Example:
		// /modules/metadata/search?caregiver_id=1&carepartner_id=2&module_type=audio
		r.Get("/metadata/search", ctrl.GetMetadataHandler)
	})
}
