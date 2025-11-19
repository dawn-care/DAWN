package caregiver

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	caregiver "server/internal/models/user/caregiver"
	services "server/internal/services/user/caregiver"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type CaregiverController struct {
	Service *services.CaregiverService
}

// Constructor
func NewCaregiverController(service *services.CaregiverService) *CaregiverController {
	return &CaregiverController{Service: service}
}

/* -----------------------------------------------------
   Helpers
----------------------------------------------------- */

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func parseID(param string) (int, error) {
	id, err := strconv.Atoi(param)
	if err != nil || id <= 0 {
		return 0, errors.New("invalid ID")
	}
	return id, nil
}

/* -----------------------------------------------------
   CREATE
----------------------------------------------------- */

func (cc *CaregiverController) CreateCaregiver(w http.ResponseWriter, r *http.Request) {
	var caregiver caregiver.Caregiver
	if err := json.NewDecoder(r.Body).Decode(&caregiver); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if err := cc.Service.CreateCaregiver(&caregiver); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusCreated, caregiver)
}

/* -----------------------------------------------------
   READ
----------------------------------------------------- */

func (cc *CaregiverController) GetCaregiverByID(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(chi.URLParam(r, "id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid ID"})
		return
	}

	caregiver, err := cc.Service.GetCaregiverByID(id)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	if caregiver == nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "caregiver not found"})
		return
	}

	writeJSON(w, http.StatusOK, caregiver)
}

func (cc *CaregiverController) GetAllCaregivers(w http.ResponseWriter, r *http.Request) {
	list, err := cc.Service.GetAllCaregivers()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, list)
}

func (cc *CaregiverController) GetCaregiverByCarereceiverID(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(chi.URLParam(r, "carereceiverId"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid carereceiver ID"})
		return
	}

	cg, err := cc.Service.GetCaregiverByCarereceiverID(id)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	if cg == nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "caregiver not found"})
		return
	}

	writeJSON(w, http.StatusOK, cg)
}

func (cc *CaregiverController) SearchCaregivers(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if len(query) < 2 {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "query must be at least 2 characters",
		})
		return
	}

	results, err := cc.Service.SearchCaregivers(query)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, results)
}

/* -----------------------------------------------------
   UPDATE
----------------------------------------------------- */

func (cc *CaregiverController) UpdateCaregiver(w http.ResponseWriter, r *http.Request) {
	var caregiver caregiver.Caregiver

	if err := json.NewDecoder(r.Body).Decode(&caregiver); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if caregiver.ID <= 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "missing or invalid caregiver ID",
		})
		return
	}

	if err := cc.Service.UpdateCaregiver(&caregiver); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "caregiver updated successfully"})
}

/* -----------------------------------------------------
   DELETE
----------------------------------------------------- */

func (cc *CaregiverController) DeleteCaregiver(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(chi.URLParam(r, "id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid ID"})
		return
	}

	err = cc.Service.DeleteCaregiver(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "caregiver not found"})
			return
		}
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "caregiver deleted successfully"})
}
