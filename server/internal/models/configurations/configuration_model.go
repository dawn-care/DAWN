package configurations

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/lib/pq"
)

type Configuration struct {
	ID                   int      `json:"id"`
	CAREPARTNER_ID       int      `json:"carepartner_id"`
	MODULE_ID            int      `json:"module_id"`
	OCCURENCE_TYPE       string   `json:"occurence_type"`
	OCCURENCE_DEFINITION []string `json:"occurence_definition"`
	POSITION             int      `json:"position"`
	DURATION             int      `json:"duration"`
}

// ConfigurationWithMetadata includes configuration details joined with module and metadata info
type ConfigurationWithMetadata struct {
	ID                   int             `json:"id"`
	CAREGIVER_ID         int             `json:"caregiver_id"`
	CAREPARTNER_ID       int             `json:"carepartner_id"`
	MODULE_ID            int             `json:"module_id"`
	OCCURENCE_TYPE       string          `json:"occurence_type"`
	OCCURENCE_DEFINITION []string        `json:"occurence_definition"`
	POSITION             int             `json:"position"`
	DURATION             int             `json:"duration"`
	MODULE_TYPE          string          `json:"module_type"`
	METADATA_ID          int             `json:"metadata_id"`
	Metadata             json.RawMessage `json:"metadata"`
}

var DAYS_OF_WEEK_STR = [7]string{"Sun", "Mon", "Tues", "Wed", "Thur", "Fri", "Sat"}
var MONTH_STR = [12]string{"Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"}

// configJSON is the JSON shape expected by the insert_configurations stored procedure.
type configJSON struct {
	ID                  *int     `json:"id"`
	ModuleID            int      `json:"module_id"`
	OccurenceType       string   `json:"occurence_type"`
	OccurenceDefinition []string `json:"occurence_definition"`
	Position            int      `json:"position"`
	Duration            int      `json:"duration"`
}

// configsToJSON safely marshals configurations into the JSON format expected by insert_configurations.
func configsToJSON(configs ...Configuration) (string, error) {
	entries := make([]configJSON, len(configs))
	for i, c := range configs {
		entries[i] = configJSON{
			ID:                  nil,
			ModuleID:            c.MODULE_ID,
			OccurenceType:       c.OCCURENCE_TYPE,
			OccurenceDefinition: c.OCCURENCE_DEFINITION,
			Position:            c.POSITION,
			Duration:            c.DURATION,
		}
	}
	data, err := json.Marshal(entries)
	if err != nil {
		return "", fmt.Errorf("failed to marshal configuration JSON: %w", err)
	}
	return string(data), nil
}

// InsertConfigurations calls the insert_configurations stored procedure
func InsertConfigurations(db *sql.DB, carepartnerID int, jsonConfig string) error {
	query := `CALL insert_configurations($1, $2::jsonb);`
	_, err := db.Exec(query, carepartnerID, jsonConfig)
	if err != nil {
		return fmt.Errorf("failed to call insert_configurations: %w", err)
	}
	return nil
}

// DeleteConfigurations calls the delete_configurations stored procedure
func DeleteConfigurations(db *sql.DB, carepartnerID int, idList []int) error {
	query := `CALL delete_configurations($1, $2);`
	_, err := db.Exec(query, carepartnerID, pq.Array(idList))
	if err != nil {
		return fmt.Errorf("failed to call delete_configurations: %w", err)
	}
	return nil
}

// CreateConfiguration creates a single configuration via the stored procedure.
func CreateConfiguration(db *sql.DB, configuration *Configuration) error {
	jsonConfig, err := configsToJSON(*configuration)
	if err != nil {
		return err
	}

	if err := InsertConfigurations(db, configuration.CAREPARTNER_ID, jsonConfig); err != nil {
		return err
	}

	err = db.QueryRow(`
			SELECT id FROM "Configuration"
			WHERE carepartner_id = $1 AND position = $2
			ORDER BY id DESC
			LIMIT 1;
		`, configuration.CAREPARTNER_ID, configuration.POSITION).Scan(&configuration.ID)
	return err
}

// Get Configuration by ID
func GetConfigurationByID(db *sql.DB, id int) (*Configuration, error) {
	configuration := &Configuration{}
	query := `
			SELECT id,
					carepartner_id,
					module_id,
					occurence_type,
					occurence_definition,
					position,
					duration
			FROM "Configuration"
			WHERE id = $1;
	`
	err := db.QueryRow(query, id).Scan(&configuration.ID,
		&configuration.CAREPARTNER_ID,
		&configuration.MODULE_ID,
		&configuration.OCCURENCE_TYPE,
		pq.Array(&configuration.OCCURENCE_DEFINITION),
		&configuration.POSITION,
		&configuration.DURATION)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return configuration, nil
}

func GetConfigurationsByDayAndCarepartner(db *sql.DB, carepartner_id int, day, month, day_of_week int) ([]ConfigurationWithMetadata, error) {
	dayStr := fmt.Sprintf("%d", day)                          //For Monthly Type - int 1 - 31
	dayOfWeekStr := DAYS_OF_WEEK_STR[day_of_week]             //For Weekly Type - string "Sun" - "Sat"
	monthDay := fmt.Sprintf("%s %d", MONTH_STR[month-1], day) //For Yearly Type - string "Month day" - i.e. "Jan 1"
	var configurations []ConfigurationWithMetadata

	query := `
			SELECT
				c.id, m.caregiver_id, c.carepartner_id, c.module_id, c.occurence_type, c.occurence_definition, c.position, c.duration,
				m.module_type, m.metadata_id,
				CASE m.module_type
					WHEN 'photo' THEN (SELECT row_to_json(t) FROM (SELECT id, uri, displayed_name, filename, extension FROM "PhotoMedia" WHERE id = m.metadata_id) t)
					WHEN 'video' THEN (SELECT row_to_json(t) FROM (SELECT id, uri, displayed_name, filename, extension FROM "VideoMedia" WHERE id = m.metadata_id) t)
					WHEN 'audio' THEN (SELECT row_to_json(t) FROM (SELECT id, uri, displayed_name, filename, extension FROM "AudioMedia" WHERE id = m.metadata_id) t)
					WHEN 'weather' THEN (SELECT row_to_json(t) FROM (SELECT id, zipcode FROM "WeatherWidget" WHERE id = m.metadata_id) t)
					WHEN 'news' THEN (SELECT row_to_json(t) FROM (SELECT id, category, language, country FROM "NewsWidget" WHERE id = m.metadata_id) t)
					WHEN 'sports' THEN (SELECT row_to_json(t) FROM (SELECT id, category, language, country FROM "NewsWidget" WHERE id = m.metadata_id) t)
					WHEN 'message' THEN (SELECT row_to_json(t) FROM (SELECT id, message, image_uri FROM "MessagePrompt" WHERE id = m.metadata_id) t)
					WHEN 'event' THEN (SELECT row_to_json(t) FROM (SELECT id, list_id, time, text FROM "EventPrompt" WHERE id = m.metadata_id) t)
					WHEN 'carepartner_location' THEN (SELECT row_to_json(t) FROM (SELECT id, location, message, image_uri FROM "CarepartnerLocationPrompt" WHERE id = m.metadata_id) t)
					WHEN 'caregiver_location' THEN (SELECT row_to_json(t) FROM (SELECT id, location, message, image_uri FROM "CaregiverLocationPrompt" WHERE id = m.metadata_id) t)
					ELSE NULL
				END AS metadata
			FROM "Configuration" c
			JOIN "Module" m ON c.module_id = m.id
			WHERE c.carepartner_id = $1
			AND (
				c.occurence_type = 'daily'
				OR (c.occurence_type = 'weekly' AND $2 = ANY(c.occurence_definition))
				OR (c.occurence_type = 'monthly' AND $3 = ANY(c.occurence_definition))
				OR (c.occurence_type = 'yearly' AND $4 = ANY(c.occurence_definition))
			)
			ORDER BY c.position;
	`
	rows, err := db.Query(query, carepartner_id, dayOfWeekStr, dayStr, monthDay)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var cfg ConfigurationWithMetadata
		var metadata []byte
		err := rows.Scan(
			&cfg.ID, &cfg.CAREGIVER_ID, &cfg.CAREPARTNER_ID, &cfg.MODULE_ID,
			&cfg.OCCURENCE_TYPE, pq.Array(&cfg.OCCURENCE_DEFINITION),
			&cfg.POSITION, &cfg.DURATION,
			&cfg.MODULE_TYPE, &cfg.METADATA_ID,
			&metadata,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning: %v", err)
		}
		if metadata != nil {
			cfg.Metadata = json.RawMessage(metadata)
		}
		configurations = append(configurations, cfg)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error in rows: %v", err)
	}

	return configurations, nil
}

// Get All Configurations by Carepartner
func GetCarePartnerConfigurations(db *sql.DB, carepartner_id int) ([]Configuration, error) {
	var configurations []Configuration
	query := `
			SELECT id,
					carepartner_id,
					module_id,
					occurence_type,
					occurence_definition,
					position,
					duration
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
			&configuration.OCCURENCE_TYPE,
			pq.Array(&configuration.OCCURENCE_DEFINITION),
			&configuration.POSITION,
			&configuration.DURATION)
		if err != nil {
			return nil, fmt.Errorf("error scanning: %v", err)
		}
		configurations = append(configurations, configuration)
	}
	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("rows error: %v", err)
	}

	return configurations, nil
}

// UpdateConfiguration deletes the old row and re-inserts with new values.
func UpdateConfiguration(db *sql.DB, configuration *Configuration) error {
	idList := []int{configuration.ID}
	if err := DeleteConfigurations(db, configuration.CAREPARTNER_ID, idList); err != nil {
		return err
	}

	jsonConfig, err := configsToJSON(*configuration)
	if err != nil {
		return err
	}

	if err := InsertConfigurations(db, configuration.CAREPARTNER_ID, jsonConfig); err != nil {
		return err
	}

	err = db.QueryRow(`
			SELECT id FROM "Configuration"
			WHERE carepartner_id = $1 AND position = $2
			ORDER BY id DESC
			LIMIT 1;
		`, configuration.CAREPARTNER_ID, configuration.POSITION).Scan(&configuration.ID)
	return err
}

// Delete A Configuration By ID and carepartner_id
func DeleteConfiguration(db *sql.DB, id, carepartner_id int) error {
	idList := []int{id}
	return DeleteConfigurations(db, carepartner_id, idList)
}

// Get Configuration Stats (i.e., total count)
func GetConfigurationStats(db *sql.DB) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM "Configuration";`
	err := db.QueryRow(query).Scan(&count)
	return count, err
}
