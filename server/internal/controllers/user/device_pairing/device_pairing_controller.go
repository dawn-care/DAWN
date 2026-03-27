package device_pairing

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"server/internal/models/user/device_pairing"

	"github.com/go-chi/chi/v5"
)

type DevicePairingServiceInterface interface {
	CreatePairing(p *device_pairing.DevicePairing) error
	GetPairingByID(id int) (*device_pairing.DevicePairing, error)
	GetPairingByCode(code string) (*device_pairing.DevicePairing, error)
	GetAllPairings() ([]device_pairing.DevicePairing, error)
	UpdatePairing(p *device_pairing.DevicePairing) error
	DeletePairing(id int) error
	GeneratePairingCode(carepartnerID int) (*device_pairing.DevicePairing, error)
	VerifyPairingCode(code string, caregiverID int, carepartnerName string) (int, error)
	IsDevicePaired(carepartnerID int) (bool, error)
}

type DevicePairingController struct {
	Service DevicePairingServiceInterface
}

func NewDevicePairingController(service DevicePairingServiceInterface) *DevicePairingController {
	return &DevicePairingController{Service: service}
}

//
// Helpers
//

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

//
// CREATE
//

func (cc *DevicePairingController) CreatePairing(w http.ResponseWriter, r *http.Request) {

	defer r.Body.Close()

	var p device_pairing.DevicePairing

	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := cc.Service.CreatePairing(&p); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, p)
}

//
// READ
//

func (cc *DevicePairingController) GetPairingByID(w http.ResponseWriter, r *http.Request) {

	id, err := extractIDFromParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid ID")
		return
	}

	p, err := cc.Service.GetPairingByID(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if p == nil {
		writeError(w, http.StatusNotFound, "pairing not found")
		return
	}

	writeJSON(w, http.StatusOK, p)
}

func (cc *DevicePairingController) GetPairingByCode(w http.ResponseWriter, r *http.Request) {

	code := chi.URLParam(r, "code")

	if code == "" {
		writeError(w, http.StatusBadRequest, "invalid pairing code")
		return
	}

	p, err := cc.Service.GetPairingByCode(code)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if p == nil {
		writeError(w, http.StatusNotFound, "pairing not found")
		return
	}

	writeJSON(w, http.StatusOK, p)
}

func (cc *DevicePairingController) GetAllPairings(w http.ResponseWriter, r *http.Request) {

	list, err := cc.Service.GetAllPairings()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, list)
}

//
// DELETE
//

func (cc *DevicePairingController) DeletePairing(w http.ResponseWriter, r *http.Request) {

	id, err := extractIDFromParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid ID")
		return
	}

	err = cc.Service.DeletePairing(id)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"message": "pairing deleted successfully",
	})
}

//
// Generate pairing code
//

func (cc *DevicePairingController) GeneratePairingCode(w http.ResponseWriter, r *http.Request) {

	defer r.Body.Close()

	type request struct {
		CarepartnerID int `json:"carepartner_id"`
	}

	var req request

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	pairing, err := cc.Service.GeneratePairingCode(req.CarepartnerID)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	expiresIn := int(time.Until(pairing.ExpiresAt).Seconds())

	response := map[string]interface{}{
		"ok":             true,
		"pairing_code":   pairing.PairingCode,
		"expires_in_sec": expiresIn,
	}

	writeJSON(w, http.StatusOK, response)
}

//
// Verify pairing code
//

func (cc *DevicePairingController) VerifyPairingCode(w http.ResponseWriter, r *http.Request) {

	defer r.Body.Close()

	type request struct {
		Code            string `json:"code"`
		CaregiverID     int    `json:"caregiver_id"`
		CarepartnerName string `json:"carepartner_name"`
	}

	var req request

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	carepartnerID, err := cc.Service.VerifyPairingCode(req.Code, req.CaregiverID, req.CarepartnerName)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":             true,
		"carepartner_id": carepartnerID,
	})
}

//
// Check if a device is paired
//

func (cc *DevicePairingController) IsDevicePaired(w http.ResponseWriter, r *http.Request) {

	defer r.Body.Close()

	type request struct {
		CarepartnerID int `json:"carepartner_id"`
	}

	var req request

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	paired, err := cc.Service.IsDevicePaired(req.CarepartnerID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := map[string]interface{}{
		"ok":     true,
		"paired": paired,
	}

	writeJSON(w, http.StatusOK, response)
}
