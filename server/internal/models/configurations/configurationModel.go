package configurations

import (
	"database/sql"
	"fmt"
)

type Configuration struct {
	ID                    int      `json:"id"`
	CAREPARTNER_ID        int      `json:"carepartner_id"`
	MODULE_ID             int      `json:"module_id"`
	OCCURRENCE_TYPE       string   `json:"occurrence_type"`
	OCCURRENCE_DEFINITION []string `json:"occurrence_definition"`
	POSITION              int      `json:"position"`
}

// Adds Configuration
func CreateConfiguration(db *sql.DB, configuration *Configuration) error {
	query := `
			INSERT INTO "Configuration" (
					carepartner_id,
					module_id,
					occurence_type,
					occurrence_definition,
					position)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING id;
	`
	return db.QueryRow(query,
		configuration.CAREPARTNER_ID,
		configuration.MODULE_ID,
		configuration.OCCURRENCE_TYPE,
		configuration.OCCURRENCE_DEFINITION,
		configuration.POSITION).Scan(configuration.ID)
}

// Get Configuration by ID
func GetConfigurationByID(db *sql.DB, id int) (*Configuration, error) {
	configuration := &Configuration{}
	query := `
			SELECT id,
					carepartner_id,
					module_id,
					occurence_type,
					occurrence_definition,
					position
			FROM "Configuration"
			WHERE id = $1;
	`
	err := db.QueryRow(query, id).Scan(&configuration.ID,
		&configuration.CAREPARTNER_ID,
		&configuration.MODULE_ID,
		&configuration.OCCURRENCE_TYPE,
		&configuration.OCCURRENCE_DEFINITION,
		&configuration.POSITION)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return configuration, nil
}

// Get All Configurations by User
func GetUserConfigurations(db *sql.DB, carepartner_id int) ([]Configuration, error) {
	var configurations []Configuration
	query := `
			SELECT id,
					carepartner_id,
					module_id,
					occurence_type,
					occurrence_definition,
					position
			FROM "Configuration"
			WHERE carepartner_id = $1;
	`
	rows, err := db.Query(query, carepartner_id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var configuration Configuration
	for rows.Next() {
		err := rows.Scan(&configuration.ID,
			&configuration.CAREPARTNER_ID,
			&configuration.MODULE_ID,
			&configuration.OCCURRENCE_TYPE,
			&configuration.OCCURRENCE_DEFINITION,
			&configuration.POSITION)
		if err != nil {
			return nil, fmt.Errorf("Error Scanning: %v", err)
		}
		configurations = append(configurations, configuration)
	}
	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("Rows Error: %v", err)
	}

	return configurations, nil
}

// Update Configuration
func UpdateConfiguration(db *sql.DB, configuration *Configuration) error {
	query := `
			UPDATE "Configuration"
			SET carepartner_id = $1,
					module_id = $2,
					occurrence_type = $3,
					occurrence_definition = $4,
					position = $5
			WHERE id = $6;
	`
	result, err := db.Exec(query,
		configuration.CAREPARTNER_ID,
		configuration.MODULE_ID,
		configuration.OCCURRENCE_TYPE,
		configuration.OCCURRENCE_DEFINITION,
		configuration.POSITION)
	if err != nil {
		return err
	}
	count, err := result.RowsAffected()
	if count == 0 {
		return fmt.Errorf("No Configurations Found With ID: %d", configuration.ID)
	}
	return nil
}

// Delete A Configuration
func DeleteConfiguration(db *sql.DB, id int) error {
	query := `
			DELETE FROM "Configuration"
			WHERE id = $1;
	`
	result, err := db.Exec(query, id)
	if err != nil {
		return err
	}
	count, err := result.RowsAffected()
	if count == 0 {
		return fmt.Errorf("No Configuration Found With ID: %d", id)
	}
	return nil
}

// Get Configuration Stats (i.e., total count)
func GetConfigurationStats(db *sql.DB) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM "Configuration";`
	err := db.QueryRow(query).Scan(&count)
	return count, err
}
