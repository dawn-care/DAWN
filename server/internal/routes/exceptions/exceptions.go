package exceptions

import (
	exceptionCtrl "server/internal/controllers/exceptions"

	"github.com/go-chi/chi/v5"
)

// RegisterExceptionRoutes registers all exception-related routes
func RegisterExceptionRoutes(r chi.Router, controller *exceptionCtrl.ExceptionController) {
	r.Route("/api/exceptions", func(r chi.Router) {
		r.Post("/", controller.CreateExceptionHandler)
		r.Get("/{configuration_id}", controller.GetExceptionsByConfigurationIDHandler)
		r.Delete("/{configuration_id}", controller.DeleteExceptionsByConfigurationIDHandler)
	})
}
