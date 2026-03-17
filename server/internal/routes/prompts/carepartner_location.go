package prompts

import (
	prompts "server/internal/controllers/prompts"

	"github.com/go-chi/chi/v5"
)

// RegisterCarepartnerLocationRoutes sets up all routes for carepartner location prompts.
func RegisterCarepartnerLocationRoutes(r chi.Router, controller *prompts.CarepartnerLocationController) {
	r.Route("/carepartner-location", func(r chi.Router) {

		// Create a new carepartner location
		r.Post("/", controller.CreateCarepartnerLocationHandler)

		// Get carepartner location by ID (expects ?id=123)
		r.Get("/", controller.GetCarepartnerLocationByIDHandler)

		// Update a carepartner location
		r.Put("/", controller.UpdateCarepartnerLocationHandler)

		// Delete a carepartner location
		r.Delete("/", controller.DeleteCarepartnerLocationHandler)
	})
}
