package media

import (
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

// CREATE

func TestCreateVideo(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	video := &VideoMedia{
		URI:       "test_uri",
		Filename:  "video.mp4",
		Extension: "mp4",
	}

	mock.ExpectQuery(regexp.QuoteMeta(`
		INSERT INTO "VideoMedia" (uri, filename, extension)
		VALUES ($1, $2, $3)
		RETURNING id;
	`)).
		WithArgs(video.URI, video.Filename, video.Extension).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	err := CreateVideo(db, video)
	assert.NoError(t, err)
	assert.Equal(t, 1, video.ID)

	assert.NoError(t, mock.ExpectationsWereMet())
}

// RETRIEVE

func TestGetVideoByID(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	rows := sqlmock.NewRows([]string{"id", "uri", "filename", "extension"}).
		AddRow(1, "uri_1", "video.mp4", "mp4")

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, uri, filename, extension
		FROM "VideoMedia"
		WHERE id = $1;
	`)).
		WithArgs(1).
		WillReturnRows(rows)

	video, err := GetVideoByID(db, 1)
	assert.NoError(t, err)
	assert.NotNil(t, video)
	assert.Equal(t, "video.mp4", video.Filename)
	assert.Equal(t, "mp4", video.Extension)

	assert.NoError(t, mock.ExpectationsWereMet())
}

// UPDATE

func TestUpdateVideo(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	video := &VideoMedia{
		ID:        1,
		URI:       "new_uri",
		Filename:  "updated.mp4",
		Extension: "mkv",
	}

	mock.ExpectExec(regexp.QuoteMeta(`
		UPDATE "VideoMedia"
		SET uri = $1, filename = $2, extension = $3
		WHERE id = $4;
	`)).
		WithArgs(video.URI, video.Filename, video.Extension, video.ID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := UpdateVideo(db, video)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// DELETE

func TestDeleteVideo(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	mock.ExpectExec(regexp.QuoteMeta(`
		DELETE FROM "VideoMedia"
		WHERE id = $1;
	`)).
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := DeleteVideo(db, 1)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ANALYSIS & STATS

func TestGetVideoStats(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT COUNT(*) FROM "VideoMedia";
	`)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))

	count, err := GetVideoStats(db)
	assert.NoError(t, err)
	assert.Equal(t, 5, count)

	assert.NoError(t, mock.ExpectationsWereMet())
}
