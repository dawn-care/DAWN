package configurations

import (
	"database/sql"
	"regexp"
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

// UPDATE

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
