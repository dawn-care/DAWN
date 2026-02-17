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

func TestCreateAudioService(t *testing.T) {
	db, mock := utils.SetupMockDB(t)
	defer db.Close()

	minio := utils.SetupMockMinIO(t)
	service := NewAudioService(db, minio, "test-audio")

	audio := &modelsMedia.AudioMedia{
		URI:           "users/test/test.mp3",
		Filename:      "test",
		Extension:     "mp3",
		DisplayedName: "Test Audio",
	}

	mock.ExpectBegin()

	mock.ExpectQuery(regexp.QuoteMeta(`
		INSERT INTO "AudioMedia" (uri, filename, extension, displayed_name)
		VALUES ($1, $2, $3, $4)
		RETURNING id;
	`)).
		WithArgs(audio.URI, audio.Filename, audio.Extension, audio.DisplayedName).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	mock.ExpectCommit()

	err := service.CreateAudio(audio)

	assert.NoError(t, err)
	assert.Equal(t, 1, audio.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateAudioService_NilAudio(t *testing.T) {
	db, _ := utils.SetupMockDB(t)
	defer db.Close()

	minio := utils.SetupMockMinIO(t)
	service := NewAudioService(db, minio, "test-audio")

	err := service.CreateAudio(nil)
	assert.EqualError(t, err, "audio cannot be nil")
}

func TestCreateAudioService_InvalidAudio(t *testing.T) {
	db, _ := utils.SetupMockDB(t)
	defer db.Close()

	minio := utils.SetupMockMinIO(t)
	service := NewAudioService(db, minio, "test-audio")

	audio := &modelsMedia.AudioMedia{}
	err := service.CreateAudio(audio)

	assert.EqualError(t, err, "uri, filename, and extension are required")
}

func TestCreateAudioFileService(t *testing.T) {
	db, _ := utils.SetupMockDB(t)
	defer db.Close()

	minio := utils.SetupMockMinIO(t)
	service := NewAudioService(db, minio, "test-audio")

	// Fake in-memory file instead of disk dependency
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", "test.mp3")
	assert.NoError(t, err)

	_, err = io.Copy(part, bytes.NewBuffer([]byte("fake-audio-data")))
	assert.NoError(t, err)

	writer.Close()

	r := httptest.NewRequest(http.MethodPost, "/upload", body)
	r.Header.Set("Content-Type", writer.FormDataContentType())

	err = r.ParseMultipartForm(10 << 20)
	assert.NoError(t, err)

	reqFile, handler, err := r.FormFile("file")
	assert.NoError(t, err)
	defer reqFile.Close()

	serviceErr := service.CreateAudioFile(reqFile, handler, "users/test-user/test.mp3")
	assert.NoError(t, serviceErr)
}
