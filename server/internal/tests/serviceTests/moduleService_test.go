package mediaTests

import (
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"

	moduleService "server/internal/services/modules"
)

func TestCreateModuleWithMedia(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock database: %v", err)
	}
	defer db.Close()

	svc := moduleService.NewModuleService(db)
	
	mock.ExpectBegin()

	// Expect media insert (for example, audio)
	mock.ExpectQuery(regexp.QuoteMeta(`
		INSERT INTO "AudioMedia" (uri, filename, extension)
		VALUES ($1, $2, $3)
		RETURNING id;
	`)).
		WithArgs("uri_test", "song.mp3", "mp3").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	// Expect module insert
	mock.ExpectQuery(regexp.QuoteMeta(`
		INSERT INTO "Module" (module_type, caregiver_id, carepartner_id, metadata_id)
		VALUES ($1, $2, $3, $4)
		RETURNING id;
	`)).
		WithArgs("audio", 10, 20, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(100))

	mock.ExpectCommit()

	mod, err := svc.CreateModuleWithMedia("audio", 10, 20, "uri_test", "song.mp3", "mp3")
	assert.NoError(t, err)
	assert.NotNil(t, mod)
	assert.Equal(t, 100, mod.ID)
	assert.Equal(t, 1, mod.MetadataID)
	assert.Equal(t, "audio", mod.ModuleType)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetModuleByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock database: %v", err)
	}
	defer db.Close()

	svc := moduleService.NewModuleService(db)

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
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock database: %v", err)
	}
	defer db.Close()

	svc := moduleService.NewModuleService(db)

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
