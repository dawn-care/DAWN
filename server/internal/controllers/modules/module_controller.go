package modules

import (
	"encoding/json"
	"net/http"
	"strconv"

	modules "server/internal/models/modules"

	"github.com/go-chi/chi/v5"
)

// ---- Interface ----
type ModuleServiceInterface interface {
	CreateModule(moduleType string, caregiverID, carepartnerID, metadataID int) (*modules.Module, error)
	GetModuleByID(id int) (*modules.Module, error)
	GetMetadataByCaregiverCarepartnerAndType(caregiverID, carepartnerID int, moduleType string) ([]int, error)
	GetAllMetadataByCaregiverCarepartner(caregiverID, carepartnerID int) ([]int, error)
	GetModulesByCaregiverAndCarepartner(caregiverID, carepartnerID int) ([]*modules.Module, error)
	DeleteModule(id int) error
}

// ---- Controller ----
type ModuleController struct {
	Service ModuleServiceInterface
}

func NewModuleController(service ModuleServiceInterface) *ModuleController {
	return &ModuleController{Service: service}
}

// ---------------------------
// CREATE Module only
// ---------------------------

// CreateModuleHandler godoc
// @Summary Create a new module referencing existing media
// @Description Creates a new module record pointing to an existing media metadata_id
// @Tags module
// @Accept json
// @Produce json
// @Param module body object{module_type=string,caregiver_id=int,carepartner_id=int,metadata_id=int} true "Module creation request"
// @Success 200 {object} modules.Module
// @Failure 400 {string} string "invalid request body"
// @Failure 500 {string} string "internal server error"
// @Router /modules [post]
func (c *ModuleController) CreateModuleHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ModuleType    string `json:"module_type"`
		CaregiverID   int    `json:"caregiver_id"`
		CarepartnerID int    `json:"carepartner_id"`
		MetadataID    int    `json:"metadata_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	mod, err := c.Service.CreateModule(
		req.ModuleType,
		req.CaregiverID,
		req.CarepartnerID,
		req.MetadataID,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(mod)
}

// ---------------------------
// READ
// ---------------------------
func (c *ModuleController) GetModulesByCaregiverAndCarepartnerHandler(w http.ResponseWriter, r *http.Request) {

	caregiverID, err := strconv.Atoi(r.URL.Query().Get("caregiver_id"))
	carepartnerID, err := strconv.Atoi(r.URL.Query().Get("carepartner_id"))

	if err != nil || caregiverID <= 0 || carepartnerID <= 0{
		http.Error(w, "Missing query parameters", http.StatusBadRequest)
		return
	}

	modules, err := c.Service.GetModulesByCaregiverAndCarepartner(caregiverID, carepartnerID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(modules)
}

func (c *ModuleController) GetModuleByIDHandler(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(mod)
}

func (c *ModuleController) GetMetadataHandler(w http.ResponseWriter, r *http.Request) {
	caregiverID, _ := strconv.Atoi(r.URL.Query().Get("caregiver_id"))
	carepartnerID, _ := strconv.Atoi(r.URL.Query().Get("carepartner_id"))
	moduleType := r.URL.Query().Get("module_type")

	metadataIDs, err := c.Service.GetMetadataByCaregiverCarepartnerAndType(caregiverID, carepartnerID, moduleType)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metadataIDs)
}

func (c *ModuleController) GetAllMetadataHandler(w http.ResponseWriter, r *http.Request) {
	caregiverID, _ := strconv.Atoi(r.URL.Query().Get("caregiver_id"))
	carepartnerID, _ := strconv.Atoi(r.URL.Query().Get("carepartner_id"))

	metadataIDs, err := c.Service.GetAllMetadataByCaregiverCarepartner(caregiverID, carepartnerID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metadataIDs)
}

// ---------------------------
// DELETE
// ---------------------------

// DeleteModuleHandler deletes a module by ID
func (c *ModuleController) DeleteModuleHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		http.Error(w, "invalid module ID", http.StatusBadRequest)
		return
	}

	if err := c.Service.DeleteModule(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("module deleted"))
}
