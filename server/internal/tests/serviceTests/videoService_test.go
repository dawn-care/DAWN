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

func TestCreateVideo(t *testing.T) {
	db, mock := tests.SetupMockDB(t)
	defer db.Close()

	minio := tests.SetupMockMinIO(t)

	service := servicesMedia.NewVideoService(db, minio, "test-video")

	video := &modelsMedia.VideoMedia{
		URI:       "video_uri",
		Filename:  "movie.mp4",
		Extension: "mp4",
	}

	mock.ExpectQuery(regexp.QuoteMeta(`
		INSERT INTO "VideoMedia" (uri, filename, extension)
		VALUES ($1, $2, $3)
		RETURNING id;
	`)).
		WithArgs(video.URI, video.Filename, video.Extension).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	err := service.CreateVideo(video)
	assert.NoError(t, err)
	assert.Equal(t, 1, video.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateVideoFileService(t *testing.T) {
	db, _ := tests.SetupMockDB(t)
	defer db.Close()

	minio := tests.SetupMockMinIO(t)

	service := servicesMedia.NewVideoService(db, minio, "test-video")

	file, err := os.Open("..\\testFiles\\test_video.mp4")
	if err != nil {
		t.Fatalf("failed to open file: %v", err)
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", filepath.Base("server\\internal\\tests\\testFiles\\test_video.mp4"))
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

	serviceErr := service.CreateVideoFile(reqfile, handler, "users/test-user/test_video.mp4")
	assert.NoError(t, serviceErr)
}

func TestGetVideoByID(t *testing.T) {
	db, mock := tests.SetupMockDB(t)
	defer db.Close()

	minio := tests.SetupMockMinIO(t)

	service := servicesMedia.NewVideoService(db, minio, "test-video")

	rows := sqlmock.NewRows([]string{"id", "uri", "filename", "extension"}).
		AddRow(1, "video_uri", "movie.mp4", "mp4")

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, uri, filename, extension
		FROM "VideoMedia"
		WHERE id = $1;
	`)).
		WithArgs(1).
		WillReturnRows(rows)

	video, err := service.GetVideoByID(1)
	assert.NoError(t, err)
	assert.NotNil(t, video)
	assert.Equal(t, "movie.mp4", video.Filename)
	assert.Equal(t, "mp4", video.Extension)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetVideoFileService(t *testing.T) {
	db, _ := tests.SetupMockDB(t)
	defer db.Close()

	minio := tests.SetupMockMinIO(t)

	service := servicesMedia.NewVideoService(db, minio, "test-video")

	file, err := service.GetVideoFile("users/test-user/test_video.mp4")
	assert.NoError(t, err)
	defer file.Close()

	gotBytes, err := io.ReadAll(file)
	assert.NoError(t, err)

	localPath := "../testFiles/test_video.mp4"
	expectedBytes, err := os.ReadFile(localPath)
	assert.NoError(t, err)

	assert.Equal(t, expectedBytes, gotBytes)
}

func TestUpdateVideo(t *testing.T) {
	db, mock := tests.SetupMockDB(t)
	defer db.Close()

	minio := tests.SetupMockMinIO(t)

	service := servicesMedia.NewVideoService(db, minio, "test-video")

	video := &modelsMedia.VideoMedia{
		ID:        1,
		URI:       "video_uri_updated",
		Filename:  "movie_updated.mp4",
		Extension: "mp4",
	}

	mock.ExpectExec(regexp.QuoteMeta(`
		UPDATE "VideoMedia"
		SET uri = $1, filename = $2, extension = $3
		WHERE id = $4;
	`)).
		WithArgs(video.URI, video.Filename, video.Extension, video.ID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := service.UpdateVideo(video)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateVideoFileService(t *testing.T) {
	db, _ := tests.SetupMockDB(t)
	defer db.Close()

	minio := tests.SetupMockMinIO(t)

	service := servicesMedia.NewVideoService(db, minio, "test-video")

	file, err := os.Open("..\\testFiles\\test_video.mp4")
	if err != nil {
		t.Fatalf("failed to open file: %v", err)
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", filepath.Base("server\\internal\\tests\\testFiles\\test_video.mp4"))
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

	serviceErr := service.UpdateVideoFile(reqfile, handler, "users/test-user/test_video.mp4")
	assert.NoError(t, serviceErr)
}

func TestDeleteVideo(t *testing.T) {
	db, mock := tests.SetupMockDB(t)
	defer db.Close()

	minio := tests.SetupMockMinIO(t)

	service := servicesMedia.NewVideoService(db, minio, "test-video")
	mock.ExpectExec(regexp.QuoteMeta(`
		DELETE FROM "VideoMedia"
		WHERE id = $1;
	`)).
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := service.DeleteVideo(1)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteVideoFileService(t *testing.T) {
	db, _ := tests.SetupMockDB(t)
	defer db.Close()

	minio := tests.SetupMockMinIO(t)

	service := servicesMedia.NewVideoService(db, minio, "test-video")

	err := service.DeleteVideoFile("users/test-user/test_video.mp4")
	assert.NoError(t, err)
}

func TestGetVideoStats(t *testing.T) {
	db, mock := tests.SetupMockDB(t)
	defer db.Close()

	minio := tests.SetupMockMinIO(t)

	service := servicesMedia.NewVideoService(db, minio, "test-video")

	rows := sqlmock.NewRows([]string{"count"}).AddRow(5)

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT COUNT(*) FROM "VideoMedia";
	`)).WillReturnRows(rows)

	count, err := service.GetVideoStats()
	assert.NoError(t, err)
	assert.Equal(t, 5, count)
	assert.NoError(t, mock.ExpectationsWereMet())
}
