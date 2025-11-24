package caregiver

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	caregiver "server/internal/models/user/caregiver"

	"github.com/go-chi/chi/v5"
)

// -----------------------------------------------------
// Service Interface (for mocking)
// -----------------------------------------------------

type CaregiverServiceInterface interface {
	CreateCaregiver(cg *caregiver.Caregiver) error
	GetCaregiverByID(id int) (*caregiver.Caregiver, error)
	GetAllCaregivers() ([]caregiver.Caregiver, error)
	GetCaregiverByCarereceiverID(id int) (*caregiver.Caregiver, error)
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

// -----------------------------------------------------
// CREATE
// -----------------------------------------------------

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

// -----------------------------------------------------
// READ
// -----------------------------------------------------

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

func (cc *CaregiverController) GetAllCaregivers(w http.ResponseWriter, r *http.Request) {
	list, err := cc.Service.GetAllCaregivers()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, list)
}

func (cc *CaregiverController) GetCaregiverByCarereceiverID(w http.ResponseWriter, r *http.Request) {
	id, err := extractIDFromParam(r, "carereceiverId")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid carereceiver ID")
		return
	}

	cg, err := cc.Service.GetCaregiverByCarereceiverID(id)
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

// -----------------------------------------------------
// UPDATE
// -----------------------------------------------------

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

// -----------------------------------------------------
// DELETE
// -----------------------------------------------------

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
