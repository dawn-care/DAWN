package carepartner

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	model "server/internal/models/user/carepartner"
	service "server/internal/services/user/carepartner"

	"github.com/go-chi/chi/v5"
)

type CarepartnerController struct {
	Service *service.CarepartnerService
}

func NewCarepartnerController(service *service.CarepartnerService) *CarepartnerController {
	return &CarepartnerController{Service: service}
}

// ─────────────────────────────────────────────
// Helpers
// ─────────────────────────────────────────────

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

// ─────────────────────────────────────────────
// CREATE
// ─────────────────────────────────────────────

func (cc *CarepartnerController) CreateCarepartner(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	// For now just have a default name
	c := &model.Carepartner{
		Name: "Placeholder Name",
	}

	err := cc.Service.CreateCarepartner(c)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, map[string]int{"carepartner_id": c.ID})
}

// ─────────────────────────────────────────────
// READ
// ─────────────────────────────────────────────

func (cc *CarepartnerController) GetCarepartnerByID(w http.ResponseWriter, r *http.Request) {
	id, err := extractIDFromParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	c, err := cc.Service.GetCarepartnerByID(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if c == nil {
		writeError(w, http.StatusNotFound, "carepartner not found")
		return
	}

	writeJSON(w, http.StatusOK, c)
}

func (cc *CarepartnerController) GetAllCarepartners(w http.ResponseWriter, r *http.Request) {
	list, err := cc.Service.GetAllCarepartners()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, list)
}

// ─────────────────────────────────────────────
// DELETE
// ─────────────────────────────────────────────

func (cc *CarepartnerController) DeleteCarepartner(w http.ResponseWriter, r *http.Request) {
	id, err := extractIDFromParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := cc.Service.DeleteCarepartner(id); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "carepartner deleted successfully"})
}
