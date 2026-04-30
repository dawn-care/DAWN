package carepartner

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	model "server/internal/models/user/carepartner"

	"github.com/go-chi/chi/v5"
)

type CarepartnerServiceInterface interface {
	CreateCarepartner(c *model.Carepartner) error
	GetCarepartnerByID(id int) (*model.Carepartner, error)
	GetAllCarepartners() ([]model.Carepartner, error)
	UpdateCarepartner(c *model.Carepartner) error
	DeleteCarepartner(id int) error
}

type CarepartnerController struct {
	Service CarepartnerServiceInterface
}

func NewCarepartnerController(service CarepartnerServiceInterface) *CarepartnerController {
	return &CarepartnerController{Service: service}
}

type CarepartnerResponse struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	AutoStart string `json:"auto_start"`
}

type CreateCarepartnerRequest struct {
	Name      string `json:"name"`
	AutoStart string `json:"auto_start"` // "HH:MM:SS"
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

	var req CreateCarepartnerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	// Default name if missing
	if req.Name == "" {
		req.Name = "Placeholder Name"
	}

	// Parse time (HH:MM:SS)
	var autoStart time.Time
	var err error

	if req.AutoStart == "" {
		// default to 08:00:00
		autoStart, _ = time.Parse("15:04:05", "08:00:00")
	} else {
		autoStart, err = time.Parse("15:04:05", req.AutoStart)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid auto_start format (use HH:MM:SS)")
			return
		}
	}

	c := &model.Carepartner{
		Name:      req.Name,
		AutoStart: autoStart,
	}

	err = cc.Service.CreateCarepartner(c)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"carepartner_id": c.ID,
		"auto_start":     c.AutoStart.Format("15:04:05"),
	})
}

func (cc *CarepartnerController) UpdateCarepartner(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var req struct {
		ID        int    `json:"id"`
		Name      string `json:"name"`
		AutoStart string `json:"auto_start"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	if req.ID <= 0 {
		writeError(w, http.StatusBadRequest, "invalid carepartner ID")
		return
	}

	existingCarepartner, err := cc.Service.GetCarepartnerByID(req.ID)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if existingCarepartner == nil {
		writeError(w, http.StatusBadRequest, "carepartner not found")
		return
	}

	// Parse time
	autoStart, err := time.Parse("15:04:05", req.AutoStart)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid auto_start format")
		return
	}

	c := &model.Carepartner{
		ID:          req.ID,
		Name:        existingCarepartner.Name,
		CaregiverID: existingCarepartner.CaregiverID,
		AutoStart:   autoStart,
	}

	if req.Name != "" {
		c.Name = req.Name
	}

	if err := cc.Service.UpdateCarepartner(c); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"message": "carepartner updated successfully",
	})
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

	resp := CarepartnerResponse{
		ID:        c.ID,
		Name:      c.Name,
		AutoStart: c.AutoStart.Format("15:04:05"),
	}

	writeJSON(w, http.StatusOK, resp)
}

func (cc *CarepartnerController) GetAllCarepartners(w http.ResponseWriter, r *http.Request) {
	list, err := cc.Service.GetAllCarepartners()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	var resp []CarepartnerResponse
	for _, c := range list {
		resp = append(resp, CarepartnerResponse{
			ID:        c.ID,
			Name:      c.Name,
			AutoStart: c.AutoStart.Format("15:04:05"),
		})
	}

	writeJSON(w, http.StatusOK, resp)
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
