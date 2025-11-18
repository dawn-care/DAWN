package prompts

import (
	"encoding/json"
	"net/http"
	"strconv"

	models "server/internal/models/prompts"
)

// ---- Interface ----
type CarepartnerLocationServiceInterface interface {
	CreateCarepartnerLocation(loc *models.CarepartnerLocationPrompt) error
	GetCarepartnerLocationByID(id int) (*models.CarepartnerLocationPrompt, error)
	UpdateCarepartnerLocation(loc *models.CarepartnerLocationPrompt) error
	DeleteCarepartnerLocation(id int) error
}

// ---- Controller ----
type CarepartnerLocationController struct {
	services CarepartnerLocationServiceInterface
}

func NewCarepartnerLocationController(services CarepartnerLocationServiceInterface) *CarepartnerLocationController {
	return &CarepartnerLocationController{services: services}
}

// Create a new carepartner location
func (c *CarepartnerLocationController) CreateCarepartnerLocationHandler(w http.ResponseWriter, r *http.Request) {
	var loc models.CarepartnerLocationPrompt

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

// Get carepartner location by ID
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

// Update carepartner location
func (c *CarepartnerLocationController) UpdateCarepartnerLocationHandler(w http.ResponseWriter, r *http.Request) {
	var loc models.CarepartnerLocationPrompt

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

// Delete carepartner location
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
