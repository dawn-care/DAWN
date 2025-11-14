package models

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"reflect"
	"testing"

	"server/internal/models/configurations"

	_ "github.com/lib/pq"
	// "github.com/joho/godotenv"
)

var testDB *sql.DB
var DB_TABLES = [11]string{"Configuration", "Module", "PhotoMedia", "VideoMedia", "AudioMedia", "WeatherWidget", "NewsWidget", "CarepartnerLocationPrompt", "CaregiverLocationPrompt", "MessagePrompt", "EventPrompt"}

// Sets Up Postgres Connections
func TestMain(m *testing.M) {
	var err error
	connStr := "host=localhost port=5432 user=myuser password=mypassword dbname=mydb sslmode=disable"
	testDB, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	code := m.Run()
	testDB.Close()
	os.Exit(code)
}

// Helper Function to Clear Test Tables
func clearDBTable(t *testing.T) {
	for _, value := range DB_TABLES {
		query := fmt.Sprintf(`DELETE FROM "%s";`, value)
		_, err := testDB.Exec(query)
		if err != nil {
			t.Fatalf("Failed to clear table: %v", err)
		}
	}
}

func TestCreateConfiguration(t *testing.T) {
	clearDBTable(t)

	var occurence_definition = []string{"Mon"}
	configuration := &configurations.Configuration{
		CAREPARTNER_ID:       1000,
		MODULE_ID:            10,
		OCCURENCE_TYPE:       "Weekly",
		OCCURENCE_DEFINITION: occurence_definition,
		POSITION:             1,
	}

	err := configurations.CreateConfiguration(testDB, configuration)
	if err != nil {
		t.Fatalf("CreateConfiguration Failed: %v", err)
	}
	if configuration.ID == 0 {
		t.Errorf("Configuration ID Not Set")
	}

	found_configuration, err := configurations.GetConfigurationByID(testDB, configuration.ID)
	if err != nil {
		t.Fatalf("GetConfigurationByID Failed: %v", err)
	}
	if found_configuration == nil {
		t.Fatalf("Configuration Not Found")
	}
	if !reflect.DeepEqual(*found_configuration, *configuration) {
		t.Fatalf("Unexpected Configuration Returned: %+v", found_configuration)
	}
}

func TestUpdateConfiguration(t *testing.T) {
	clearDBTable(t)

	var occurence_definition = []string{}
	configuration := &configurations.Configuration{
		CAREPARTNER_ID:       1000,
		MODULE_ID:            10,
		OCCURENCE_TYPE:       "Daily",
		OCCURENCE_DEFINITION: occurence_definition,
		POSITION:             1,
	}

	err := configurations.CreateConfiguration(testDB, configuration)
	if err != nil {
		t.Fatalf("CreateConfiguration Failed: %v", err)
	}

	var updated_occurence_definition = []string{"mon"}
	configuration.OCCURENCE_DEFINITION = updated_occurence_definition
	configuration.OCCURENCE_TYPE = "Weekly"
	err = configurations.UpdateConfiguration(testDB, configuration)
	if err != nil {
		t.Fatalf("UpdateConfiguration Failed: %v", err)
	}
	found_configuration, err := configurations.GetConfigurationByID(testDB, configuration.ID)
	if !reflect.DeepEqual(*found_configuration, *configuration) {
		t.Errorf("Unexpected Configuration Returned (UpdateConfiguration): %+v", found_configuration)
	}
	if err != nil {
		t.Fatalf("GetConfigurationByID Failed: %v", err)
	}
}

func TestDeleteConfiguration(t *testing.T) {
	clearDBTable(t)

	var occurence_definition = []string{}
	configuration := &configurations.Configuration{
		CAREPARTNER_ID:       1000,
		MODULE_ID:            10,
		OCCURENCE_TYPE:       "Daily",
		OCCURENCE_DEFINITION: occurence_definition,
		POSITION:             1,
	}

	err := configurations.CreateConfiguration(testDB, configuration)
	if err != nil {
		t.Fatalf("CreateConfiguration Failed: %v", err)
	}

	err = configurations.DeleteConfiguration(testDB, configuration.ID, configuration.CAREPARTNER_ID)
	if err != nil {
		t.Fatalf("DeleteConfiguration Failed: %v", err)
	}
	found_configuration, err := configurations.GetConfigurationByID(testDB, configuration.ID)
	if found_configuration != nil {
		t.Errorf("Unexpected Configuration Returned (DeleteConfiguration): %+v", found_configuration)
	}
	if err != nil {
		t.Fatalf("GetConfigurationByID Failed: %v", err)
	}
}

func TestGetConfigurationsByDayAndCarepartnerID(t *testing.T) {
	clearDBTable(t)
	// Define Test Occurence Definiton for All Configurations
	var occurence_definition_daily = []string{""}
	var occurence_definition_weekly = []string{"mon"}
	var occurence_definition_monthly = []string{"17"}
	var occurence_definition_yearly = []string{"Nov 17"}
	// Create Configurations
	test_configurations := []configurations.Configuration{
		{
			CAREPARTNER_ID:       1000,
			MODULE_ID:            10,
			OCCURENCE_TYPE:       "Daily",
			OCCURENCE_DEFINITION: occurence_definition_daily,
			POSITION:             1,
		},
		{
			CAREPARTNER_ID:       1000,
			MODULE_ID:            11,
			OCCURENCE_TYPE:       "Weekly",
			OCCURENCE_DEFINITION: occurence_definition_weekly,
			POSITION:             2,
		},
		{
			CAREPARTNER_ID:       1000,
			MODULE_ID:            12,
			OCCURENCE_TYPE:       "Monthly",
			OCCURENCE_DEFINITION: occurence_definition_monthly,
			POSITION:             3,
		},
		{
			CAREPARTNER_ID:       1000,
			MODULE_ID:            13,
			OCCURENCE_TYPE:       "Yearly",
			OCCURENCE_DEFINITION: occurence_definition_yearly,
			POSITION:             4,
		},
	}
	err := configurations.CreateConfigurationsBulk(testDB, test_configurations)
	if err != nil {
		t.Fatalf("CreateConfigurationsBulk Failed: %v", err)
	}

	found_configurations, err := configurations.GetConfigurationsByDayAndCarepartner(testDB, 1000, 17, 11, 1)
	if err != nil {
		t.Fatalf("GetConfigurationsByDayAndCarepartner Failed: %v", err)
	}
	for i := range found_configurations {
		if test_configurations[i].CAREPARTNER_ID != found_configurations[i].CAREPARTNER_ID &&
			test_configurations[i].MODULE_ID != found_configurations[i].MODULE_ID &&
			test_configurations[i].OCCURENCE_TYPE != found_configurations[i].OCCURENCE_TYPE &&
			!reflect.DeepEqual(test_configurations[i].OCCURENCE_DEFINITION, found_configurations[i].OCCURENCE_DEFINITION) &&
			test_configurations[i].POSITION != found_configurations[i].POSITION {
			t.Fatalf("Unexpected Configuration Returned in Position %d: %+v\nExpected: %+v", i+1, found_configurations[i], test_configurations[i])
		}
	}
}

func TestGetConfigurationStats(t *testing.T) {
	clearDBTable(t)

	var occurence_definition = []string{""}
	test_configurations := []configurations.Configuration{
		{
			CAREPARTNER_ID:       1000,
			MODULE_ID:            10,
			OCCURENCE_TYPE:       "Daily",
			OCCURENCE_DEFINITION: occurence_definition,
			POSITION:             1,
		},
		{
			CAREPARTNER_ID:       1000,
			MODULE_ID:            11,
			OCCURENCE_TYPE:       "Daily",
			OCCURENCE_DEFINITION: occurence_definition,
			POSITION:             2,
		},
		{
			CAREPARTNER_ID:       1000,
			MODULE_ID:            12,
			OCCURENCE_TYPE:       "Daily",
			OCCURENCE_DEFINITION: occurence_definition,
			POSITION:             3,
		},
	}
	err := configurations.CreateConfigurationsBulk(testDB, test_configurations)
	if err != nil {
		t.Fatalf("CreateConfigurationsBulk Failed: %v", err)
	}

	count, err := configurations.GetConfigurationStats(testDB)
	if count != 3 {
		t.Errorf("Unexpected Count Returned: %d", count)
	}
}
