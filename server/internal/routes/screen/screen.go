package screen

import (
	ctrl "server/internal/controllers/screen"

	"github.com/go-chi/chi/v5"
)

// RegisterScreenRoutes wires the tablet's screen endpoint
func RegisterScreenRoutes(r chi.Router, controller *ctrl.ScreenController) {
	r.Get("/api/screen", controller.GetScreenHandler) // GET /api/screen?carepartner_id=
}
