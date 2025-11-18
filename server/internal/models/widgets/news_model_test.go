package widgets

import (
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

// CREATE

func TestCreateNewsWidget(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	w := &NewsWidget{
		CATEGORY: "Animals",
		LANGUAGE: "en",
		COUNTRY:  "us",
	}

	mock.ExpectQuery(regexp.QuoteMeta(`
		INSERT INTO "NewsWidget" (category, language, country)
		VALUES ($1, $2, $3)
		RETURNING id;
	`)).
		WithArgs(w.CATEGORY, w.LANGUAGE, w.COUNTRY).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, category, language, country
		FROM "NewsWidget"
		WHERE id = $1;
	`)).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "category", "language", "country",
		}).AddRow(1, "Animals", "en", "us"))

	err := CreateNews(db, w)
	assert.NoError(t, err)
	assert.Equal(t, 1, w.ID)

	found, err := GetNewsByID(db, 1)
	assert.NoError(t, err)
	assert.NotNil(t, found)
	assert.Equal(t, "Animals", found.CATEGORY)
	assert.Equal(t, "en", found.LANGUAGE)
	assert.Equal(t, "us", found.COUNTRY)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateNewsBulkWidgets(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	widgets := []NewsWidget{
		{CATEGORY: "Animals", LANGUAGE: "en", COUNTRY: "us"},
		{CATEGORY: "Charity", LANGUAGE: "fr", COUNTRY: "ca"},
		{CATEGORY: "Good News", LANGUAGE: "es", COUNTRY: "mx"},
	}

	mock.ExpectBegin()

	mock.ExpectPrepare(regexp.QuoteMeta(`
		INSERT INTO "NewsWidget" (category, language, country)
		VALUES ($1, $2, $3)
		RETURNING id;
	`))

	for i, w := range widgets {
		mock.ExpectQuery(regexp.QuoteMeta(`
			INSERT INTO "NewsWidget" (category, language, country)
			VALUES ($1, $2, $3)
			RETURNING id;
		`)).
			WithArgs(w.CATEGORY, w.LANGUAGE, w.COUNTRY).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(i + 1))
	}

	mock.ExpectCommit()

	err := CreateNewsBulk(db, widgets)
	assert.NoError(t, err)

	for i, w := range widgets {
		assert.Equal(t, i+1, w.ID)
	}

	assert.NoError(t, mock.ExpectationsWereMet())
}

// RETRIEVE

func TestUpdateNewsWidget(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	w := &NewsWidget{
		ID:       1,
		CATEGORY: "Animals",
		LANGUAGE: "es",
		COUNTRY:  "us",
	}

	mock.ExpectExec(regexp.QuoteMeta(`
		UPDATE "NewsWidget"
		SET category = $1, language = $2, country = $3
		WHERE id = $4;
	`)).
		WithArgs(w.CATEGORY, w.LANGUAGE, w.COUNTRY, w.ID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, category, language, country
		FROM "NewsWidget"
		WHERE id = $1;
	`)).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "category", "language", "country",
		}).AddRow(1, "Animals", "es", "us"))

	err := UpdateNews(db, w)
	assert.NoError(t, err)

	found, err := GetNewsByID(db, 1)
	assert.NoError(t, err)
	assert.Equal(t, "es", found.LANGUAGE)

	assert.NoError(t, mock.ExpectationsWereMet())
}

// DELETE

func TestDeleteNewsWidget(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	mock.ExpectExec(regexp.QuoteMeta(`
		DELETE FROM "NewsWidget"
		WHERE id = $1;
	`)).
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, category, language, country
		FROM "NewsWidget"
		WHERE id = $1;
	`)).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{}))

	err := DeleteNews(db, 1)
	assert.NoError(t, err)

	found, err := GetNewsByID(db, 1)
	assert.NoError(t, err)
	assert.Nil(t, found)

	assert.NoError(t, mock.ExpectationsWereMet())
}

// ANALYSIS/STATS

func TestGetNewsStatsWidget(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT COUNT(*)
		FROM "NewsWidget";
	`)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(3))

	count, err := GetNewsStats(db)
	assert.NoError(t, err)
	assert.Equal(t, 3, count)

	assert.NoError(t, mock.ExpectationsWereMet())
}
