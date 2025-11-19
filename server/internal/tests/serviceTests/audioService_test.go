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

func TestCreateAudioService(t *testing.T) {
	db, mock := tests.SetupMockDB(t)
	defer db.Close()

	minio := tests.SetupMockMinIO(t)

	service := servicesMedia.NewAudioService(db, minio, "test-audio")

	audio := &modelsMedia.AudioMedia{
		URI:       "https://example.com/test.mp3",
		Filename:  "test",
		Extension: "mp3",
	}

	mock.ExpectQuery(regexp.QuoteMeta(
		`INSERT INTO "AudioMedia" (uri, filename, extension)
		VALUES ($1, $2, $3)
		RETURNING id;`,
	)).
		WithArgs(audio.URI, audio.Filename, audio.Extension).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	err := service.CreateAudio(audio)
	assert.NoError(t, err)
	assert.Equal(t, 1, audio.ID)
}

func TestCreateAudioFileService(t *testing.T) {
	db, _ := tests.SetupMockDB(t)
	defer db.Close()

	minio := tests.SetupMockMinIO(t)

	service := servicesMedia.NewAudioService(db, minio, "test-audio")

	file, err := os.Open("..\\testFiles\\test_audio.mp3")
	if err != nil {
		t.Fatalf("failed to open file: %v", err)
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", filepath.Base("server\\internal\\tests\\testFiles\\test_audio.mp3"))
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

	serviceErr := service.CreateAudioFile(reqfile, handler, "users/test-user/test_audio.mp3")
	assert.NoError(t, serviceErr)
}

func TestCreateAudioService_NilAudio(t *testing.T) {
	db, _ := tests.SetupMockDB(t)
	defer db.Close()
	
	minio := tests.SetupMockMinIO(t)

	service := servicesMedia.NewAudioService(db, minio, "test-audio")

	err := service.CreateAudio(nil)
	assert.EqualError(t, err, "audio cannot be nil")
}

func TestCreateAudioService_InvalidAudio(t *testing.T) {
	db, _ := tests.SetupMockDB(t)
	defer db.Close()

	minio := tests.SetupMockMinIO(t)

	service := servicesMedia.NewAudioService(db, minio, "test-audio")

	audio := &modelsMedia.AudioMedia{}
	err := service.CreateAudio(audio)
	assert.EqualError(t, err, "uri, filename, and extension are required")
}

func TestGetAudioByIDService(t *testing.T) {
	db, mock := tests.SetupMockDB(t)
	defer db.Close()

	minio := tests.SetupMockMinIO(t)

	service := servicesMedia.NewAudioService(db, minio, "test-audio")

	expectedAudio := &modelsMedia.AudioMedia{
		ID:        1,
		URI:       "https://example.com/test.mp3",
		Filename:  "test",
		Extension: "mp3",
	}

	mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT id, uri, filename, extension
		FROM "AudioMedia"
		WHERE id = $1;`,
	)).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "uri", "filename", "extension"}).
			AddRow(expectedAudio.ID, expectedAudio.URI, expectedAudio.Filename, expectedAudio.Extension))

	audio, err := service.GetAudioByID(1)
	assert.NoError(t, err)
	assert.Equal(t, expectedAudio, audio)
}

func TestGetAudioFileService(t *testing.T) {
	db, _ := tests.SetupMockDB(t)
	defer db.Close()

	minio := tests.SetupMockMinIO(t)

	service := servicesMedia.NewAudioService(db, minio, "test-audio")

	file, err := service.GetAudioFile("users/test-user/test_audio.mp3")
	assert.NoError(t, err)
	defer file.Close()

	gotBytes, err := io.ReadAll(file)
	assert.NoError(t, err)

	localPath := "../testFiles/test_audio.mp3"
	expectedBytes, err := os.ReadFile(localPath)
	assert.NoError(t, err)

	assert.Equal(t, expectedBytes, gotBytes)
}


func TestGetAudioByIDService_InvalidID(t *testing.T) {
	db, _ := tests.SetupMockDB(t)
	defer db.Close()

	minio := tests.SetupMockMinIO(t)

	service := servicesMedia.NewAudioService(db, minio, "test-audio")

	_, err := service.GetAudioByID(0)
	assert.EqualError(t, err, "invalid audio ID")
}

func TestUpdateAudioService(t *testing.T) {
	db, mock := tests.SetupMockDB(t)
	defer db.Close()

	minio := tests.SetupMockMinIO(t)

	service := servicesMedia.NewAudioService(db, minio, "test-audio")

	audio := &modelsMedia.AudioMedia{
		ID:        1,
		URI:       "https://example.com/test.mp3",
		Filename:  "test",
		Extension: "mp3",
	}

	mock.ExpectExec(regexp.QuoteMeta(
		`UPDATE "AudioMedia"
		SET uri = $1, filename = $2, extension = $3
		WHERE id = $4;`,
	)).
		WithArgs(audio.URI, audio.Filename, audio.Extension, audio.ID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := service.UpdateAudio(audio)
	assert.NoError(t, err)
}

func TestUpdateAudioFileService(t *testing.T) {
	db, _ := tests.SetupMockDB(t)
	defer db.Close()

	minio := tests.SetupMockMinIO(t)

	service := servicesMedia.NewAudioService(db, minio, "test-audio")

	file, err := os.Open("..\\testFiles\\test_audio.mp3")
	if err != nil {
		t.Fatalf("failed to open file: %v", err)
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", filepath.Base("server\\internal\\tests\\testFiles\\test_audio.mp3"))
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

	serviceErr := service.UpdateAudioFile(reqfile, handler, "users/test-user/test_audio.mp3")
	assert.NoError(t, serviceErr)
}

func TestUpdateAudioService_InvalidAudio(t *testing.T) {
	db, _ := tests.SetupMockDB(t)
	defer db.Close()
	minio := tests.SetupMockMinIO(t)

	service := servicesMedia.NewAudioService(db, minio, "test-audio")

	err := service.UpdateAudio(nil)
	assert.EqualError(t, err, "audio cannot be nil")

	err = service.UpdateAudio(&modelsMedia.AudioMedia{ID: 0})
	assert.EqualError(t, err, "invalid audio ID")
}

func TestDeleteAudioService(t *testing.T) {
	db, mock := tests.SetupMockDB(t)
	defer db.Close()

	minio := tests.SetupMockMinIO(t)

	service := servicesMedia.NewAudioService(db, minio, "test-audio")

	mock.ExpectExec(regexp.QuoteMeta(
		`DELETE FROM "AudioMedia" WHERE id = $1;`,
	)).
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := service.DeleteAudio(1)
	assert.NoError(t, err)
}

func TestDeleteAudioFileService(t *testing.T) {
	db, _ := tests.SetupMockDB(t)
	defer db.Close()

	minio := tests.SetupMockMinIO(t)

	service := servicesMedia.NewAudioService(db, minio, "test-audio")

	err := service.DeleteAudioFile("users/test-user/test_audio.mp3")
	assert.NoError(t, err)
}


func TestDeleteAudioService_InvalidID(t *testing.T) {
	db, _ := tests.SetupMockDB(t)
	defer db.Close()

	minio := tests.SetupMockMinIO(t)

	service := servicesMedia.NewAudioService(db, minio, "test-audio")

	err := service.DeleteAudio(0)
	assert.EqualError(t, err, "invalid audio ID")
}

func TestGetMediaStatsService(t *testing.T) {
	db, mock := tests.SetupMockDB(t)
	defer db.Close()

	minio := tests.SetupMockMinIO(t)

	service := servicesMedia.NewAudioService(db, minio, "test-audio")

	mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT
			(SELECT COUNT(*) FROM "AudioMedia") AS audio_count,
			(SELECT COUNT(*) FROM "VideoMedia") AS video_count,
			(SELECT COUNT(*) FROM "PhotoMedia") AS photo_count;`,
	)).WillReturnRows(sqlmock.NewRows([]string{"audio_count", "video_count", "photo_count"}).AddRow(3, 2, 5))

	stats, err := service.GetMediaStats()
	assert.NoError(t, err)
	assert.Equal(t, 3, stats.AudioCount)
	assert.Equal(t, 2, stats.VideoCount)
	assert.Equal(t, 5, stats.PhotoCount)
	assert.Equal(t, 10, stats.Total)
}

func TestSearchAudiosService(t *testing.T) {
	db, mock := tests.SetupMockDB(t)
	defer db.Close()

	minio := tests.SetupMockMinIO(t)

	service := servicesMedia.NewAudioService(db, minio, "test-audio")

	keyword := "test"
	expectedResults := []modelsMedia.AudioMedia{
		{ID: 1, URI: "https://example.com/test.mp3", Filename: "test", Extension: "mp3"},
	}

	mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT id, uri, filename, extension
		FROM "AudioMedia"
		WHERE filename ILIKE $1;`,
	)).WithArgs("%test%").WillReturnRows(sqlmock.NewRows([]string{"id", "uri", "filename", "extension"}).
		AddRow(1, "https://example.com/test.mp3", "test", "mp3"))

	results, err := service.SearchAudios(keyword)
	assert.NoError(t, err)
	assert.Equal(t, expectedResults, results)
}

func TestSearchAudiosService_EmptyKeyword(t *testing.T) {
	db, _ := tests.SetupMockDB(nil)
	defer db.Close()

	minio := tests.SetupMockMinIO(t)

	service := servicesMedia.NewAudioService(db, minio, "test-audio")

	_, err := service.SearchAudios("")
	assert.EqualError(t, err, "keyword cannot be empty")
}
