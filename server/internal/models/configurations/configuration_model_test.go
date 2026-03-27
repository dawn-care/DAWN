package configurations

import (
	"database/sql"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
)

// SETUP

func setupMockDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock database: %v", err)
	}
	return db, mock
}

// CREATE

func TestCreateConfiguration(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	occ := []string{"Mon"}
	cfg := &Configuration{
		CAREPARTNER_ID:       1000,
		MODULE_ID:            10,
		OCCURENCE_TYPE:       "Weekly",
		OCCURENCE_DEFINITION: occ,
		POSITION:             1,
		DURATION:             60,
		TIMED:                false,
		TIME:                 nil,
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
		]`, cfg.MODULE_ID, cfg.OCCURENCE_TYPE, strings.Join(cfg.OCCURENCE_DEFINITION, `","`), cfg.POSITION, cfg.DURATION)

	mock.ExpectExec(regexp.QuoteMeta(`
		CALL insert_configurations($1, $2::jsonb);
	`)).
		WithArgs(cfg.CAREPARTNER_ID, jsonConfig).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id FROM "Configuration"
			WHERE carepartner_id = $1 AND position = $2
			ORDER BY id DESC
			LIMIT 1;
	`)).
		WithArgs(cfg.CAREPARTNER_ID, cfg.POSITION).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	err := CreateConfiguration(db, cfg)
	assert.NoError(t, err)
	assert.Equal(t, 1, cfg.ID)

	assert.NoError(t, mock.ExpectationsWereMet())
}

// RETRIEVE

func TestGetConfigurationByID(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	rows := sqlmock.NewRows([]string{
		"id",
		"carepartner_id",
		"module_id",
		"occurence_type",
		"occurence_definition",
		"position",
		"duration",
		"timed",
		"time",
	}).AddRow(
		1, 1000, 10, "Weekly", pq.Array([]string{"Mon"}), 1, 60, false, nil,
	)

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, carepartner_id, module_id, occurence_type, occurence_definition, position, duration, timed, "time"
		FROM "Configuration"
		WHERE id = $1;
	`)).
		WithArgs(1).
		WillReturnRows(rows)

	cfg, err := GetConfigurationByID(db, 1)
	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Equal(t, "Weekly", cfg.OCCURENCE_TYPE)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetConfigurationsByDayAndCarepartner(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	rows := sqlmock.NewRows([]string{
		"id", "carepartner_id", "module_id", "occurence_type", "occurence_definition", "position", "duration", "timed", "time",
	}).
		AddRow(1, 1000, 10, "Daily", pq.Array([]string{""}), 1, 60, false, nil).
		AddRow(2, 1000, 11, "Weekly", pq.Array([]string{"mon"}), 2, 120, false, nil).
		AddRow(3, 1000, 12, "Monthly", pq.Array([]string{"17"}), 3, 300, false, nil).
		AddRow(4, 1000, 13, "Yearly", pq.Array([]string{"Nov 17"}), 4, 600, false, nil)

	mock.ExpectQuery(regexp.QuoteMeta(`
	SELECT id, carepartner_id, module_id, occurence_type, occurence_definition, position, duration, timed, "time"
	FROM "Configuration"
	WHERE carepartner_id = $1
	  AND (
			occurence_type = 'Daily'
		OR (occurence_type = 'Weekly' AND $2 = ANY(occurence_definition))
		OR (occurence_type = 'Monthly' AND $3 = ANY(occurence_definition))
		OR (occurence_type = 'Yearly' AND $4 = ANY(occurence_definition))
	  )
	ORDER BY position;
	`)).
		WithArgs(
			1000,     // carepartner_id
			"Mon",    // weekly occurence_definition
			"17",     // monthly occurence_definition
			"Nov 17", // yearly occurence_definition
		).
		WillReturnRows(rows)

	cfgs, err := GetConfigurationsByDayAndCarepartner(db, 1000, 17, 11, 1, "2026-11-17")
	assert.NoError(t, err)
	assert.Len(t, cfgs, 4)

	assert.Equal(t, 10, cfgs[0].MODULE_ID)
	assert.Equal(t, "Weekly", cfgs[1].OCCURENCE_TYPE)

	assert.NoError(t, mock.ExpectationsWereMet())
}

// UPDATE

func TestUpdateConfiguration(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	cfg := &Configuration{
		ID:                   1,
		CAREPARTNER_ID:       1000,
		MODULE_ID:            10,
		OCCURENCE_TYPE:       "Weekly",
		OCCURENCE_DEFINITION: []string{"mon"},
		POSITION:             1,
		DURATION:             60, TIMED: false,
		TIME: nil}

	mock.ExpectExec(regexp.QuoteMeta(`
		CALL delete_configurations($1, $2);
	`)).
		WithArgs(
			cfg.CAREPARTNER_ID,
			pq.Array([]int{cfg.ID}),
		).
		WillReturnResult(sqlmock.NewResult(0, 1))

	jsonConfig := fmt.Sprintf(`[
			{
				"id": null,
				"module_id": %d,
				"occurence_type": "%s",
				"occurence_definition": ["%s"],
				"position": %d,
				"duration": %d
			}
		]`, cfg.MODULE_ID, cfg.OCCURENCE_TYPE, strings.Join(cfg.OCCURENCE_DEFINITION, `","`), cfg.POSITION, cfg.DURATION)

	mock.ExpectExec(regexp.QuoteMeta(`
		CALL insert_configurations($1, $2::jsonb);
	`)).
		WithArgs(cfg.CAREPARTNER_ID, jsonConfig).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id FROM "Configuration"
		WHERE carepartner_id = $1 AND position = $2
		ORDER BY id DESC
		LIMIT 1;
	`)).
		WithArgs(cfg.CAREPARTNER_ID, cfg.POSITION).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(99))

	err := UpdateConfiguration(db, cfg)
	assert.NoError(t, err)
	assert.Equal(t, 99, cfg.ID)

	assert.NoError(t, mock.ExpectationsWereMet())
}

// DELETE

func TestDeleteConfiguration(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	mock.ExpectExec(regexp.QuoteMeta(`
		CALL delete_configurations($1, $2);
	`)).
		WithArgs(
			1000,
			pq.Array([]int{1}),
		).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := DeleteConfiguration(db, 1, 1000)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ANALYTICS & STATS

func TestGetConfigurationStats(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	rows := sqlmock.NewRows([]string{"count"}).AddRow(3)

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT COUNT(*) FROM "Configuration";
	`)).
		WillReturnRows(rows)

	count, err := GetConfigurationStats(db)
	assert.NoError(t, err)
	assert.Equal(t, 3, count)

	assert.NoError(t, mock.ExpectationsWereMet())
}
