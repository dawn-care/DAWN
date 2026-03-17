package prompts

import (
	ctrlr "server/internal/controllers/prompts"

	"github.com/go-chi/chi/v5"
)

// RegisterMessageRoutes registers all routes for messages
func RegisterMessageRoutes(r chi.Router, controller *ctrlr.MessageController) {
	r.Route("/api/messages", func(r chi.Router) {

		// Create a new message
		r.Post("/", controller.CreateMessageHandler)

		// Get a message by ID
		r.Get("/", controller.GetMessageByIDHandler) // Get by ID (?id=)

		// Update a message
		r.Put("/update/{id}", controller.UpdateMessageHandler)

		// Delete a message by ID
		r.Delete("/delete/{id}", controller.DeleteMessageHandler)
	})
}
