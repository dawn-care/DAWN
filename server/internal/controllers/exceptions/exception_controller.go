package exceptions

import (
	"encoding/json"
	"net/http"
	"strconv"

	exceptions "server/internal/models/exceptions"

	"github.com/go-chi/chi/v5"
)

// ---- Interface ----
type ExceptionServiceInterface interface {
	CreateException(exception *exceptions.Exception) error
	GetExceptionsByConfigurationID(configurationID int) ([]exceptions.Exception, error)
	DeleteExceptionsByConfigurationID(configurationID int) error
}

// ---- Controller ----
type ExceptionController struct {
	Service ExceptionServiceInterface
}

func NewExceptionController(service ExceptionServiceInterface) *ExceptionController {
	return &ExceptionController{Service: service}
}

// CreateExceptionHandler creates a new exception for a configuration.
func (c *ExceptionController) CreateExceptionHandler(w http.ResponseWriter, r *http.Request) {
	var exception exceptions.Exception

	if err := json.NewDecoder(r.Body).Decode(&exception); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := c.Service.CreateException(&exception); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "exception created successfully"})
}

// GetExceptionsByConfigurationIDHandler returns all exceptions for a configuration.
func (c *ExceptionController) GetExceptionsByConfigurationIDHandler(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "configuration_id")
	configurationID, err := strconv.Atoi(idStr)
	if err != nil || configurationID <= 0 {
		http.Error(w, "invalid configuration_id", http.StatusBadRequest)
		return
	}

	excs, err := c.Service.GetExceptionsByConfigurationID(configurationID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(excs)
}

// DeleteExceptionsByConfigurationIDHandler removes all exceptions for a configuration.
func (c *ExceptionController) DeleteExceptionsByConfigurationIDHandler(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "configuration_id")
	configurationID, err := strconv.Atoi(idStr)
	if err != nil || configurationID <= 0 {
		http.Error(w, "invalid configuration_id", http.StatusBadRequest)
		return
	}

	if err := c.Service.DeleteExceptionsByConfigurationID(configurationID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "exceptions deleted successfully"})
}
