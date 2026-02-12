package media

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

	modelsMedia "server/internal/models/media"
	"server/internal/utils"
)

func TestCreateVideo(t *testing.T) {
	db, mock := utils.SetupMockDB(t)
	defer db.Close()

	minio := utils.SetupMockMinIO(t)
	service := NewVideoService(db, minio, "test-video")

	video := &modelsMedia.VideoMedia{
		URI:           "video_uri",
		DisplayedName:   "My Video",
		Filename:      "movie.mp4",
		Extension:     "mp4",
	}

	mock.ExpectBegin()

	mock.ExpectQuery(regexp.QuoteMeta(`
		INSERT INTO "VideoMedia" (uri, displayed_name, filename, extension)
		VALUES ($1, $2, $3, $4)
		RETURNING id;
	`)).
		WithArgs(video.URI, video.DisplayedName, video.Filename, video.Extension).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	mock.ExpectCommit()

	err := service.CreateVideo(video)
	assert.NoError(t, err)
	assert.Equal(t, 1, video.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}


func TestCreateVideoFileService(t *testing.T) {
	db, _ := utils.SetupMockDB(t)
	defer db.Close()

	minio := utils.SetupMockMinIO(t)
	service := NewVideoService(db, minio, "test-video")

	file, err := os.Open("../../utils/testFiles/test_video.mp4")
	if err != nil {
		t.Fatalf("failed to open file: %v", err)
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", filepath.Base("../testFiles/test_video.mp4"))
	if err != nil {
		t.Fatalf("failed to create form file: %v", err)
	}

	if _, err := io.Copy(part, file); err != nil {
		t.Fatalf("failed to copy file content: %v", err)
	}

	writer.Close()

	r := httptest.NewRequest(http.MethodPost, "/upload", body)
	r.Header.Set("Content-Type", writer.FormDataContentType())

	reqfile, handler, err := r.FormFile("file")
	if err != nil {
		t.Fatalf("failed to get form file: %v", err)
	}
	defer reqfile.Close()

	err = service.CreateVideoFile(reqfile, handler, "users/test-user/test_video.mp4")
	assert.NoError(t, err)
}

func TestGetVideoByID(t *testing.T) {
	db, mock := utils.SetupMockDB(t)
	defer db.Close()

	minio := utils.SetupMockMinIO(t)
	service := NewVideoService(db, minio, "test-video")

	rows := sqlmock.NewRows([]string{"id", "uri", "displayed_name", "filename", "extension"}).
    AddRow(1, "video_uri", "My Video", "movie.mp4", "mp4")

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, uri, displayed_name, filename, extension
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

func TestUpdateVideo(t *testing.T) {
	db, mock := utils.SetupMockDB(t)
	defer db.Close()

	minio := utils.SetupMockMinIO(t)
	service := NewVideoService(db, minio, "test-video")

	video := &modelsMedia.VideoMedia{
		ID:           1,
		URI:          "video_uri_updated",
		DisplayedName: "Updated Name",
		Filename:     "movie_updated.mp4",
		Extension:    "mp4",
	}

	mock.ExpectExec(regexp.QuoteMeta(`
		UPDATE "VideoMedia"
		SET uri = $1, displayed_name = $2, filename = $3, extension = $4
		WHERE id = $5;
	`)).
    WithArgs(video.URI, video.DisplayedName, video.Filename, video.Extension, video.ID).
    WillReturnResult(sqlmock.NewResult(1, 1))

	err := service.UpdateVideo(video)
	assert.NoError(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateVideoFileService(t *testing.T) {
	db, _ := utils.SetupMockDB(t)
	defer db.Close()

	minio := utils.SetupMockMinIO(t)
	service := NewVideoService(db, minio, "test-video")

	file, err := os.Open("../../utils/testFiles/test_video.mp4")
	if err != nil {
		t.Fatalf("failed to open file: %v", err)
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", filepath.Base("../testFiles/test_video.mp4"))
	if err != nil {
		t.Fatalf("failed to create form file: %v", err)
	}

	if _, err := io.Copy(part, file); err != nil {
		t.Fatalf("failed to copy file content: %v", err)
	}

	writer.Close()

	r := httptest.NewRequest(http.MethodPost, "/upload", body)
	r.Header.Set("Content-Type", writer.FormDataContentType())

	reqfile, handler, err := r.FormFile("file")
	if err != nil {
		t.Fatalf("failed to get form file: %v", err)
	}
	defer reqfile.Close()

	err = service.UpdateVideoFile(reqfile, handler, "users/test-user/test_video.mp4")
	assert.NoError(t, err)
}

func TestDeleteVideo(t *testing.T) {
	db, mock := utils.SetupMockDB(t)
	defer db.Close()

	minio := utils.SetupMockMinIO(t)
	service := NewVideoService(db, minio, "test-video")

	mock.ExpectExec(regexp.QuoteMeta(`
		DELETE FROM "VideoMedia"
		WHERE id = $1;
	`)).
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := service.DeleteVideo(1)
	assert.NoError(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteVideoFileService(t *testing.T) {
	db, _ := utils.SetupMockDB(t)
	defer db.Close()

	minio := utils.SetupMockMinIO(t)
	service := NewVideoService(db, minio, "test-video")

	err := service.DeleteVideoFile("users/test-user/test_video.mp4")
	assert.NoError(t, err)
}
