package configurations

import (
	"database/sql"
	"fmt"
	"strings"

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

var DAYS_OF_WEEK_STR = [7]string{"Sun", "Mon", "Tues", "Wed", "Thur", "Fri", "Sat"}
var MONTH_STR = [12]string{"Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"}

// Adds Configuration
func CreateConfiguration(db *sql.DB, configuration *Configuration) error {

	jsonConfig := fmt.Sprintf(`[
			{
				"id": null,
				"module_id": %d,
				"occurence_type": "%s",
				"occurence_definition": ["%s"],
				"position": %d,
				"duration": %d
			}
		]`, configuration.MODULE_ID, configuration.OCCURENCE_TYPE, strings.Join(configuration.OCCURENCE_DEFINITION, `","`), configuration.POSITION, configuration.DURATION)
	query := `CALL insert_configurations($1, $2::jsonb);`
	_, err := db.Exec(query, configuration.CAREPARTNER_ID, jsonConfig)
	if err != nil {
		return fmt.Errorf("failed to call insert_configurations: %w", err)
	}
	err = db.QueryRow(`
			SELECT id FROM "Configuration"
			WHERE carepartner_id = $1 AND position = $2
			ORDER BY id DESC
			LIMIT 1;
		`, configuration.CAREPARTNER_ID, configuration.POSITION).Scan(&configuration.ID)
	return err
}

func CreateConfigurationsBulk(db *sql.DB, configurations []Configuration) error {
	jsonConfig := `[`
	for i, configuration := range configurations {
		if i < len(configurations)-1 {
			jsonConfig = fmt.Sprintf(`%s 
					{
						"id": null,
						"module_id": %d,
						"occurence_type": "%s",
						"occurence_definition": ["%s"],
						"position": %d,
						"duration": %d
					},
				`, jsonConfig, configuration.MODULE_ID, configuration.OCCURENCE_TYPE, strings.Join(configuration.OCCURENCE_DEFINITION, `","`), configuration.POSITION, configuration.DURATION)
		} else {
			jsonConfig = fmt.Sprintf(`%s 
					{
						"id": null,
						"module_id": %d,
						"occurence_type": "%s",
						"occurence_definition": ["%s"],
						"position": %d,
						"duration": %d
					}
				]
				`, jsonConfig, configuration.MODULE_ID, configuration.OCCURENCE_TYPE, strings.Join(configuration.OCCURENCE_DEFINITION, `","`), configuration.POSITION, configuration.DURATION)
		}
	}
	query := `CALL insert_configurations($1, $2::jsonb);`
	_, err := db.Exec(query, configurations[0].CAREPARTNER_ID, jsonConfig)
	if err != nil {
		return fmt.Errorf("failed to call insert_configurations: %w", err)
	}
	for _, configuration := range configurations {
		err = db.QueryRow(`
			SELECT id FROM "Configuration"
			WHERE carepartner_id = $1 AND position = $2
			ORDER BY id DESC
			LIMIT 1;
		`, configuration.CAREPARTNER_ID, configuration.POSITION).Scan(&configuration.ID)
	}
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

func GetConfigurationsByDayAndCarepartner(db *sql.DB, carepartner_id int, day, month, day_of_week int) ([]Configuration, error) {
	dayStr := fmt.Sprintf("%d", day)                          //For Monthly Type - int 1 - 31
	dayOfWeekStr := DAYS_OF_WEEK_STR[day_of_week]             //For Weekly Type - string "Sun" - "Sat"
	monthDay := fmt.Sprintf("%s %d", MONTH_STR[month-1], day) //For Yearly Type - string "Month day" - i.e. "Jan 1"
	var configurations []Configuration

	query := `
			SELECT id, carepartner_id, module_id, occurence_type, occurence_definition, position, duration
			FROM "Configuration"
			WHERE carepartner_id = $1
			AND (
				occurence_type = 'Daily'
				OR (occurence_type = 'Weekly' AND $2 = ANY(occurence_definition))
				OR (occurence_type = 'Monthly' AND $3 = ANY(occurence_definition))
				OR (occurence_type = 'Yearly' AND $4 = ANY(occurence_definition))
			)
			ORDER BY position;
	`
	rows, err := db.Query(query, carepartner_id, dayOfWeekStr, dayStr, monthDay)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var configuration Configuration
	for rows.Next() {
		err := rows.Scan(&configuration.ID, &configuration.CAREPARTNER_ID, &configuration.MODULE_ID, &configuration.OCCURENCE_TYPE, pq.Array(&configuration.OCCURENCE_DEFINITION), &configuration.POSITION, &configuration.DURATION)
		if err != nil {
			return nil, fmt.Errorf("error scanning: %v", err)
		}
		configurations = append(configurations, configuration)
	}
	err = rows.Err()
	if err != nil {
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

// Update Configuration
func UpdateConfiguration(db *sql.DB, configuration *Configuration) error {
	query := `CALL delete_configurations($1, $2);`
	id_array := []int{configuration.ID}
	_, err := db.Exec(query, configuration.CAREPARTNER_ID, pq.Array(id_array))
	if err != nil {
		return err
	}
	jsonConfig := fmt.Sprintf(`[
			{
				"id": null,
				"module_id": %d,
				"occurence_type": "%s",
				"occurence_definition": ["%s"],
				"position": %d,
				"duration": %d
			}
		]`, configuration.MODULE_ID, configuration.OCCURENCE_TYPE, strings.Join(configuration.OCCURENCE_DEFINITION, `","`), configuration.POSITION, configuration.DURATION)
	query = `CALL insert_configurations($1, $2::jsonb);`
	_, err = db.Exec(query, configuration.CAREPARTNER_ID, jsonConfig)
	if err != nil {
		return fmt.Errorf("failed to call insert_configurations: %w", err)
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
	query := `CALL delete_configurations($1, $2);`
	id_array := []int{id}
	_, err := db.Exec(query, carepartner_id, pq.Array(id_array))
	if err != nil {
		return err
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
