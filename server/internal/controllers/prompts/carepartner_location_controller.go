package prompts

import (
	"encoding/json"
	"net/http"
	"strconv"

	"server/internal/models/prompts"
)

// ---- Interface ----
type CarepartnerLocationServiceInterface interface {
	CreateCarepartnerLocation(loc *prompts.CarepartnerLocationPrompt) error
	GetCarepartnerLocationByID(id int) (*prompts.CarepartnerLocationPrompt, error)
	UpdateCarepartnerLocation(loc *prompts.CarepartnerLocationPrompt) error
	DeleteCarepartnerLocation(id int) error
}

// ---- Controller ----
type CarepartnerLocationController struct {
	services CarepartnerLocationServiceInterface
}

func NewCarepartnerLocationController(services CarepartnerLocationServiceInterface) *CarepartnerLocationController {
	return &CarepartnerLocationController{services: services}
}

// CreateCarepartnerLocationHandler godoc
// @Summary Create a new carepartner location
// @Description Creates a new carepartner location record
// @Tags carepartner-location
// @Accept json
// @Produce json
// @Param location body prompts.CarepartnerLocationPrompt true "Carepartner location data"
// @Success 200 {object} prompts.CarepartnerLocationPrompt
// @Failure 400 {string} string "invalid request body"
// @Failure 500 {string} string "internal server error"
// @Router /carepartner-location [post]
func (c *CarepartnerLocationController) CreateCarepartnerLocationHandler(w http.ResponseWriter, r *http.Request) {
	var loc prompts.CarepartnerLocationPrompt

	if err := json.NewDecoder(r.Body).Decode(&loc); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := c.services.CreateCarepartnerLocation(&loc); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(loc)
}

// GetCarepartnerLocationByIDHandler godoc
// @Summary Get a carepartner location by ID
// @Description Retrieves a carepartner location record by its unique ID
// @Tags carepartner-location
// @Accept json
// @Produce json
// @Param id query int true "Carepartner location ID"
// @Success 200 {object} prompts.CarepartnerLocationPrompt
// @Failure 400 {string} string "invalid carepartner location ID"
// @Failure 404 {string} string "carepartner location not found"
// @Failure 500 {string} string "internal server error"
// @Router /carepartner-location [get]
func (c *CarepartnerLocationController) GetCarepartnerLocationByIDHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		http.Error(w, "invalid carepartner location ID", http.StatusBadRequest)
		return
	}

	loc, err := c.services.GetCarepartnerLocationByID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if loc == nil {
		http.Error(w, "carepartner location not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(loc)
}

// UpdateCarepartnerLocationHandler godoc
// @Summary Update an existing carepartner location
// @Description Updates a carepartner location record by ID
// @Tags carepartner-location
// @Accept json
// @Produce json
// @Param location body prompts.CarepartnerLocationPrompt true "Updated carepartner location data (must include ID)"
// @Success 200 {object} prompts.CarepartnerLocationPrompt
// @Failure 400 {string} string "invalid request body or missing ID"
// @Failure 500 {string} string "internal server error"
// @Router /carepartner-location [put]
func (c *CarepartnerLocationController) UpdateCarepartnerLocationHandler(w http.ResponseWriter, r *http.Request) {
	var loc prompts.CarepartnerLocationPrompt

	if err := json.NewDecoder(r.Body).Decode(&loc); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if loc.ID <= 0 {
		http.Error(w, "location ID is required", http.StatusBadRequest)
		return
	}

	if err := c.services.UpdateCarepartnerLocation(&loc); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(loc)
}

// DeleteCarepartnerLocationHandler godoc
// @Summary Delete a carepartner location
// @Description Deletes a carepartner location record by its unique ID
// @Tags carepartner-location
// @Accept json
// @Produce json
// @Param id query int true "Carepartner location ID"
// @Success 204 {string} string "No Content"
// @Failure 400 {string} string "invalid carepartner location ID"
// @Failure 500 {string} string "internal server error"
// @Router /carepartner-location [delete]
func (c *CarepartnerLocationController) DeleteCarepartnerLocationHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil || id <= 0 {
		http.Error(w, "invalid carepartner location ID", http.StatusBadRequest)
		return
	}

	if err := c.services.DeleteCarepartnerLocation(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
