package modules

import (
	"database/sql"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func setupMockDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock database: %v", err)
	}
	return db, mock
}

func TestCreateModule(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	mod := &Module{
		ModuleType:    "video",
		CaregiverID:   301,
		CarepartnerID: 201,
		MetadataID:    1,
	}

	mock.ExpectQuery(regexp.QuoteMeta(`
		INSERT INTO "Module" (module_type, caregiver_id, carepartner_id, metadata_id)
		VALUES ($1, $2, $3, $4)
		RETURNING id;
	`)).
		WithArgs(mod.ModuleType, mod.CaregiverID, mod.CarepartnerID, mod.MetadataID).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(10))

	err := CreateModule(db, mod)
	assert.NoError(t, err)
	assert.Equal(t, 10, mod.ID)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetModuleByID(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	rows := sqlmock.NewRows([]string{"id", "module_type", "caregiver_id", "carepartner_id", "metadata_id"}).
		AddRow(10, "video", 301, 201, 1)

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, module_type, caregiver_id, carepartner_id, metadata_id
		FROM "Module"
		WHERE id = $1;
	`)).
		WithArgs(10).
		WillReturnRows(rows)

	mod, err := GetModuleByID(db, 10)
	assert.NoError(t, err)
	assert.NotNil(t, mod)
	assert.Equal(t, "video", mod.ModuleType)
	assert.Equal(t, 301, mod.CaregiverID)
	assert.Equal(t, 201, mod.CarepartnerID)
	assert.Equal(t, 1, mod.MetadataID)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetMetadataByCaregiverCarepartnerAndType(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	rows := sqlmock.NewRows([]string{"metadata_id"}).
		AddRow(1).
		AddRow(2)

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT metadata_id
		FROM "Module"
		WHERE caregiver_id = $1 AND carepartner_id = $2 AND module_type = $3;
	`)).
		WithArgs(301, 201, "video").
		WillReturnRows(rows)

	metadataIDs, err := GetMetadataByCaregiverCarepartnerAndType(db, 301, 201, "video")
	assert.NoError(t, err)
	assert.Equal(t, []int{1, 2}, metadataIDs)

	assert.NoError(t, mock.ExpectationsWereMet())
}
