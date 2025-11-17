package routes

import (
	"github.com/gofiber/fiber/v2"
	"server/internal/controllers"
	"server/internal/services"
)

func CaregiverRoutes(app *fiber.App, service *services.CaregiverService) {
	controller := controllers.NewCaregiverController(service)

	api := app.Group("/caregivers")

	api.Post("/", controller.CreateCaregiver)
	api.Get("/", controller.GetAllCaregivers)
	api.Get("/search", controller.SearchCaregivers)
	api.Get("/:id", controller.GetCaregiverByID)
	api.Get("/carereceiver/:carereceiverId", controller.GetCaregiverByCarereceiverID)
	api.Put("/", controller.UpdateCaregiver)
	api.Delete("/:id", controller.DeleteCaregiver)
}
