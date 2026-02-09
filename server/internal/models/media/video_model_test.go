package media

import (
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"

	"server/internal/utils"
)

//
// CREATE
//

func TestCreateVideo(t *testing.T) {
	db, mock := utils.SetupMockDB(t)
	defer db.Close()

	video := &VideoMedia{
		URI:           "test_uri",
		DisplayedName: "Test Video",
		Filename:      "video.mp4",
		Extension:     "mp4",
	}

	mock.ExpectQuery(regexp.QuoteMeta(`
		INSERT INTO "VideoMedia" (uri, displayed_name, filename, extension)
		VALUES ($1, $2, $3, $4)
		RETURNING id;
	`)).
		WithArgs(video.URI, video.DisplayedName, video.Filename, video.Extension).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	err := CreateVideo(db, video)
	assert.NoError(t, err)
	assert.Equal(t, 1, video.ID)

	assert.NoError(t, mock.ExpectationsWereMet())
}

//
// READ — By ID
//

func TestGetVideoByID(t *testing.T) {
	db, mock := utils.SetupMockDB(t)
	defer db.Close()

	rows := sqlmock.NewRows([]string{
		"id", "uri", "displayed_name", "filename", "extension",
	}).AddRow(
		1,
		"uri_1",
		"My Video",
		"video.mp4",
		"mp4",
	)

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, uri, displayed_name, filename, extension
		FROM "VideoMedia"
		WHERE id = $1;
	`)).
		WithArgs(1).
		WillReturnRows(rows)

	video, err := GetVideoByID(db, 1)
	assert.NoError(t, err)
	assert.NotNil(t, video)
	assert.Equal(t, "My Video", video.DisplayedName)

	assert.NoError(t, mock.ExpectationsWereMet())
}

//
// READ — By Carepartner & Caregiver (JOIN)
//

func TestGetVideosByCarepartnerAndCaregiver(t *testing.T) {
	db, mock := utils.SetupMockDB(t)
	defer db.Close()

	rows := sqlmock.NewRows([]string{
		"id", "uri", "displayed_name", "filename", "extension",
	}).
		AddRow(1, "uri1", "Video 1", "v1.mp4", "mp4").
		AddRow(2, "uri2", "Video 2", "v2.mp4", "mp4")

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT v.id, v.uri, v.displayed_name, v.filename, v.extension
		FROM "VideoMedia" v
		JOIN "Module" m ON m.metadata_id = v.id
		WHERE m.carepartner_id = $1 AND m.caregiver_id = $2;
	`)).
		WithArgs(201, 301).
		WillReturnRows(rows)

	videos, err := GetVideosByCarepartnerAndCaregiver(db, 201, 301)
	assert.NoError(t, err)
	assert.Len(t, videos, 2)
	assert.Equal(t, "v2.mp4", videos[1].Filename)

	assert.NoError(t, mock.ExpectationsWereMet())
}

//
// UPDATE
//

func TestUpdateVideo(t *testing.T) {
	db, mock := utils.SetupMockDB(t)
	defer db.Close()

	video := &VideoMedia{
		ID:            1,
		URI:           "new_uri",
		DisplayedName: "Updated Video",
		Filename:      "updated.mp4",
		Extension:     "mkv",
	}

	mock.ExpectExec(regexp.QuoteMeta(`
		UPDATE "VideoMedia"
		SET uri = $1,
		    displayed_name = $2,
		    filename = $3,
		    extension = $4
		WHERE id = $5;
	`)).
		WithArgs(video.URI, video.DisplayedName, video.Filename, video.Extension, video.ID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := UpdateVideo(db, video)
	assert.NoError(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateVideo_NotFound(t *testing.T) {
	db, mock := utils.SetupMockDB(t)
	defer db.Close()

	video := &VideoMedia{
		ID:            99,
		URI:           "missing",
		DisplayedName: "Missing",
		Filename:      "missing.mp4",
		Extension:     "mp4",
	}

	mock.ExpectExec(regexp.QuoteMeta(`
		UPDATE "VideoMedia"
		SET uri = $1,
		    displayed_name = $2,
		    filename = $3,
		    extension = $4
		WHERE id = $5;
	`)).
		WithArgs(video.URI, video.DisplayedName, video.Filename, video.Extension, video.ID).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err := UpdateVideo(db, video)
	assert.Error(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

//
// DELETE
//

func TestDeleteVideo(t *testing.T) {
	db, mock := utils.SetupMockDB(t)
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

func TestDeleteVideo_NotFound(t *testing.T) {
	db, mock := utils.SetupMockDB(t)
	defer db.Close()

	mock.ExpectExec(regexp.QuoteMeta(`
		DELETE FROM "VideoMedia"
		WHERE id = $1;
	`)).
		WithArgs(99).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err := DeleteVideo(db, 99)
	assert.Error(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

//
// STATS
//

func TestGetVideoStats(t *testing.T) {
	db, mock := utils.SetupMockDB(t)
	defer db.Close()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT COUNT(*) FROM "VideoMedia";`)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))

	count, err := GetVideoStats(db)
	assert.NoError(t, err)
	assert.Equal(t, 5, count)

	assert.NoError(t, mock.ExpectationsWereMet())
}
