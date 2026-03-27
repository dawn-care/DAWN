package caregiver

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"server/internal/models/user/caregiver"

	"github.com/go-chi/chi/v5"
)

// -----------------------------------------------------
// Service Interface (for mocking)
// -----------------------------------------------------

type CaregiverServiceInterface interface {
	CreateCaregiver(cg *caregiver.Caregiver) error
	GetCaregiverByID(id int) (*caregiver.Caregiver, error)
	GetCarepartnersByCaregiverID(id int) ([]caregiver.Caregiver, error)
	GetAllCaregivers() ([]caregiver.Caregiver, error)
	// GetCaregiverByCarepartnerID(id int) (*caregiver.Caregiver, error)
	SearchCaregivers(query string) ([]caregiver.Caregiver, error)
	UpdateCaregiver(cg *caregiver.Caregiver) error
	DeleteCaregiver(id int) error
}

// -----------------------------------------------------
// Controller
// -----------------------------------------------------

type CaregiverController struct {
	Service CaregiverServiceInterface
}

func NewCaregiverController(service CaregiverServiceInterface) *CaregiverController {
	return &CaregiverController{Service: service}
}

// -----------------------------------------------------
// Helpers
// -----------------------------------------------------

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func extractIDFromParam(r *http.Request, name string) (int, error) {
	raw := chi.URLParam(r, name)
	id, err := strconv.Atoi(raw)
	if err != nil || id <= 0 {
		return 0, errors.New("invalid ID")
	}
	return id, nil
}

// CreateCaregiver godoc
// @Summary Create a new caregiver
// @Description Creates a new caregiver record
// @Tags caregiver
// @Accept json
// @Produce json
// @Param caregiver body caregiver.Caregiver true "Caregiver to create"
// @Success 201 {object} caregiver.Caregiver
// @Failure 400 {object} map[string]string "invalid request body or validation error"
// @Failure 500 {object} map[string]string "internal server error"
// @Router /caregivers [post]
func (cc *CaregiverController) CreateCaregiver(w http.ResponseWriter, r *http.Request) {
	var cg caregiver.Caregiver

	if err := json.NewDecoder(r.Body).Decode(&cg); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := cc.Service.CreateCaregiver(&cg); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, cg)
}

// GetCaregiverByID godoc
// @Summary Get a caregiver by ID
// @Description Retrieves a single caregiver by ID
// @Tags caregiver
// @Accept json
// @Produce json
// @Param id path int true "Caregiver ID"
// @Success 200 {object} caregiver.Caregiver
// @Failure 400 {object} map[string]string "invalid ID"
// @Failure 404 {object} map[string]string "caregiver not found"
// @Failure 500 {object} map[string]string "internal server error"
// @Router /caregivers/{id} [get]
func (cc *CaregiverController) GetCaregiverByID(w http.ResponseWriter, r *http.Request) {
	id, err := extractIDFromParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid ID")
		return
	}

	cg, err := cc.Service.GetCaregiverByID(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if cg == nil {
		writeError(w, http.StatusNotFound, "caregiver not found")
		return
	}

	writeJSON(w, http.StatusOK, cg)
}

// GetCarepartnersByCaregiverID godoc
// @Summary Get carepartners for a caregiver
// @Description Retrieves all carepartners associated with a caregiver
// @Tags caregiver
// @Accept json
// @Produce json
// @Param caregiver_id query int true "Caregiver ID"
// @Success 200 {array} map[string]interface{}
// @Failure 400 {object} map[string]string "invalid caregiver_id"
// @Failure 500 {object} map[string]string "internal server error"
// @Router /caregivers/carepartners [get]
func (cc *CaregiverController) GetCarepartnersByCaregiverID(w http.ResponseWriter, r *http.Request) {
	caregiverIDStr := chi.URLParam(r, "id")
	if caregiverIDStr == "" {
		writeError(w, http.StatusBadRequest, "missing caregiver id")
		return
	}

	caregiverID, err := strconv.Atoi(caregiverIDStr)
	if err != nil || caregiverID <= 0 {
		writeError(w, http.StatusBadRequest, "invalid caregiver id")
		return
	}

	carepartners, err := cc.Service.GetCarepartnersByCaregiverID(caregiverID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, carepartners)
}

// GetAllCaregivers godoc
// @Summary Get all caregivers
// @Description Retrieves a list of all caregivers
// @Tags caregiver
// @Accept json
// @Produce json
// @Success 200 {array} caregiver.Caregiver
// @Failure 500 {object} map[string]string "internal server error"
// @Router /caregivers [get]
func (cc *CaregiverController) GetAllCaregivers(w http.ResponseWriter, r *http.Request) {
	list, err := cc.Service.GetAllCaregivers()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, list)
}

// GetCaregiverByCarepartnerID godoc
// @Summary Get a caregiver by Carepartner ID
// @Description Retrieves a caregiver associated with a specific Carepartner
// @Tags caregiver
// @Accept json
// @Produce json
// @Param CarepartnerId path int true "Carepartner ID"
// @Success 200 {object} caregiver.Caregiver
// @Failure 400 {object} map[string]string "invalid Carepartner ID"
// @Failure 404 {object} map[string]string "caregiver not found"
// @Failure 500 {object} map[string]string "internal server error"
// @Router /caregivers/carepartner/{CarepartnerId} [get]
// func (cc *CaregiverController) GetCaregiverByCarepartnerID(w http.ResponseWriter, r *http.Request) {
// 	id, err := extractIDFromParam(r, "CarepartnerId")
// 	if err != nil {
// 		writeError(w, http.StatusBadRequest, "invalid Carepartner ID")
// 		return
// 	}

// 	cg, err := cc.Service.GetCaregiverByCarepartnerID(id)
// 	if err != nil {
// 		writeError(w, http.StatusInternalServerError, err.Error())
// 		return
// 	}
// 	if cg == nil {
// 		writeError(w, http.StatusNotFound, "caregiver not found")
// 		return
// 	}

// 	writeJSON(w, http.StatusOK, cg)
// }

// SearchCaregivers godoc
// @Summary Search caregivers
// @Description Searches caregivers by query string (minimum 2 characters)
// @Tags caregiver
// @Accept json
// @Produce json
// @Param q query string true "Search query"
// @Success 200 {array} caregiver.Caregiver
// @Failure 400 {object} map[string]string "query too short or invalid"
// @Failure 500 {object} map[string]string "internal server error"
// @Router /caregivers/search [get]
func (cc *CaregiverController) SearchCaregivers(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if len(query) < 2 {
		writeError(w, http.StatusBadRequest, "query must be at least 2 characters")
		return
	}

	results, err := cc.Service.SearchCaregivers(query)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, results)
}

// UpdateCaregiver godoc
// @Summary Update a caregiver
// @Description Updates an existing caregiver by ID
// @Tags caregiver
// @Accept json
// @Produce json
// @Param caregiver body caregiver.Caregiver true "Caregiver to update (must include ID)"
// @Success 200 {object} map[string]string "caregiver updated successfully"
// @Failure 400 {object} map[string]string "invalid request body or missing ID"
// @Failure 500 {object} map[string]string "internal server error"
// @Router /caregivers [put]
func (cc *CaregiverController) UpdateCaregiver(w http.ResponseWriter, r *http.Request) {
	var cg caregiver.Caregiver

	if err := json.NewDecoder(r.Body).Decode(&cg); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if cg.ID <= 0 {
		writeError(w, http.StatusBadRequest, "missing or invalid caregiver ID")
		return
	}

	if err := cc.Service.UpdateCaregiver(&cg); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "caregiver updated successfully"})
}

// DeleteCaregiver godoc
// @Summary Delete a caregiver
// @Description Deletes a caregiver by ID
// @Tags caregiver
// @Accept json
// @Produce json
// @Param id path int true "Caregiver ID"
// @Success 200 {object} map[string]string "caregiver deleted successfully"
// @Failure 400 {object} map[string]string "invalid ID or deletion error"
// @Failure 404 {object} map[string]string "caregiver not found"
// @Failure 500 {object} map[string]string "internal server error"
// @Router /caregivers/{id} [delete]
func (cc *CaregiverController) DeleteCaregiver(w http.ResponseWriter, r *http.Request) {
	id, err := extractIDFromParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid ID")
		return
	}

	err = cc.Service.DeleteCaregiver(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "caregiver not found")
			return
		}
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "caregiver deleted successfully"})
}
