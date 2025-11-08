package mediaTests

import (
	"database/sql"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"

	modelsMedia "server/internal/models/modules/media"
	servicesMedia "server/internal/services/modules/media"
)

func setupMockDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock database: %v", err)
	}
	return db, mock
}

func TestCreatePhotoService(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	service := servicesMedia.NewPhotoService(db)

	photo := &modelsMedia.PhotoMedia{
		URI:       "https://example.com/test.jpg",
		Filename:  "test",
		Extension: "jpg",
	}

	mock.ExpectQuery(regexp.QuoteMeta(
		`INSERT INTO "PhotoMedia" (uri, filename, extension)
		VALUES ($1, $2, $3)
		RETURNING id;`,
	)).
		WithArgs(photo.URI, photo.Filename, photo.Extension).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	err := service.CreatePhoto(photo)
	assert.NoError(t, err)
	assert.Equal(t, 1, photo.ID)
}

func TestCreatePhotoService_NilPhoto(t *testing.T) {
	db, _ := setupMockDB(t)
	defer db.Close()
	service := servicesMedia.NewPhotoService(db)

	err := service.CreatePhoto(nil)
	assert.EqualError(t, err, "photo cannot be nil")
}

func TestCreatePhotoService_InvalidPhoto(t *testing.T) {
	db, _ := setupMockDB(t)
	defer db.Close()
	service := servicesMedia.NewPhotoService(db)

	photo := &modelsMedia.PhotoMedia{}
	err := service.CreatePhoto(photo)
	assert.EqualError(t, err, "uri, filename, and extension are required")
}

func TestGetPhotoByIDService(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	service := servicesMedia.NewPhotoService(db)

	expectedPhoto := &modelsMedia.PhotoMedia{
		ID:        1,
		URI:       "https://example.com/test.jpg",
		Filename:  "test",
		Extension: "jpg",
	}

	mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT id, uri, filename, extension
		FROM "PhotoMedia"
		WHERE id = $1;`,
	)).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "uri", "filename", "extension"}).
			AddRow(expectedPhoto.ID, expectedPhoto.URI, expectedPhoto.Filename, expectedPhoto.Extension))

	photo, err := service.GetPhotoByID(1)
	assert.NoError(t, err)
	assert.Equal(t, expectedPhoto, photo)
}

func TestGetPhotoByIDService_InvalidID(t *testing.T) {
	db, _ := setupMockDB(t)
	defer db.Close()
	service := servicesMedia.NewPhotoService(db)

	_, err := service.GetPhotoByID(0)
	assert.EqualError(t, err, "invalid photo ID")
}

func TestUpdatePhotoService(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()
	service := servicesMedia.NewPhotoService(db)

	photo := &modelsMedia.PhotoMedia{
		ID:        1,
		URI:       "https://example.com/test.jpg",
		Filename:  "test",
		Extension: "jpg",
	}

	mock.ExpectExec(regexp.QuoteMeta(
		`UPDATE "PhotoMedia"
		SET uri = $1, filename = $2, extension = $3
		WHERE id = $4;`,
	)).
		WithArgs(photo.URI, photo.Filename, photo.Extension, photo.ID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := service.UpdatePhoto(photo)
	assert.NoError(t, err)
}

func TestUpdatePhotoService_InvalidPhoto(t *testing.T) {
	db, _ := setupMockDB(t)
	defer db.Close()
	service := servicesMedia.NewPhotoService(db)

	err := service.UpdatePhoto(nil)
	assert.EqualError(t, err, "photo cannot be nil")

	err = service.UpdatePhoto(&modelsMedia.PhotoMedia{ID: 0})
	assert.EqualError(t, err, "invalid photo ID")
}

func TestDeletePhotoService(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()
	service := servicesMedia.NewPhotoService(db)

	mock.ExpectExec(regexp.QuoteMeta(
		`DELETE FROM "PhotoMedia" WHERE id = $1;`,
	)).
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := service.DeletePhoto(1)
	assert.NoError(t, err)
}

func TestDeletePhotoService_InvalidID(t *testing.T) {
	db, _ := setupMockDB(t)
	defer db.Close()
	service := servicesMedia.NewPhotoService(db)

	err := service.DeletePhoto(0)
	assert.EqualError(t, err, "invalid photo ID")
}

func TestGetPhotoStatsService(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()
	service := servicesMedia.NewPhotoService(db)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT COUNT(*) FROM "PhotoMedia";`)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))

	count, err := service.GetPhotoStats()
	assert.NoError(t, err)
	assert.Equal(t, 5, count)
}
