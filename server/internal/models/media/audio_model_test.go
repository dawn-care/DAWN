package media

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"

	"server/internal/utils"
)

func TestCreateAudio(t *testing.T) {
	db, mock := utils.SetupMockDB(t)
	defer db.Close()

	audio := &AudioMedia{
		URI:           "test_uri",
		Filename:      "song.mp3",
		Extension:     "mp3",
		DisplayedName: "Test Song",
	}

	mock.ExpectQuery(regexp.QuoteMeta(`
		INSERT INTO "AudioMedia" (uri, filename, extension, displayed_name)
		VALUES ($1, $2, $3, $4)
		RETURNING id;
	`)).
		WithArgs(audio.URI, audio.Filename, audio.Extension, audio.DisplayedName).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	err := CreateAudio(db, audio)
	assert.NoError(t, err)
	assert.Equal(t, 1, audio.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetAudioByID(t *testing.T) {
	db, mock := utils.SetupMockDB(t)
	defer db.Close()

	rows := sqlmock.NewRows([]string{
		"id", "uri", "filename", "extension", "displayed_name",
	}).AddRow(1, "uri_1", "song.mp3", "mp3", "My Song")

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, uri, filename, extension, displayed_name
		FROM "AudioMedia"
		WHERE id = $1;
	`)).
		WithArgs(1).
		WillReturnRows(rows)

	audio, err := GetAudioByID(db, 1)
	assert.NoError(t, err)
	assert.NotNil(t, audio)
	assert.Equal(t, "song.mp3", audio.Filename)
	assert.Equal(t, "mp3", audio.Extension)
	assert.Equal(t, "My Song", audio.DisplayedName)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetAudiosByCarepartnerAndCaregiver(t *testing.T) {
	db, mock := utils.SetupMockDB(t)
	defer db.Close()

	rows := sqlmock.NewRows([]string{
		"id", "uri", "filename", "extension", "displayed_name",
	}).AddRow(1, "uri", "file1.mp3", "mp3", "File 1")

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT a.id, a.uri, a.filename, a.extension, a.displayed_name
		FROM "AudioMedia" a
		JOIN "Module" m ON m.metadata_id = a.id
		WHERE m.carepartner_id = $1 AND m.caregiver_id = $2;
	`)).
		WithArgs(2, 1).
		WillReturnRows(rows)

	result, err := GetAudiosByCarepartnerAndCaregiver(db, 2, 1)
	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "file1.mp3", result[0].Filename)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSearchAudios(t *testing.T) {
	db, mock := utils.SetupMockDB(t)
	defer db.Close()

	rows := sqlmock.NewRows([]string{
		"id", "uri", "filename", "extension", "displayed_name",
	}).AddRow(1, "uri", "love.mp3", "mp3", "Love Song")

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, uri, filename, extension, displayed_name
		FROM "AudioMedia"
		WHERE filename ILIKE $1
		   OR displayed_name ILIKE $1;
	`)).
		WithArgs("%love%").
		WillReturnRows(rows)

	results, err := SearchAudios(db, "love")
	assert.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "love.mp3", results[0].Filename)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetMediaStats(t *testing.T) {
	db, mock := utils.SetupMockDB(t)
	defer db.Close()

	rows := sqlmock.NewRows([]string{
		"audio_count", "video_count", "photo_count",
	}).AddRow(3, 5, 7)

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT
			(SELECT COUNT(*) FROM "AudioMedia") AS audio_count,
			(SELECT COUNT(*) FROM "VideoMedia") AS video_count,
			(SELECT COUNT(*) FROM "PhotoMedia") AS photo_count;
	`)).
		WillReturnRows(rows)

	stats, err := GetMediaStats(db)
	assert.NoError(t, err)
	assert.Equal(t, 3, stats.AudioCount)
	assert.Equal(t, 5, stats.VideoCount)
	assert.Equal(t, 7, stats.PhotoCount)
	assert.Equal(t, 15, stats.Total)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteAudio_NotFound(t *testing.T) {
	db, mock := utils.SetupMockDB(t)
	defer db.Close()

	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "AudioMedia" WHERE id = $1;`)).
		WithArgs(99).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err := DeleteAudio(db, 99)
	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetMediaByType_InvalidType(t *testing.T) {
	db, _ := utils.SetupMockDB(t)
	defer db.Close()

	_, err := GetMediaByTypeCarepartnerAndCaregiver(db, "invalid", 1, 1)
	assert.Error(t, err)
	assert.Equal(t, fmt.Sprintf("invalid media type: %s", "invalid"), err.Error())
}
