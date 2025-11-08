package mediaTests

import (
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"

	"server/internal/tests"
	modelsMedia "server/internal/models/modules/media"
	servicesMedia "server/internal/services/modules/media"
)

func TestCreateAudioService(t *testing.T) {
	db, mock := tests.SetupMockDB(t)
	defer db.Close()

	service := servicesMedia.NewAudioService(db)

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

func TestCreateAudioService_NilAudio(t *testing.T) {
	db, _ := tests.SetupMockDB(t)
	defer db.Close()
	service := servicesMedia.NewAudioService(db)

	err := service.CreateAudio(nil)
	assert.EqualError(t, err, "audio cannot be nil")
}

func TestCreateAudioService_InvalidAudio(t *testing.T) {
	db, _ := tests.SetupMockDB(t)
	defer db.Close()
	service := servicesMedia.NewAudioService(db)

	audio := &modelsMedia.AudioMedia{}
	err := service.CreateAudio(audio)
	assert.EqualError(t, err, "uri, filename, and extension are required")
}

func TestGetAudioByIDService(t *testing.T) {
	db, mock := tests.SetupMockDB(t)
	defer db.Close()

	service := servicesMedia.NewAudioService(db)

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

func TestGetAudioByIDService_InvalidID(t *testing.T) {
	db, _ := tests.SetupMockDB(t)
	defer db.Close()
	service := servicesMedia.NewAudioService(db)

	_, err := service.GetAudioByID(0)
	assert.EqualError(t, err, "invalid audio ID")
}

func TestUpdateAudioService(t *testing.T) {
	db, mock := tests.SetupMockDB(t)
	defer db.Close()
	service := servicesMedia.NewAudioService(db)

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

func TestUpdateAudioService_InvalidAudio(t *testing.T) {
	db, _ := tests.SetupMockDB(t)
	defer db.Close()
	service := servicesMedia.NewAudioService(db)

	err := service.UpdateAudio(nil)
	assert.EqualError(t, err, "audio cannot be nil")

	err = service.UpdateAudio(&modelsMedia.AudioMedia{ID: 0})
	assert.EqualError(t, err, "invalid audio ID")
}

func TestDeleteAudioService(t *testing.T) {
	db, mock := tests.SetupMockDB(t)
	defer db.Close()
	service := servicesMedia.NewAudioService(db)

	mock.ExpectExec(regexp.QuoteMeta(
		`DELETE FROM "AudioMedia" WHERE id = $1;`,
	)).
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := service.DeleteAudio(1)
	assert.NoError(t, err)
}

func TestDeleteAudioService_InvalidID(t *testing.T) {
	db, _ := tests.SetupMockDB(t)
	defer db.Close()
	service := servicesMedia.NewAudioService(db)

	err := service.DeleteAudio(0)
	assert.EqualError(t, err, "invalid audio ID")
}

func TestGetMediaStatsService(t *testing.T) {
	db, mock := tests.SetupMockDB(t)
	defer db.Close()
	service := servicesMedia.NewAudioService(db)

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
	service := servicesMedia.NewAudioService(db)

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
	service := servicesMedia.NewAudioService(db)

	_, err := service.SearchAudios("")
	assert.EqualError(t, err, "keyword cannot be empty")
}
