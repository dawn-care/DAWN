package prompts

import (
	"encoding/json"
	"net/http"
	"strconv"

	models "server/internal/models/prompts"
)

// ---- Interface ----
type CaregiverLocationServiceInterface interface {
	CreateCaregiverLocation(loc *models.CaregiverLocationPrompt) error
	GetCaregiverLocationByID(id int) (*models.CaregiverLocationPrompt, error)
	UpdateCaregiverLocation(loc *models.CaregiverLocationPrompt) error
	DeleteCaregiverLocation(id int) error
}

// ---- Controller ----
type CaregiverLocationController struct {
	services CaregiverLocationServiceInterface
}

func NewCaregiverLocationController(services CaregiverLocationServiceInterface) *CaregiverLocationController {
	return &CaregiverLocationController{services: services}
}

// Create a new caregiver location
func (c *CaregiverLocationController) CreateCaregiverLocationHandler(w http.ResponseWriter, r *http.Request) {
	var loc models.CaregiverLocationPrompt

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

// Get caregiver location by ID
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

// Update caregiver location
func (c *CaregiverLocationController) UpdateCaregiverLocationHandler(w http.ResponseWriter, r *http.Request) {
	var loc models.CaregiverLocationPrompt

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

// Delete caregiver location
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
