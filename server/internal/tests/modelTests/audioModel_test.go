package modelTests

import (
	"database/sql"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"

	"server/internal/models/modules/media"
)

func setupMockDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock database: %v", err)
	}
	return db, mock
}

func TestCreateAudio(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	audio := &media.AudioMedia{
		URI:       "test_uri",
		Filename:  "song.mp3",
		Extension: "mp3",
	}

	mock.ExpectQuery(regexp.QuoteMeta(`
		INSERT INTO "AudioMedia" (uri, filename, extension)
		VALUES ($1, $2, $3)
		RETURNING id;
	`)).
		WithArgs(audio.URI, audio.Filename, audio.Extension).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	err := media.CreateAudio(db, audio)
	assert.NoError(t, err)
	assert.Equal(t, 1, audio.ID)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetAudioByID(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	rows := sqlmock.NewRows([]string{"id", "uri", "filename", "extension"}).
		AddRow(1, "uri_1", "song.mp3", "mp3")

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, uri, filename, extension
		FROM "AudioMedia"
		WHERE id = $1;
	`)).
		WithArgs(1).
		WillReturnRows(rows)

	audio, err := media.GetAudioByID(db, 1)
	assert.NoError(t, err)
	assert.NotNil(t, audio)
	assert.Equal(t, "song.mp3", audio.Filename)
	assert.Equal(t, "mp3", audio.Extension)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateAudio(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	audio := &media.AudioMedia{
		ID:        1,
		URI:       "uri_new",
		Filename:  "updated.mp3",
		Extension: "mp3",
	}

	mock.ExpectExec(regexp.QuoteMeta(`
		UPDATE "AudioMedia"
		SET uri = $1, filename = $2, extension = $3
		WHERE id = $4;
	`)).
		WithArgs(audio.URI, audio.Filename, audio.Extension, audio.ID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := media.UpdateAudio(db, audio)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteAudio(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	mock.ExpectExec(regexp.QuoteMeta(`
		DELETE FROM "AudioMedia"
		WHERE id = $1;
	`)).
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := media.DeleteAudio(db, 1)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetMediaStats(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	rows := sqlmock.NewRows([]string{"audio_count", "video_count", "photo_count"}).
		AddRow(3, 5, 7)

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT
			(SELECT COUNT(*) FROM "AudioMedia") AS audio_count,
			(SELECT COUNT(*) FROM "VideoMedia") AS video_count,
			(SELECT COUNT(*) FROM "PhotoMedia") AS photo_count;
	`)).
		WillReturnRows(rows)

	stats, err := media.GetMediaStats(db)
	assert.NoError(t, err)
	assert.Equal(t, 3, stats.AudioCount)
	assert.Equal(t, 5, stats.VideoCount)
	assert.Equal(t, 7, stats.PhotoCount)
	assert.Equal(t, 15, stats.Total)

	assert.NoError(t, mock.ExpectationsWereMet())
}
