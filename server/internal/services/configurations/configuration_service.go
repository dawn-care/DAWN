package configurations

import (
	"database/sql"
	"errors"
	"fmt"

	configurations "server/internal/models/configurations"
)

type ConfigurationService struct {
	DB *sql.DB
}

func NewConfigurationService(db *sql.DB) *ConfigurationService {
	return &ConfigurationService{DB: db}
}

// InsertConfigurations directly calls the insert_configurations stored procedure
func (s *ConfigurationService) InsertConfigurations(carepartnerID int, jsonConfig string) error {
	if carepartnerID <= 0 {
		return errors.New("invalid carepartner ID")
	}
	if jsonConfig == "" {
		return errors.New("json configuration cannot be empty")
	}

	if err := configurations.InsertConfigurations(s.DB, carepartnerID, jsonConfig); err != nil {
		return fmt.Errorf("failed to insert configurations: %w", err)
	}

	return nil
}

// DeleteConfigurations directly calls the delete_configurations stored procedure
func (s *ConfigurationService) DeleteConfigurations(carepartnerID int, idList []int) error {
	if carepartnerID <= 0 {
		return errors.New("invalid carepartner ID")
	}
	if len(idList) == 0 {
		return errors.New("id list cannot be empty")
	}

	if err := configurations.DeleteConfigurations(s.DB, carepartnerID, idList); err != nil {
		return fmt.Errorf("failed to delete configurations: %w", err)
	}

	return nil
}

// CreateConfiguration creates a single configuration
func (s *ConfigurationService) CreateConfiguration(configuration *configurations.Configuration) (*configurations.Configuration, error) {
	if configuration == nil {
		return nil, errors.New("configuration cannot be nil")
	}
	if configuration.CAREPARTNER_ID <= 0 {
		return nil, errors.New("carepartner_id is required and must be valid")
	}
	if configuration.MODULE_ID <= 0 {
		return nil, errors.New("module_id is required and must be valid")
	}
	if configuration.OCCURENCE_TYPE == "" {
		return nil, errors.New("occurence_type is required")
	}
	if len(configuration.OCCURENCE_DEFINITION) == 0 {
		return nil, errors.New("occurence_definition is required")
	}
	if configuration.POSITION < 0 {
		return nil, errors.New("position must be a non-negative integer")
	}
	if configuration.DURATION < 0 {
		return nil, errors.New("duration must be a non-negative integer")
	}

	if err := configurations.CreateConfiguration(s.DB, configuration); err != nil {
		return nil, fmt.Errorf("failed to create configuration: %w", err)
	}

	return configuration, nil
}

// GetConfigurationByID retrieves a configuration by its ID
func (s *ConfigurationService) GetConfigurationByID(id int) (*configurations.Configuration, error) {
	if id <= 0 {
		return nil, errors.New("invalid configuration ID")
	}

	configuration, err := configurations.GetConfigurationByID(s.DB, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get configuration: %w", err)
	}
	if configuration == nil {
		return nil, fmt.Errorf("configuration with id %d not found", id)
	}

	return configuration, nil
}

// GetConfigurationsByDayAndCarepartner retrieves configurations for a specific day and carepartner, joined with module and metadata
func (s *ConfigurationService) GetConfigurationsByDayAndCarepartner(carepartnerID int, day, month, dayOfWeek int, fullDate string) ([]configurations.ConfigurationWithMetadata, error) {
	if carepartnerID <= 0 {
		return nil, errors.New("invalid carepartner ID")
	}
	if day < 1 || day > 31 {
		return nil, errors.New("day must be between 1 and 31")
	}
	if month < 1 || month > 12 {
		return nil, errors.New("month must be between 1 and 12")
	}
	if dayOfWeek < 0 || dayOfWeek > 6 {
		return nil, errors.New("day of week must be between 0 and 6")
	}

	cfgs, err := configurations.GetConfigurationsByDayAndCarepartner(s.DB, carepartnerID, day, month, dayOfWeek, fullDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get configurations: %w", err)
	}

	return cfgs, nil
}

// GetCarePartnerConfigurations retrieves all configurations for a carepartner
func (s *ConfigurationService) GetCarePartnerConfigurations(carepartnerID int) ([]configurations.Configuration, error) {
	if carepartnerID <= 0 {
		return nil, errors.New("invalid carepartner ID")
	}

	cfgs, err := configurations.GetCarePartnerConfigurations(s.DB, carepartnerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get configurations: %w", err)
	}

	return cfgs, nil
}

// UpdateConfiguration updates an existing configuration
func (s *ConfigurationService) UpdateConfiguration(configuration *configurations.Configuration) (*configurations.Configuration, error) {
	if configuration == nil {
		return nil, errors.New("configuration cannot be nil")
	}
	if configuration.ID <= 0 {
		return nil, errors.New("configuration ID is required and must be valid")
	}
	if configuration.CAREPARTNER_ID <= 0 {
		return nil, errors.New("carepartner_id is required and must be valid")
	}
	if configuration.MODULE_ID <= 0 {
		return nil, errors.New("module_id is required and must be valid")
	}
	if configuration.OCCURENCE_TYPE == "" {
		return nil, errors.New("occurence_type is required")
	}
	if len(configuration.OCCURENCE_DEFINITION) == 0 {
		return nil, errors.New("occurence_definition is required")
	}

	if err := configurations.UpdateConfiguration(s.DB, configuration); err != nil {
		return nil, fmt.Errorf("failed to update configuration: %w", err)
	}

	return configuration, nil
}

// DeleteConfiguration deletes a configuration by ID and carepartner ID
func (s *ConfigurationService) DeleteConfiguration(id, carepartnerID int) error {
	if id <= 0 {
		return errors.New("invalid configuration ID")
	}
	if carepartnerID <= 0 {
		return errors.New("invalid carepartner ID")
	}

	if err := configurations.DeleteConfiguration(s.DB, id, carepartnerID); err != nil {
		return fmt.Errorf("failed to delete configuration: %w", err)
	}

	return nil
}

// GetConfigurationStats retrieves the total count of configurations
func (s *ConfigurationService) GetConfigurationStats() (int, error) {
	count, err := configurations.GetConfigurationStats(s.DB)
	if err != nil {
		return 0, fmt.Errorf("failed to get configuration stats: %w", err)
	}

	return count, nil
}
