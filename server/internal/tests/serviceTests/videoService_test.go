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

func TestCreateVideo(t *testing.T) {
	db, mock := tests.SetupMockDB(t)
	defer db.Close()

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

	service := servicesMedia.NewVideoService(db)
	err := service.CreateVideo(video)
	assert.NoError(t, err)
	assert.Equal(t, 1, video.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetVideoByID(t *testing.T) {
	db, mock := tests.SetupMockDB(t)
	defer db.Close()

	rows := sqlmock.NewRows([]string{"id", "uri", "filename", "extension"}).
		AddRow(1, "video_uri", "movie.mp4", "mp4")

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, uri, filename, extension
		FROM "VideoMedia"
		WHERE id = $1;
	`)).
		WithArgs(1).
		WillReturnRows(rows)

	service := servicesMedia.NewVideoService(db)
	video, err := service.GetVideoByID(1)
	assert.NoError(t, err)
	assert.NotNil(t, video)
	assert.Equal(t, "movie.mp4", video.Filename)
	assert.Equal(t, "mp4", video.Extension)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateVideo(t *testing.T) {
	db, mock := tests.SetupMockDB(t)
	defer db.Close()

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

	service := servicesMedia.NewVideoService(db)
	err := service.UpdateVideo(video)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteVideo(t *testing.T) {
	db, mock := tests.SetupMockDB(t)
	defer db.Close()

	mock.ExpectExec(regexp.QuoteMeta(`
		DELETE FROM "VideoMedia"
		WHERE id = $1;
	`)).
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(1, 1))

	service := servicesMedia.NewVideoService(db)
	err := service.DeleteVideo(1)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetVideoStats(t *testing.T) {
	db, mock := tests.SetupMockDB(t)
	defer db.Close()

	rows := sqlmock.NewRows([]string{"count"}).AddRow(5)

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT COUNT(*) FROM "VideoMedia";
	`)).WillReturnRows(rows)

	service := servicesMedia.NewVideoService(db)
	count, err := service.GetVideoStats()
	assert.NoError(t, err)
	assert.Equal(t, 5, count)
	assert.NoError(t, mock.ExpectationsWereMet())
}
