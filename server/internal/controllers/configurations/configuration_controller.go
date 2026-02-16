package configurations

import (
	"encoding/json"
	"net/http"
	"strconv"

	configurations "server/internal/models/configurations"

	"github.com/go-chi/chi/v5"
)

// ---- Interface ----
type ConfigurationServiceInterface interface {
	InsertConfigurations(carepartnerID int, jsonConfig string) error
	DeleteConfigurations(carepartnerID int, idList []int) error
	CreateConfiguration(configuration *configurations.Configuration) (*configurations.Configuration, error)
	GetConfigurationByID(id int) (*configurations.Configuration, error)
	GetConfigurationsByDayAndCarepartner(carepartnerID int, day, month, dayOfWeek int) ([]configurations.ConfigurationWithMetadata, error)
	GetCarePartnerConfigurations(carepartnerID int) ([]configurations.Configuration, error)
	UpdateConfiguration(configuration *configurations.Configuration) (*configurations.Configuration, error)
	DeleteConfiguration(id, carepartnerID int) error
	GetConfigurationStats() (int, error)
}

// ---- Controller ----
type ConfigurationController struct {
	Service ConfigurationServiceInterface
}

func NewConfigurationController(service ConfigurationServiceInterface) *ConfigurationController {
	return &ConfigurationController{Service: service}
}

// CreateConfigurationHandler godoc
// @Summary Create a new configuration
// @Description Creates a new configuration record for a carepartner
// @Tags configuration
// @Accept json
// @Produce json
// @Param configuration body object{carepartner_id=int,module_id=int,occurence_type=string,occurence_definition=[]string,position=int,duration=int} true "Configuration creation request"
// @Success 200 {object} configurations.Configuration
// @Failure 400 {string} string "invalid request body"
// @Failure 500 {string} string "internal server error"
// @Router /configurations [post]
func (c *ConfigurationController) CreateConfigurationHandler(w http.ResponseWriter, r *http.Request) {
	var configuration configurations.Configuration

	if err := json.NewDecoder(r.Body).Decode(&configuration); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	result, err := c.Service.CreateConfiguration(&configuration)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// GetConfigurationByIDHandler godoc
// @Summary Get a configuration by ID
// @Description Retrieves a configuration record by its unique ID
// @Tags configuration
// @Accept json
// @Produce json
// @Param id path int true "Configuration ID"
// @Success 200 {object} configurations.Configuration
// @Failure 400 {string} string "invalid configuration ID"
// @Failure 500 {string} string "internal server error"
// @Router /configurations/{id} [get]
func (c *ConfigurationController) GetConfigurationByIDHandler(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		http.Error(w, "invalid configuration ID", http.StatusBadRequest)
		return
	}

	configuration, err := c.Service.GetConfigurationByID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(configuration)
}

// GetConfigurationsByDayAndCarepartnerHandler godoc
// @Summary Get configurations for a specific day and carepartner
// @Description Retrieves configurations for a carepartner on a specific day
// @Tags configuration
// @Accept json
// @Produce json
// @Param carepartner_id query int true "Carepartner ID"
// @Param day query int true "Day of month (1-31)"
// @Param month query int true "Month (1-12)"
// @Param day_of_week query int true "Day of week (0-6, 0=Sunday)"
// @Success 200 {array} configurations.Configuration
// @Failure 400 {string} string "invalid query parameters"
// @Failure 500 {string} string "internal server error"
// @Router /configurations/day [get]
func (c *ConfigurationController) GetConfigurationsByDayAndCarepartnerHandler(w http.ResponseWriter, r *http.Request) {
	carepartnerID, err := strconv.Atoi(r.URL.Query().Get("carepartner_id"))
	if err != nil || carepartnerID <= 0 {
		http.Error(w, "invalid carepartner_id", http.StatusBadRequest)
		return
	}

	day, err := strconv.Atoi(r.URL.Query().Get("day"))
	if err != nil || day < 1 || day > 31 {
		http.Error(w, "invalid day (must be 1-31)", http.StatusBadRequest)
		return
	}

	month, err := strconv.Atoi(r.URL.Query().Get("month"))
	if err != nil || month < 1 || month > 12 {
		http.Error(w, "invalid month (must be 1-12)", http.StatusBadRequest)
		return
	}

	dayOfWeek, err := strconv.Atoi(r.URL.Query().Get("day_of_week"))
	if err != nil || dayOfWeek < 0 || dayOfWeek > 6 {
		http.Error(w, "invalid day_of_week (must be 0-6)", http.StatusBadRequest)
		return
	}

	configurationList, err := c.Service.GetConfigurationsByDayAndCarepartner(carepartnerID, day, month, dayOfWeek)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(configurationList)
}

// GetCarePartnerConfigurationsHandler godoc
// @Summary Get all configurations for a carepartner
// @Description Retrieves all configuration records for a specific carepartner
// @Tags configuration
// @Accept json
// @Produce json
// @Param carepartner_id query int true "Carepartner ID"
// @Success 200 {array} configurations.Configuration
// @Failure 400 {string} string "invalid carepartner_id"
// @Failure 500 {string} string "internal server error"
// @Router /configurations/carepartner [get]
func (c *ConfigurationController) GetCarePartnerConfigurationsHandler(w http.ResponseWriter, r *http.Request) {
	carepartnerID, err := strconv.Atoi(r.URL.Query().Get("carepartner_id"))
	if err != nil || carepartnerID <= 0 {
		http.Error(w, "invalid carepartner_id", http.StatusBadRequest)
		return
	}

	configurationList, err := c.Service.GetCarePartnerConfigurations(carepartnerID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(configurationList)
}

// UpdateConfigurationHandler godoc
// @Summary Update a configuration
// @Description Updates an existing configuration record
// @Tags configuration
// @Accept json
// @Produce json
// @Param configuration body object{id=int,carepartner_id=int,module_id=int,occurence_type=string,occurence_definition=[]string,position=int,duration=int} true "Configuration update request"
// @Success 200 {object} configurations.Configuration
// @Failure 400 {string} string "invalid request body"
// @Failure 500 {string} string "internal server error"
// @Router /configurations [put]
func (c *ConfigurationController) UpdateConfigurationHandler(w http.ResponseWriter, r *http.Request) {
	var configuration configurations.Configuration

	if err := json.NewDecoder(r.Body).Decode(&configuration); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	result, err := c.Service.UpdateConfiguration(&configuration)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// DeleteConfigurationHandler godoc
// @Summary Delete a configuration
// @Description Deletes a configuration record by its ID and carepartner ID
// @Tags configuration
// @Accept json
// @Produce json
// @Param id path int true "Configuration ID"
// @Param carepartner_id query int true "Carepartner ID"
// @Success 200 {string} string "configuration deleted successfully"
// @Failure 400 {string} string "invalid parameters"
// @Failure 500 {string} string "internal server error"
// @Router /configurations/{id} [delete]
func (c *ConfigurationController) DeleteConfigurationHandler(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		http.Error(w, "invalid configuration ID", http.StatusBadRequest)
		return
	}

	carepartnerID, err := strconv.Atoi(r.URL.Query().Get("carepartner_id"))
	if err != nil || carepartnerID <= 0 {
		http.Error(w, "invalid carepartner_id", http.StatusBadRequest)
		return
	}

	if err := c.Service.DeleteConfiguration(id, carepartnerID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "configuration deleted successfully"})
}

// InsertConfigurationsHandler godoc
// @Summary Insert configurations using stored procedure
// @Description Directly calls the insert_configurations stored procedure with JSON configuration data
// @Tags configuration
// @Accept json
// @Produce json
// @Param request body object{carepartner_id=int,json_config=string} true "Insert configurations request"
// @Success 200 {string} string "configurations inserted successfully"
// @Failure 400 {string} string "invalid request body"
// @Failure 500 {string} string "internal server error"
// @Router /configurations/insert [post]
func (c *ConfigurationController) InsertConfigurationsHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		CarepartnerID int    `json:"carepartner_id"`
		JSONConfig    string `json:"json_config"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := c.Service.InsertConfigurations(req.CarepartnerID, req.JSONConfig); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "configurations inserted successfully"})
}

// DeleteConfigurationsHandler godoc
// @Summary Delete multiple configurations using stored procedure
// @Description Directly calls the delete_configurations stored procedure to delete multiple configurations by ID
// @Tags configuration
// @Accept json
// @Produce json
// @Param request body object{carepartner_id=int,id_list=[]int} true "Delete configurations request"
// @Success 200 {string} string "configurations deleted successfully"
// @Failure 400 {string} string "invalid request body"
// @Failure 500 {string} string "internal server error"
// @Router /configurations/delete [post]
func (c *ConfigurationController) DeleteConfigurationsHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		CarepartnerID int   `json:"carepartner_id"`
		IDList        []int `json:"id_list"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := c.Service.DeleteConfigurations(req.CarepartnerID, req.IDList); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "configurations deleted successfully"})
}

// GetConfigurationStatsHandler godoc
// @Summary Get configuration statistics
// @Description Retrieves the total count of all configurations
// @Tags configuration
// @Accept json
// @Produce json
// @Success 200 {object} object{count=int}
// @Failure 500 {string} string "internal server error"
// @Router /configurations/stats [get]
func (c *ConfigurationController) GetConfigurationStatsHandler(w http.ResponseWriter, r *http.Request) {
	count, err := c.Service.GetConfigurationStats()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]int{"count": count})
}
