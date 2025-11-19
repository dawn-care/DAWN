package controllers

import (
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"server/internal/models/user"
	"server/internal/services"
)

type CaregiverController struct {
	Service *services.CaregiverService
}

// Constructor
func NewCaregiverController(service *services.CaregiverService) *CaregiverController {
	return &CaregiverController{
		Service: service,
	}
}

// ─────────────────────────────────────────────
// CREATE
// ─────────────────────────────────────────────

func (cc *CaregiverController) CreateCaregiver(c *fiber.Ctx) error {
	var caregiver user.Caregiver
	if err := c.BodyParser(&caregiver); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	if err := cc.Service.CreateCaregiver(&caregiver); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(caregiver)
}

// ─────────────────────────────────────────────
// READ
// ─────────────────────────────────────────────

func (cc *CaregiverController) GetCaregiverByID(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid ID"})
	}

	caregiver, err := cc.Service.GetCaregiverByID(id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	if caregiver == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "caregiver not found"})
	}

	return c.JSON(caregiver)
}

func (cc *CaregiverController) GetAllCaregivers(c *fiber.Ctx) error {
	list, err := cc.Service.GetAllCaregivers()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.JSON(list)
}

func (cc *CaregiverController) GetCaregiverByCarereceiverID(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("carereceiverId"))
	if err != nil || id <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid carereceiver ID"})
	}

	cg, err := cc.Service.GetCaregiverByCarereceiverID(id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	if cg == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "caregiver not found"})
	}

	return c.JSON(cg)
}

func (cc *CaregiverController) SearchCaregivers(c *fiber.Ctx) error {
	query := c.Query("q")
	if len(query) < 2 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "query must be at least 2 characters",
		})
	}

	results, err := cc.Service.SearchCaregivers(query)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(results)
}

// ─────────────────────────────────────────────
// UPDATE
// ─────────────────────────────────────────────

func (cc *CaregiverController) UpdateCaregiver(c *fiber.Ctx) error {
	var caregiver user.Caregiver

	if err := c.BodyParser(&caregiver); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	if caregiver.ID <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "missing or invalid caregiver ID",
		})
	}

	if err := cc.Service.UpdateCaregiver(&caregiver); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{"message": "caregiver updated successfully"})
}

// ─────────────────────────────────────────────
// DELETE
// ─────────────────────────────────────────────

func (cc *CaregiverController) DeleteCaregiver(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid ID"})
	}

	err = cc.Service.DeleteCaregiver(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "caregiver not found"})
		}
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "caregiver deleted successfully"})
}
