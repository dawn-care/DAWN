package family

import (
	ctrl "server/internal/controllers/family"

	"github.com/go-chi/chi/v5"
)

// RegisterFamilyRoutes wires FamilyController handlers to HTTP routes
func RegisterFamilyRoutes(r chi.Router, controller *ctrl.FamilyController) {
	r.Route("/api/family-members", func(r chi.Router) {
		r.Get("/", controller.ListFamilyMembersHandler)   // GET /api/family-members?carepartner_id=
		r.Post("/", controller.CreateFamilyMemberHandler) // POST /api/family-members
	})
}
