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

func TestCreatePhoto(t *testing.T) {
	db, mock := utils.SetupMockDB(t)
	defer db.Close()

	photo := &PhotoMedia{
		URI:           "https://example.com/photo1.jpg",
		DisplayedName: "My Photo",
		Filename:      "photo1.jpg",
		Extension:     "jpg",
	}

	mock.ExpectQuery(regexp.QuoteMeta(`
		INSERT INTO "PhotoMedia" (uri, displayed_name, filename, extension)
		VALUES ($1, $2, $3, $4)
		RETURNING id;
	`)).
		WithArgs(photo.URI, photo.DisplayedName, photo.Filename, photo.Extension).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	err := CreatePhoto(db, photo)
	assert.NoError(t, err)
	assert.Equal(t, 1, photo.ID)

	assert.NoError(t, mock.ExpectationsWereMet())
}

//
// READ
//

func TestGetPhotoByID(t *testing.T) {
	db, mock := utils.SetupMockDB(t)
	defer db.Close()

	rows := sqlmock.NewRows([]string{
		"id", "uri", "displayed_name", "filename", "extension",
	}).AddRow(
		1,
		"https://example.com/photo1.jpg",
		"My Photo",
		"photo1.jpg",
		"jpg",
	)

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, uri, displayed_name, filename, extension
		FROM "PhotoMedia"
		WHERE id = $1;
	`)).
		WithArgs(1).
		WillReturnRows(rows)

	photo, err := GetPhotoByID(db, 1)
	assert.NoError(t, err)
	assert.NotNil(t, photo)
	assert.Equal(t, "photo1.jpg", photo.Filename)
	assert.Equal(t, "My Photo", photo.DisplayedName)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetPhotosByCarepartnerAndCaregiver(t *testing.T) {
	db, mock := utils.SetupMockDB(t)
	defer db.Close()

	rows := sqlmock.NewRows([]string{
		"id", "uri", "displayed_name", "filename", "extension",
	}).
		AddRow(1, "uri1", "Photo 1", "p1.jpg", "jpg").
		AddRow(2, "uri2", "Photo 2", "p2.jpg", "jpg")

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT p.id, p.uri, p.displayed_name, p.filename, p.extension
		FROM "PhotoMedia" p
		JOIN "Module" m ON m.metadata_id = p.id
		WHERE m.carepartner_id = $1 AND m.caregiver_id = $2;
	`)).
		WithArgs(201, 301).
		WillReturnRows(rows)

	photos, err := GetPhotosByCarepartnerAndCaregiver(db, 201, 301)
	assert.NoError(t, err)
	assert.Len(t, photos, 2)
	assert.Equal(t, "p2.jpg", photos[1].Filename)

	assert.NoError(t, mock.ExpectationsWereMet())
}

//
// UPDATE
//

func TestUpdatePhoto(t *testing.T) {
	db, mock := utils.SetupMockDB(t)
	defer db.Close()

	photo := &PhotoMedia{
		ID:            1,
		URI:           "updated_uri",
		DisplayedName: "Updated Photo",
		Filename:      "updated.jpg",
		Extension:     "jpg",
	}

	mock.ExpectExec(regexp.QuoteMeta(`
		UPDATE "PhotoMedia"
		SET uri = $1,
		    displayed_name = $2,
		    filename = $3,
		    extension = $4
		WHERE id = $5;
	`)).
		WithArgs(photo.URI, photo.DisplayedName, photo.Filename, photo.Extension, photo.ID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := UpdatePhoto(db, photo)
	assert.NoError(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdatePhoto_NotFound(t *testing.T) {
	db, mock := utils.SetupMockDB(t)
	defer db.Close()

	photo := &PhotoMedia{
		ID:            99,
		URI:           "missing",
		DisplayedName: "Missing",
		Filename:      "missing.jpg",
		Extension:     "jpg",
	}

	mock.ExpectExec(regexp.QuoteMeta(`
		UPDATE "PhotoMedia"
		SET uri = $1,
		    displayed_name = $2,
		    filename = $3,
		    extension = $4
		WHERE id = $5;
	`)).
		WithArgs(photo.URI, photo.DisplayedName, photo.Filename, photo.Extension, photo.ID).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err := UpdatePhoto(db, photo)
	assert.Error(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

//
// DELETE
//

func TestDeletePhoto(t *testing.T) {
	db, mock := utils.SetupMockDB(t)
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

func TestDeletePhoto_NotFound(t *testing.T) {
	db, mock := utils.SetupMockDB(t)
	defer db.Close()

	mock.ExpectExec(regexp.QuoteMeta(`
		DELETE FROM "PhotoMedia"
		WHERE id = $1;
	`)).
		WithArgs(99).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err := DeletePhoto(db, 99)
	assert.Error(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

//
// STATS
//

func TestGetPhotoStats(t *testing.T) {
	db, mock := utils.SetupMockDB(t)
	defer db.Close()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT COUNT(*) FROM "PhotoMedia";`)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(10))

	count, err := GetPhotoStats(db)
	assert.NoError(t, err)
	assert.Equal(t, 10, count)

	assert.NoError(t, mock.ExpectationsWereMet())
}
