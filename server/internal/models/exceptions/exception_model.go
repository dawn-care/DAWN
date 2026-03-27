package exceptions

import (
	"database/sql"
	"fmt"

	"github.com/lib/pq"
)

type Exception struct {
	ConfigurationID     int      `json:"configuration_id"`
	ExceptionType       string   `json:"exception_type"`
	ExceptionDefinition []string `json:"exception_definition"`
}

// CreateException inserts a new exception row.
func CreateException(db *sql.DB, exception *Exception) error {
	query := `INSERT INTO "Exception" (configuration_id, exception_type, exception_definition) VALUES ($1, $2, $3);`
	_, err := db.Exec(query, exception.ConfigurationID, exception.ExceptionType, pq.Array(exception.ExceptionDefinition))
	if err != nil {
		return fmt.Errorf("failed to create exception: %w", err)
	}
	return nil
}

// GetExceptionsByConfigurationID returns all exceptions for a given configuration.
func GetExceptionsByConfigurationID(db *sql.DB, configurationID int) ([]Exception, error) {
	query := `SELECT configuration_id, exception_type, exception_definition FROM "Exception" WHERE configuration_id = $1;`
	rows, err := db.Query(query, configurationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get exceptions: %w", err)
	}
	defer rows.Close()

	var exceptions []Exception
	for rows.Next() {
		var e Exception
		if err := rows.Scan(&e.ConfigurationID, &e.ExceptionType, pq.Array(&e.ExceptionDefinition)); err != nil {
			return nil, fmt.Errorf("error scanning exception: %w", err)
		}
		exceptions = append(exceptions, e)
	}
	return exceptions, nil
}

// DeleteExceptionsByConfigurationID removes all exceptions for a configuration.
func DeleteExceptionsByConfigurationID(db *sql.DB, configurationID int) error {
	query := `DELETE FROM "Exception" WHERE configuration_id = $1;`
	_, err := db.Exec(query, configurationID)
	if err != nil {
		return fmt.Errorf("failed to delete exceptions: %w", err)
	}
	return nil
}
