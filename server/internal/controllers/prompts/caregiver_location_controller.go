package prompts

import (
	"encoding/json"
	"net/http"
	"strconv"

	"server/internal/models/prompts"
)

// ---- Interface ----
type CaregiverLocationServiceInterface interface {
	CreateCaregiverLocation(loc *prompts.CaregiverLocationPrompt) error
	GetCaregiverLocationByID(id int) (*prompts.CaregiverLocationPrompt, error)
	UpdateCaregiverLocation(loc *prompts.CaregiverLocationPrompt) error
	DeleteCaregiverLocation(id int) error
}

// ---- Controller ----
type CaregiverLocationController struct {
	services CaregiverLocationServiceInterface
}

func NewCaregiverLocationController(services CaregiverLocationServiceInterface) *CaregiverLocationController {
	return &CaregiverLocationController{services: services}
}

// CreateCaregiverLocationHandler godoc
// @Summary Create a new caregiver location
// @Description Creates a new caregiver location record
// @Tags caregiver-location
// @Accept json
// @Produce json
// @Param location body prompts.CaregiverLocationPrompt true "Caregiver location data"
// @Success 200 {object} prompts.CaregiverLocationPrompt
// @Failure 400 {string} string "invalid request body"
// @Failure 500 {string} string "internal server error"
// @Router /caregiver-location [post]
func (c *CaregiverLocationController) CreateCaregiverLocationHandler(w http.ResponseWriter, r *http.Request) {
	var loc prompts.CaregiverLocationPrompt

	if err := json.NewDecoder(r.Body).Decode(&loc); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := c.services.CreateCaregiverLocation(&loc); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(loc)
}

// GetCaregiverLocationByIDHandler godoc
// @Summary Get a caregiver location by ID
// @Description Retrieves a caregiver location record by its unique ID
// @Tags caregiver-location
// @Accept json
// @Produce json
// @Param id query int true "Caregiver location ID"
// @Success 200 {object} prompts.CaregiverLocationPrompt
// @Failure 400 {string} string "invalid caregiver location ID"
// @Failure 404 {string} string "caregiver location not found"
// @Failure 500 {string} string "internal server error"
// @Router /caregiver-location [get]
func (c *CaregiverLocationController) GetCaregiverLocationByIDHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		http.Error(w, "invalid caregiver location ID", http.StatusBadRequest)
		return
	}

	loc, err := c.services.GetCaregiverLocationByID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if loc == nil {
		http.Error(w, "caregiver location not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(loc)
}

// UpdateCaregiverLocationHandler godoc
// @Summary Update an existing caregiver location
// @Description Updates a caregiver location record by ID
// @Tags caregiver-location
// @Accept json
// @Produce json
// @Param location body prompts.CaregiverLocationPrompt true "Updated caregiver location data (must include ID)"
// @Success 200 {object} prompts.CaregiverLocationPrompt
// @Failure 400 {string} string "invalid request body or missing ID"
// @Failure 500 {string} string "internal server error"
// @Router /caregiver-location [put]
func (c *CaregiverLocationController) UpdateCaregiverLocationHandler(w http.ResponseWriter, r *http.Request) {
	var loc prompts.CaregiverLocationPrompt

	if err := json.NewDecoder(r.Body).Decode(&loc); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if loc.ID <= 0 {
		http.Error(w, "location ID is required", http.StatusBadRequest)
		return
	}

	if err := c.services.UpdateCaregiverLocation(&loc); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(loc)
}

// DeleteCaregiverLocationHandler godoc
// @Summary Delete a caregiver location
// @Description Deletes a caregiver location record by its unique ID
// @Tags caregiver-location
// @Accept json
// @Produce json
// @Param id query int true "Caregiver location ID"
// @Success 204 {string} string "No Content"
// @Failure 400 {string} string "invalid caregiver location ID"
// @Failure 500 {string} string "internal server error"
// @Router /caregiver-location [delete]
func (c *CaregiverLocationController) DeleteCaregiverLocationHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil || id <= 0 {
		http.Error(w, "invalid caregiver location ID", http.StatusBadRequest)
		return
	}

	if err := c.services.DeleteCaregiverLocation(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
