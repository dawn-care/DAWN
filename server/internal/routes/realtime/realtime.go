package realtime

import (
	realtimeCtrl "server/internal/controllers/realtime"

	"github.com/go-chi/chi/v5"
)

func RegisterRealtimeRoutes(r chi.Router, controller *realtimeCtrl.RealtimeController) {
	r.Get("/api/realtime/tablet", controller.ConnectHandler)
	r.Get("/api/realtime/carepartners/{id}/connection", controller.ConnectionHandler)
	r.Post("/api/realtime/carepartners/{id}/refresh", controller.RefreshHandler)
}
