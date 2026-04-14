package news

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"

	modelsWidget "server/internal/models/widgets/news"

	"server/internal/utils"
)

func TestCreateSportsNewsWidget(t *testing.T) {
	db, mock := utils.SetupMockDB(t)
	defer db.Close()

	news_service := NewSportsNewsWidgetService(db)

	news_widget := &modelsWidget.NewsWidget{
		ID:       1,
		CATEGORY: "sports",
		LANGUAGE: "BUF",
		COUNTRY:  "nfl",
	}

	mock.ExpectQuery(regexp.QuoteMeta(`
		INSERT INTO "NewsWidget" (category, language, country)
		VALUES ($1, $2, $3)
		RETURNING id;
	`)).
		WithArgs("sports", "BUF", "nfl").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	err := news_service.CreateSportsNewsWidget(news_widget)
	assert.NoError(t, err)
	assert.Equal(t, 1, news_widget.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetSportsNewsWidgetByID(t *testing.T) {
	db, mock := utils.SetupMockDB(t)
	defer db.Close()

	news_service := NewSportsNewsWidgetService(db)

	rows := sqlmock.NewRows([]string{"id", "category", "language", "country"}).
		AddRow(1, "sports", "BUF", "nfl")

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, category, language, country
		FROM "NewsWidget"
		WHERE id = $1;
	`)).
		WithArgs(1).
		WillReturnRows(rows)

	news_widget, err := news_service.GetSportsNewsWidgetByID(1)
	assert.NoError(t, err)
	assert.NotNil(t, news_widget)
	assert.Equal(t, 1, news_widget.ID)
	assert.Equal(t, "sports", news_widget.CATEGORY)
	assert.Equal(t, "BUF", news_widget.LANGUAGE)
	assert.Equal(t, "nfl", news_widget.COUNTRY)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateSportsNewsWidget(t *testing.T) {
	db, mock := utils.SetupMockDB(t)
	defer db.Close()

	news_service := NewSportsNewsWidgetService(db)

	news_widget := &modelsWidget.NewsWidget{
		ID:       1,
		CATEGORY: "sports",
		LANGUAGE: "BUF",
		COUNTRY:  "nfl",
	}

	mock.ExpectExec(regexp.QuoteMeta(`
		UPDATE "NewsWidget"
		SET category = $1, language = $2, country = $3
		WHERE id = $4;
	`)).
		WithArgs(news_widget.CATEGORY, news_widget.LANGUAGE, news_widget.COUNTRY, news_widget.ID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := news_service.UpdateSportsNewsWidget(news_widget)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteSportsNewsWidget(t *testing.T) {
	db, mock := utils.SetupMockDB(t)
	defer db.Close()

	news_service := NewSportsNewsWidgetService(db)

	mock.ExpectExec(regexp.QuoteMeta(`
		DELETE FROM "NewsWidget"
		WHERE id = $1;
	`)).
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := news_service.DeleteSportsNewsWidget(1)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetTeamGame(t *testing.T) {
	db, _ := utils.SetupMockDB(t)
	defer db.Close()

	sports_news_service := NewSportsNewsWidgetService(db)

	news_widget := &modelsWidget.NewsWidget{
		ID:       1,
		LANGUAGE: "Bills (BUF)",
		COUNTRY:  "nfl",
	}

	game_overview, err := sports_news_service.GetTeamGame(news_widget)

	fmt.Printf("%v+\n", game_overview)
	assert.NoError(t, err)
}
