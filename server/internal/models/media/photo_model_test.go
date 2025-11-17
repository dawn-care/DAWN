package media

import (
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestCreatePhoto(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	photo := &PhotoMedia{
		URI:       "https://example.com/photo1.jpg",
		Filename:  "photo1.jpg",
		Extension: "jpg",
	}

	mock.ExpectQuery(regexp.QuoteMeta(`
		INSERT INTO "PhotoMedia" (uri, filename, extension)
		VALUES ($1, $2, $3)
		RETURNING id;
	`)).
		WithArgs(photo.URI, photo.Filename, photo.Extension).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	err := CreatePhoto(db, photo)
	assert.NoError(t, err)
	assert.Equal(t, 1, photo.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreatePhotosBulk(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	photos := []PhotoMedia{
		{URI: "https://example.com/photo1.jpg", Filename: "photo1.jpg", Extension: "jpg"},
		{URI: "https://example.com/photo2.jpg", Filename: "photo2.jpg", Extension: "jpg"},
	}

	mock.ExpectBegin()
	mock.ExpectPrepare(regexp.QuoteMeta(`
		INSERT INTO "PhotoMedia" (uri, filename, extension)
		VALUES ($1, $2, $3)
		RETURNING id
	`))

	for i := range photos {
		mock.ExpectQuery(regexp.QuoteMeta(`
			INSERT INTO "PhotoMedia" (uri, filename, extension)
			VALUES ($1, $2, $3)
			RETURNING id
		`)).
			WithArgs(photos[i].URI, photos[i].Filename, photos[i].Extension).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(i + 1))
	}
	mock.ExpectCommit()

	err := CreatePhotosBulk(db, photos)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetPhotoByID(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	rows := sqlmock.NewRows([]string{"id", "uri", "filename", "extension"}).
		AddRow(1, "https://example.com/photo1.jpg", "photo1.jpg", "jpg")

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, uri, filename, extension
		FROM "PhotoMedia"
		WHERE id = $1;
	`)).
		WithArgs(1).
		WillReturnRows(rows)

	photo, err := GetPhotoByID(db, 1)
	assert.NoError(t, err)
	assert.NotNil(t, photo)
	assert.Equal(t, "photo1.jpg", photo.Filename)
	assert.Equal(t, "jpg", photo.Extension)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetPhotosByCarepartnerAndCaregiver(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	rows := sqlmock.NewRows([]string{"id", "uri", "filename", "extension"}).
		AddRow(1, "https://example.com/photo1.jpg", "photo1.jpg", "jpg").
		AddRow(2, "https://example.com/photo2.jpg", "photo2.jpg", "jpg")

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT p.id, p.uri, p.filename, p.extension
		FROM "PhotoMedia" p
		JOIN "Module" m ON m.metadata_id = p.id
		WHERE m.carepartner_id = $1 AND m.caregiver_id = $2;
	`)).
		WithArgs(201, 301).
		WillReturnRows(rows)

	photos, err := GetPhotosByCarepartnerAndCaregiver(db, 201, 301)
	assert.NoError(t, err)
	assert.Len(t, photos, 2)
	assert.Equal(t, "photo2.jpg", photos[1].Filename)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdatePhoto(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	photo := &PhotoMedia{
		ID:        1,
		URI:       "https://example.com/photo1_updated.jpg",
		Filename:  "photo1_updated.jpg",
		Extension: "jpg",
	}

	mock.ExpectExec(regexp.QuoteMeta(`
		UPDATE "PhotoMedia"
		SET uri = $1, filename = $2, extension = $3
		WHERE id = $4;
	`)).
		WithArgs(photo.URI, photo.Filename, photo.Extension, photo.ID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := UpdatePhoto(db, photo)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeletePhoto(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	mock.ExpectExec(regexp.QuoteMeta(`
		DELETE FROM "PhotoMedia"
		WHERE id = $1;
	`)).
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := DeletePhoto(db, 1)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetPhotoStats(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT COUNT(*) FROM "PhotoMedia";
	`)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(10))

	count, err := GetPhotoStats(db)
	assert.NoError(t, err)
	assert.Equal(t, 10, count)
	assert.NoError(t, mock.ExpectationsWereMet())
}
