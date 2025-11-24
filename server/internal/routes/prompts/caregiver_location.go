package prompts

import (
	prompts "server/internal/services/prompts"

	"github.com/go-chi/chi/v5"
)

// RegisterCaregiverLocationRoutes sets up all routes for caregiver location prompts.
func RegisterCaregiverLocationRoutes(r chi.Router, controller *prompts.CaregiverLocationController) {
	r.Route("/caregiver-location", func(r chi.Router) {

		// Create a new caregiver location
		r.Post("/", controller.CreateCaregiverLocationHandler)

		// Get caregiver location by ID
		r.Get("/", controller.GetCaregiverLocationByIDHandler)

		// Update a caregiver location
		r.Put("/", controller.UpdateCaregiverLocationHandler)

		// Delete a caregiver location
		r.Delete("/", controller.DeleteCaregiverLocationHandler)
	})
}
