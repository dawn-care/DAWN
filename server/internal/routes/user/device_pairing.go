package user

import (
	dpController "server/internal/controllers/user/device_pairing"

	"github.com/go-chi/chi/v5"
)

// RegisterDevicePairingRoutes registers device pairing endpoints
func RegisterDevicePairingRoutes(r chi.Router, controller *dpController.DevicePairingController) {
	r.Route("/api/users/device-pairing", func(r chi.Router) {
		// Generate pairing code for a carepartner device
		r.Post("/code", controller.GeneratePairingCode) // POST /users/device-pairing/code

		// Caregiver enters code to pair device
		r.Post("/verify", controller.VerifyPairingCode) // POST /users/device-pairing/verify

		// Check if carepartner device is already paired
		r.Post("/status", controller.IsDevicePaired) // POST /users/device-pairing/status
	})
}
