package modules

import (
	"encoding/json"
	"net/http"
	"strconv"

	modules "server/internal/models/modules"
)

// ---- Interface ----
type ModuleServiceInterface interface {
	CreateModuleWithMedia(moduleType string, caregiverID, carepartnerID int, uri, filename, extension string) (*modules.Module, error)
	GetModuleByID(id int) (*modules.Module, error)
	GetMetadataByCaregiverCarepartnerAndType(caregiverID, carepartnerID int, moduleType string) ([]int, error)
}

// ---- Controller ----
type ModuleController struct {
	Service ModuleServiceInterface
}

func NewModuleController(service ModuleServiceInterface) *ModuleController {
	return &ModuleController{Service: service}
}

// Create a new module with associated media
func (c *ModuleController) CreateModuleHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ModuleType    string `json:"module_type"`
		CaregiverID   int    `json:"caregiver_id"`
		CarepartnerID int    `json:"carepartner_id"`
		URI           string `json:"uri"`
		Filename      string `json:"filename"`
		Extension     string `json:"extension"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	mod, err := c.Service.CreateModuleWithMedia(
		req.ModuleType,
		req.CaregiverID,
		req.CarepartnerID,
		req.URI,
		req.Filename,
		req.Extension,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(mod)
}

// Get a module by ID
func (c *ModuleController) GetModuleByIDHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		http.Error(w, "invalid module ID", http.StatusBadRequest)
		return
	}

	mod, err := c.Service.GetModuleByID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(mod)
}

// Get metadata IDs by caregiver, carepartner, and module type
func (c *ModuleController) GetMetadataHandler(w http.ResponseWriter, r *http.Request) {
	caregiverID, _ := strconv.Atoi(r.URL.Query().Get("caregiver_id"))
	carepartnerID, _ := strconv.Atoi(r.URL.Query().Get("carepartner_id"))
	moduleType := r.URL.Query().Get("module_type")

	metadataIDs, err := c.Service.GetMetadataByCaregiverCarepartnerAndType(caregiverID, carepartnerID, moduleType)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(metadataIDs)
}
