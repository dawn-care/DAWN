package media

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"

	modelsMedia "server/internal/models/media"
	"server/internal/utils"
)

func TestCreatePhotoService(t *testing.T) {
	db, mock := utils.SetupMockDB(t)
	defer db.Close()

	minio := utils.SetupMockMinIO(t)
	service := NewPhotoService(db, minio, "test-photo")

	photo := &modelsMedia.PhotoMedia{
		URI:           "users/test/test.jpg",
		Filename:      "test",
		Extension:     "jpg",
		DisplayedName: "Test Photo",
	}

	mock.ExpectBegin()

	mock.ExpectQuery(regexp.QuoteMeta(`
	INSERT INTO "PhotoMedia" (uri, displayed_name, filename, extension)
	VALUES ($1, $2, $3, $4)
	RETURNING id;
	`)).
		WithArgs(
			photo.URI,
			photo.DisplayedName,
			photo.Filename,
			photo.Extension,
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	mock.ExpectCommit()

	err := service.CreatePhoto(photo)

	assert.NoError(t, err)
	assert.Equal(t, 1, photo.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreatePhotoService_NilPhoto(t *testing.T) {
	db, _ := utils.SetupMockDB(t)
	defer db.Close()

	minio := utils.SetupMockMinIO(t)
	service := NewPhotoService(db, minio, "test-photo")

	err := service.CreatePhoto(nil)
	assert.EqualError(t, err, "photo cannot be nil")
}

func TestCreatePhotoService_InvalidPhoto(t *testing.T) {
	db, _ := utils.SetupMockDB(t)
	defer db.Close()

	minio := utils.SetupMockMinIO(t)
	service := NewPhotoService(db, minio, "test-photo")

	photo := &modelsMedia.PhotoMedia{}
	err := service.CreatePhoto(photo)

	assert.EqualError(t, err, "uri, filename, and extension are required")
}

func TestCreatePhotoFileService(t *testing.T) {
	db, _ := utils.SetupMockDB(t)
	defer db.Close()

	minio := utils.SetupMockMinIO(t)
	service := NewPhotoService(db, minio, "test-photo")

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", "test.png")
	assert.NoError(t, err)

	_, err = io.Copy(part, bytes.NewBuffer([]byte("fake-image-data")))
	assert.NoError(t, err)

	writer.Close()

	r := httptest.NewRequest(http.MethodPost, "/upload", body)
	r.Header.Set("Content-Type", writer.FormDataContentType())

	err = r.ParseMultipartForm(10 << 20)
	assert.NoError(t, err)

	reqFile, handler, err := r.FormFile("file")
	assert.NoError(t, err)
	defer reqFile.Close()

	serviceErr := service.CreatePhotoFile(reqFile, handler, "users/test-user/test.png")
	assert.NoError(t, serviceErr)
}

func TestGetPhotoByIDService(t *testing.T) {
	db, mock := utils.SetupMockDB(t)
	defer db.Close()

	minio := utils.SetupMockMinIO(t)
	service := NewPhotoService(db, minio, "test-photo")

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, uri, displayed_name, filename, extension
		FROM "PhotoMedia"
		WHERE id = $1;
	`)).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows(
			[]string{"id", "uri", "displayed_name", "filename", "extension"},
		).AddRow(1, "users/test/test.jpg", "Test Photo", "test", "jpg"))

	photo, err := service.GetPhotoByID(1)

	assert.NoError(t, err)
	assert.Equal(t, 1, photo.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetPhotoByID_InvalidID(t *testing.T) {
	db, _ := utils.SetupMockDB(t)
	defer db.Close()

	minio := utils.SetupMockMinIO(t)
	service := NewPhotoService(db, minio, "test-photo")

	_, err := service.GetPhotoByID(0)
	assert.EqualError(t, err, "invalid photo ID")
}

func TestUpdatePhotoService(t *testing.T) {
	db, mock := utils.SetupMockDB(t)
	defer db.Close()

	minio := utils.SetupMockMinIO(t)
	service := NewPhotoService(db, minio, "test-photo")

	photo := &modelsMedia.PhotoMedia{
		ID:            1,
		URI:           "users/test/test.jpg",
		Filename:      "updated",
		DisplayedName: "Updated Name",
		Extension:     "jpg",
	}

	mock.ExpectExec(regexp.QuoteMeta(`
		UPDATE "PhotoMedia"
		SET uri = $1, displayed_name = $2, filename = $3, extension = $4
		WHERE id = $5;
	`)).
	WithArgs(
		photo.URI,
		photo.DisplayedName,
		photo.Filename,
		photo.Extension,
		photo.ID,
	).
	WillReturnResult(sqlmock.NewResult(0, 1))

	err := service.UpdatePhoto(photo)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeletePhotoService(t *testing.T) {
	db, mock := utils.SetupMockDB(t)
	defer db.Close()

	minio := utils.SetupMockMinIO(t)
	service := NewPhotoService(db, minio, "test-photo")

	mock.ExpectExec(regexp.QuoteMeta(
		`DELETE FROM "PhotoMedia" WHERE id = $1;`,
	)).
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := service.DeletePhoto(1)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeletePhoto_InvalidID(t *testing.T) {
	db, _ := utils.SetupMockDB(t)
	defer db.Close()

	minio := utils.SetupMockMinIO(t)
	service := NewPhotoService(db, minio, "test-photo")

	err := service.DeletePhoto(0)
	assert.EqualError(t, err, "invalid photo ID")
}
