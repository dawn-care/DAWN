package prompts

import (
	ctrlr "server/internal/controllers/prompts"

	"github.com/go-chi/chi/v5"
)

// RegisterEventRoutes registers all routes for event prompts
func RegisterEventRoutes(r chi.Router, controller *ctrlr.EventController) {
	r.Route("/api/events", func(r chi.Router) {

		// Create a new list of events
		r.Post("/", controller.CreateEventListHandler)

		// Get events by list ID
		r.Get("/", controller.GetEventListByListIDHandler)

		// Update a specific event
		r.Put("/", controller.UpdateEventHandler)

		// Delete a specific event
		r.Delete("/", controller.DeleteEventHandler)

		// Delete all events by list ID
		r.Delete("/list", controller.DeleteEventListHandler)
	})
}
