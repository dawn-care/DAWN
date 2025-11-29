package prompts

import (
	ctrlr "server/internal/controllers/prompts"

	"github.com/go-chi/chi/v5"
)

// RegisterMessageRoutes registers all routes for messages
func RegisterMessageRoutes(r chi.Router, controller *ctrlr.MessageController) {
	r.Route("/messages", func(r chi.Router) {

		// Create a new message
		r.Post("/", controller.CreateMessageHandler)

		// Get a message by ID
		r.Get("/", controller.GetMessageByIDHandler)

		// Update a message
		r.Put("/", controller.UpdateMessageHandler)

		// Delete a message by ID
		r.Delete("/", controller.DeleteMessageHandler)
	})
}
