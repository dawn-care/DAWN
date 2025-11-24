package user

import (
	cgController "server/internal/controllers/user/caregiver"

	"github.com/go-chi/chi/v5"
)

// RegisterCaregiverRoutes registers all caregiver-related routes
func RegisterCaregiverRoutes(r chi.Router, controller *cgController.CaregiverController) {
	r.Route("/caregivers", func(r chi.Router) {

		r.Post("/", controller.CreateCaregiver)                                          // POST /users/caregivers
		r.Get("/", controller.GetAllCaregivers)                                          // GET /users/caregivers
		r.Get("/search", controller.SearchCaregivers)                                    // GET /users/caregivers/search?q=keyword
		r.Get("/{id}", controller.GetCaregiverByID)                                      // GET /users/caregivers/{id}
		r.Get("/carereceiver/{carereceiverId}", controller.GetCaregiverByCarereceiverID) // GET /users/caregivers/carereceiver/{id}
		r.Put("/", controller.UpdateCaregiver)                                           // PUT /users/caregivers
		r.Delete("/{id}", controller.DeleteCaregiver)                                    // DELETE /users/caregivers/{id}
	})
}
