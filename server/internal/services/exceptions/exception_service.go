package exceptions

import (
	"database/sql"
	"errors"
	"fmt"

	exceptions "server/internal/models/exceptions"
)

type ExceptionService struct {
	DB *sql.DB
}

func NewExceptionService(db *sql.DB) *ExceptionService {
	return &ExceptionService{DB: db}
}

// CreateException creates a new exception for a configuration.
func (s *ExceptionService) CreateException(exception *exceptions.Exception) error {
	if exception == nil {
		return errors.New("exception cannot be nil")
	}
	if exception.ConfigurationID <= 0 {
		return errors.New("configuration_id is required and must be valid")
	}
	if exception.ExceptionType == "" {
		return errors.New("exception_type is required")
	}
	if len(exception.ExceptionDefinition) == 0 {
		return errors.New("exception_definition is required")
	}

	if err := exceptions.CreateException(s.DB, exception); err != nil {
		return fmt.Errorf("failed to create exception: %w", err)
	}
	return nil
}

// GetExceptionsByConfigurationID returns all exceptions for a configuration.
func (s *ExceptionService) GetExceptionsByConfigurationID(configurationID int) ([]exceptions.Exception, error) {
	if configurationID <= 0 {
		return nil, errors.New("invalid configuration ID")
	}

	excs, err := exceptions.GetExceptionsByConfigurationID(s.DB, configurationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get exceptions: %w", err)
	}
	return excs, nil
}

// DeleteExceptionsByConfigurationID removes all exceptions for a configuration.
func (s *ExceptionService) DeleteExceptionsByConfigurationID(configurationID int) error {
	if configurationID <= 0 {
		return errors.New("invalid configuration ID")
	}

	if err := exceptions.DeleteExceptionsByConfigurationID(s.DB, configurationID); err != nil {
		return fmt.Errorf("failed to delete exceptions: %w", err)
	}
	return nil
}
