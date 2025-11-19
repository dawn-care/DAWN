package mediaTests

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"

	modelsMedia "server/internal/models/modules/media"
	servicesMedia "server/internal/services/modules/media"
	"server/internal/tests"
)

func TestCreatePhotoService(t *testing.T) {
	db, mock := tests.SetupMockDB(t)
	defer db.Close()

	minio := tests.SetupMockMinIO(t)

	service := servicesMedia.NewPhotoService(db, minio, "test-photo")

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

func TestCreatePhotoFileService(t *testing.T) {
	db, _ := tests.SetupMockDB(t)
	defer db.Close()

	minio := tests.SetupMockMinIO(t)

	service := servicesMedia.NewPhotoService(db, minio, "test-photo")

	file, err := os.Open("..\\testFiles\\test_photo.png")
	if err != nil {
		t.Fatalf("failed to open file: %v", err)
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", filepath.Base("server\\internal\\tests\\testFiles\\test_photo.png"))
	if err != nil {
		t.Fatalf("failed to create form file: %v", err)
	}

	if _, err := io.Copy(part, file); err != nil {
		t.Fatalf("failed to copy file content: %v", err)
	}

	if err := writer.Close(); err != nil {
		t.Fatalf("failed to close writer: %v", err)
	}

	r := httptest.NewRequest(http.MethodPost, "/upload", body)
	r.Header.Set("Content-Type", writer.FormDataContentType())

	if err := r.ParseMultipartForm(100 << 20); err != nil {
        t.Fatalf("bad file: %v", err)
    }

	reqfile, handler, err := r.FormFile("file")
    if err != nil {
        t.Fatalf("bad file: %v", err)
    }
    defer file.Close()

	serviceErr := service.CreatePhotoFile(reqfile, handler, "users/test-user/test_photo.png")
	assert.NoError(t, serviceErr)
}

func TestCreatePhotoService_NilPhoto(t *testing.T) {
	db, _ := tests.SetupMockDB(t)
	defer db.Close()

	minio := tests.SetupMockMinIO(t)

	service := servicesMedia.NewPhotoService(db, minio, "test-photo")

	err := service.CreatePhoto(nil)
	assert.EqualError(t, err, "photo cannot be nil")
}

func TestCreatePhotoService_InvalidPhoto(t *testing.T) {
	db, _ := tests.SetupMockDB(t)
	defer db.Close()

	minio := tests.SetupMockMinIO(t)

	service := servicesMedia.NewPhotoService(db, minio, "test-photo")

	photo := &modelsMedia.PhotoMedia{}
	err := service.CreatePhoto(photo)
	assert.EqualError(t, err, "uri, filename, and extension are required")
}

func TestGetPhotoByIDService(t *testing.T) {
	db, mock := tests.SetupMockDB(t)
	defer db.Close()

	minio := tests.SetupMockMinIO(t)

	service := servicesMedia.NewPhotoService(db, minio, "test-photo")

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

func TestGetPhotoFileService(t *testing.T) {
	db, _ := tests.SetupMockDB(t)
	defer db.Close()

	minio := tests.SetupMockMinIO(t)

	service := servicesMedia.NewAudioService(db, minio, "test-photo")

	file, err := service.GetAudioFile("users/test-user/test_photo.png")
	assert.NoError(t, err)
	defer file.Close()

	gotBytes, err := io.ReadAll(file)
	assert.NoError(t, err)

	localPath := "../testFiles/test_photo.png"
	expectedBytes, err := os.ReadFile(localPath)
	assert.NoError(t, err)

	assert.Equal(t, expectedBytes, gotBytes)
}

func TestGetPhotoByIDService_InvalidID(t *testing.T) {
	db, _ := tests.SetupMockDB(t)
	defer db.Close()

	minio := tests.SetupMockMinIO(t)

	service := servicesMedia.NewPhotoService(db, minio, "test-photo")

	_, err := service.GetPhotoByID(0)
	assert.EqualError(t, err, "invalid photo ID")
}

func TestUpdatePhotoService(t *testing.T) {
	db, mock := tests.SetupMockDB(t)
	defer db.Close()

	minio := tests.SetupMockMinIO(t)

	service := servicesMedia.NewPhotoService(db, minio, "test-photo")

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

func TestUpdatePhotoFileService(t *testing.T) {
	db, _ := tests.SetupMockDB(t)
	defer db.Close()

	minio := tests.SetupMockMinIO(t)

	service := servicesMedia.NewPhotoService(db, minio, "test-photo")

	file, err := os.Open("..\\testFiles\\test_photo.png")
	if err != nil {
		t.Fatalf("failed to open file: %v", err)
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", filepath.Base("server\\internal\\tests\\testFiles\\test_photo.png"))
	if err != nil {
		t.Fatalf("failed to create form file: %v", err)
	}

	if _, err := io.Copy(part, file); err != nil {
		t.Fatalf("failed to copy file content: %v", err)
	}

	if err := writer.Close(); err != nil {
		t.Fatalf("failed to close writer: %v", err)
	}

	r := httptest.NewRequest(http.MethodPost, "/upload", body)
	r.Header.Set("Content-Type", writer.FormDataContentType())

	if err := r.ParseMultipartForm(100 << 20); err != nil {
        t.Fatalf("bad file: %v", err)
    }

	reqfile, handler, err := r.FormFile("file")
    if err != nil {
        t.Fatalf("bad file: %v", err)
    }
    defer file.Close()

	serviceErr := service.UpdatePhotoFile(reqfile, handler, "users/test-user/test_photo.png")
	assert.NoError(t, serviceErr)
}


func TestUpdatePhotoService_InvalidPhoto(t *testing.T) {
	db, _ := tests.SetupMockDB(t)
	defer db.Close()

	minio := tests.SetupMockMinIO(t)

	service := servicesMedia.NewPhotoService(db, minio, "test-photo")

	err := service.UpdatePhoto(nil)
	assert.EqualError(t, err, "photo cannot be nil")

	err = service.UpdatePhoto(&modelsMedia.PhotoMedia{ID: 0})
	assert.EqualError(t, err, "invalid photo ID")
}

func TestDeletePhotoService(t *testing.T) {
	db, mock := tests.SetupMockDB(t)
	defer db.Close()

	minio := tests.SetupMockMinIO(t)

	service := servicesMedia.NewPhotoService(db, minio, "test-photo")

	mock.ExpectExec(regexp.QuoteMeta(
		`DELETE FROM "PhotoMedia" WHERE id = $1;`,
	)).
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := service.DeletePhoto(1)
	assert.NoError(t, err)
}

func TestDeletePhotoileService(t *testing.T) {
	db, _ := tests.SetupMockDB(t)
	defer db.Close()

	minio := tests.SetupMockMinIO(t)

	service := servicesMedia.NewAudioService(db, minio, "test-photo")

	err := service.DeleteAudioFile("users/test-user/test_photo.png")
	assert.NoError(t, err)
}


func TestDeletePhotoService_InvalidID(t *testing.T) {
	db, _ := tests.SetupMockDB(t)
	defer db.Close()

	minio := tests.SetupMockMinIO(t)

	service := servicesMedia.NewPhotoService(db, minio, "test-photo")

	err := service.DeletePhoto(0)
	assert.EqualError(t, err, "invalid photo ID")
}

func TestGetPhotoStatsService(t *testing.T) {
	db, mock := tests.SetupMockDB(t)
	defer db.Close()

	minio := tests.SetupMockMinIO(t)

	service := servicesMedia.NewPhotoService(db, minio, "test-photo")

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT COUNT(*) FROM "PhotoMedia";`)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))

	count, err := service.GetPhotoStats()
	assert.NoError(t, err)
	assert.Equal(t, 5, count)
}
