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

// CreateModuleHandler godoc
// @Summary Create a new module with associated media
// @Description Creates a new module record and associates media (audio, video, photo) with it
// @Tags module
// @Accept json
// @Produce json
// @Param module body object{module_type=string,caregiver_id=int,carepartner_id=int,uri=string,filename=string,extension=string} true "Module creation request"
// @Success 200 {object} modules.Module
// @Failure 400 {string} string "invalid request body"
// @Failure 500 {string} string "internal server error"
// @Router /module [post]
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

// GetModuleByIDHandler godoc
// @Summary Get a module by ID
// @Description Retrieves a module record by its unique ID
// @Tags module
// @Accept json
// @Produce json
// @Param id path int true "Module ID"
// @Success 200 {object} modules.Module
// @Failure 400 {string} string "invalid module ID"
// @Failure 500 {string} string "internal server error"
// @Router /module/{id} [get]
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

	json.NewEncoder(w).Encode(mod)
}

// GetMetadataHandler godoc
// @Summary Get metadata IDs for modules
// @Description Retrieves metadata IDs for modules filtered by caregiver, carepartner, and module type
// @Tags module
// @Accept json
// @Produce json
// @Param caregiver_id query int true "Caregiver ID"
// @Param carepartner_id query int true "Carepartner ID"
// @Param module_type query string true "Module type"
// @Success 200 {array} int "Array of metadata IDs"
// @Failure 500 {string} string "internal server error"
// @Router /module/metadata [get]
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
