package modules

import (
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestCreateModule(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	svc := NewModuleService(db)

	rows := sqlmock.NewRows([]string{"id"}).AddRow(1)
	mock.ExpectQuery(regexp.QuoteMeta(`
		INSERT INTO "Module" (module_type, caregiver_id, carepartner_id, metadata_id)
		VALUES ($1, $2, $3, $4)
		RETURNING id;
	`)).
		WithArgs("audio", 1, 2, 99).
		WillReturnRows(rows)

	mod, err := svc.CreateModule("audio", 1, 2, 99)
	assert.NoError(t, err)
	assert.NotNil(t, mod)
	assert.Equal(t, "audio", mod.ModuleType)
	assert.Equal(t, 1, mod.CaregiverID)
	assert.Equal(t, 2, mod.CarepartnerID)
	assert.Equal(t, 99, mod.MetadataID)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetModuleByID(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	svc := NewModuleService(db)

	rows := sqlmock.NewRows([]string{"id", "module_type", "caregiver_id", "carepartner_id", "metadata_id"}).
		AddRow(1, "video", 10, 20, 30)

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, module_type, caregiver_id, carepartner_id, metadata_id
		FROM "Module"
		WHERE id = $1;
	`)).
		WithArgs(1).
		WillReturnRows(rows)

	mod, err := svc.GetModuleByID(1)
	assert.NoError(t, err)
	assert.NotNil(t, mod)
	assert.Equal(t, "video", mod.ModuleType)
	assert.Equal(t, 10, mod.CaregiverID)
	assert.Equal(t, 20, mod.CarepartnerID)
	assert.Equal(t, 30, mod.MetadataID)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetMetadataByCaregiverCarepartnerAndType(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	svc := NewModuleService(db)

	rows := sqlmock.NewRows([]string{"metadata_id"}).
		AddRow(101).AddRow(102)

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT metadata_id
		FROM "Module"
		WHERE caregiver_id = $1 AND carepartner_id = $2 AND module_type = $3;
	`)).
		WithArgs(5, 6, "photo").
		WillReturnRows(rows)

	metadataIDs, err := svc.GetMetadataByCaregiverCarepartnerAndType(5, 6, "photo")
	assert.NoError(t, err)
	assert.Equal(t, []int{101, 102}, metadataIDs)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetAllMetadataByCaregiverCarepartner(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	svc := NewModuleService(db)

	rows := sqlmock.NewRows([]string{"metadata_id"}).AddRow(201).AddRow(202)

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT metadata_id
		FROM "Module"
		WHERE caregiver_id = $1 AND carepartner_id = $2;
	`)).
		WithArgs(11, 12).
		WillReturnRows(rows)

	metadataIDs, err := svc.GetAllMetadataByCaregiverCarepartner(11, 12)
	assert.NoError(t, err)
	assert.Equal(t, []int{201, 202}, metadataIDs)

	assert.NoError(t, mock.ExpectationsWereMet())
}
