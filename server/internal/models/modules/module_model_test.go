package modules

import (
	"regexp"
	"testing"

	"server/internal/utils"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
)

// ---------------------------
// CREATE
// ---------------------------

func TestCreateModule(t *testing.T) {
	db, mock := utils.SetupMockDB(t)
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

// ---------------------------
// RETRIEVE
// ---------------------------

func TestGetModuleByID(t *testing.T) {
	db, mock := utils.SetupMockDB(t)
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

// ---------------------------
// METADATA
// ---------------------------

func TestGetMetadataByCaregiverCarepartnerAndType(t *testing.T) {
	db, mock := utils.SetupMockDB(t)
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

// ---------------------------
// CHILD RECORDS (generic fetch test example)
// ---------------------------

func TestFetchChildRecords(t *testing.T) {
	db, mock := utils.SetupMockDB(t)
	defer db.Close()

	ids := []int{1, 2}
	rows := sqlmock.NewRows([]string{"id", "uri", "filename"}).
		AddRow(1, "uri1", "file1").
		AddRow(2, "uri2", "file2")

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "PhotoMedia" WHERE id = ANY($1)`)).
		WithArgs(pq.Array(ids)).
		WillReturnRows(rows)

	results, err := GetPhotoMediaByIDs(db, ids)
	assert.NoError(t, err)
	assert.Len(t, results, 2)
	assert.Equal(t, "uri1", results[0]["uri"])
	assert.Equal(t, "file2", results[1]["filename"])

	assert.NoError(t, mock.ExpectationsWereMet())
}
