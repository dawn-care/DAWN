package routes

import (
	"database/sql"
	cgController "server/internal/controllers/user/caregiver"
	cgService "server/internal/services/user/caregiver"

	"github.com/go-chi/chi/v5"
)

func RegisterCaregiverRoutes(db *sql.DB) chi.Router {
	r := chi.NewRouter()

	// Initialize caregiver service & controller
	cSvc := cgService.NewCaregiverService(db)
	cCtrlr := cgController.NewCaregiverController(cSvc)

	// Caregiver routes
	r.Route("/caregivers", func(r chi.Router) {
		r.Post("/", cCtrlr.CreateCaregiver)                                          // POST /users/caregivers
		r.Get("/", cCtrlr.GetAllCaregivers)                                          // GET /users/caregivers
		r.Get("/search", cCtrlr.SearchCaregivers)                                    // GET /users/caregivers/search?q=keyword
		r.Get("/{id}", cCtrlr.GetCaregiverByID)                                      // GET /users/caregivers/{id}
		r.Get("/carereceiver/{carereceiverId}", cCtrlr.GetCaregiverByCarereceiverID) // GET /users/caregivers/carereceiver/{id}
		r.Put("/", cCtrlr.UpdateCaregiver)                                           // PUT /users/caregivers
		r.Delete("/{id}", cCtrlr.DeleteCaregiver)                                    // DELETE /users/caregivers/{id}
	})

	return r
}
