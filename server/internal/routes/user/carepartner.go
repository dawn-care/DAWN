package user

import (
	ctrl "server/internal/controllers/user/carepartner"

	"github.com/go-chi/chi/v5"
)

func RegisterCarepartnerRoutes(r chi.Router, controller *ctrl.CarepartnerController) {
	r.Route("/api/carepartner", func(r chi.Router) {
		r.Post("/create-carepartner", controller.CreateCarepartner) // POST /api/users/carepartners
		r.Get("/{id}", controller.GetCarepartnerByID)               // GET /api/users/carepartners/{id}
		r.Get("/", controller.GetAllCarepartners)                   // GET /api/users/carepartners
		r.Delete("/{id}", controller.DeleteCarepartner)             // DELETE /api/users/carepartners/{id}
	})
}
