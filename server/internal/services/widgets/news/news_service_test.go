package news

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"

	modelsWidget "server/internal/models/widgets/news"
	"server/internal/utils"
)

func TestCreateNewsWidget(t *testing.T) {
	db, mock := utils.SetupMockDB(t)
	defer db.Close()

	news_service := NewNewsWidgetService(db, "test_api_key")

	news_widget := &modelsWidget.NewsWidget{
		ID:       1,
		CATEGORY: "news",
		LANGUAGE: "en",
		COUNTRY:  "us",
	}

	mock.ExpectQuery(regexp.QuoteMeta(`
		INSERT INTO "NewsWidget" (category, language, country)
		VALUES ($1, $2, $3)
		RETURNING id;
	`)).
		WithArgs("news", "en", "us").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	err := news_service.CreateNewsWidget(news_widget)
	assert.NoError(t, err)
	assert.Equal(t, 1, news_widget.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetNewsWidgetByID(t *testing.T) {
	db, mock := utils.SetupMockDB(t)
	defer db.Close()

	news_service := NewNewsWidgetService(db, "test_api_key")

	rows := sqlmock.NewRows([]string{"id", "category", "language", "country"}).
		AddRow(1, "news", "en", "us")

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, category, language, country
		FROM "NewsWidget"
		WHERE id = $1;
	`)).
		WithArgs(1).
		WillReturnRows(rows)

	news_widget, err := news_service.GetNewsWidgetByID(1)
	assert.NoError(t, err)
	assert.NotNil(t, news_widget)
	assert.Equal(t, 1, news_widget.ID)
	assert.Equal(t, "news", news_widget.CATEGORY)
	assert.Equal(t, "en", news_widget.LANGUAGE)
	assert.Equal(t, "us", news_widget.COUNTRY)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateNewsWidget(t *testing.T) {
	db, mock := utils.SetupMockDB(t)
	defer db.Close()

	news_service := NewNewsWidgetService(db, "test_api_key")

	news_widget := &modelsWidget.NewsWidget{
		ID:       1,
		CATEGORY: "news",
		LANGUAGE: "en",
		COUNTRY:  "us",
	}

	mock.ExpectExec(regexp.QuoteMeta(`
		UPDATE "NewsWidget"
		SET category = $1, language = $2, country = $3
		WHERE id = $4;
	`)).
		WithArgs(news_widget.CATEGORY, news_widget.LANGUAGE, news_widget.COUNTRY, news_widget.ID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := news_service.UpdateNewsWidget(news_widget)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteNewsWidget(t *testing.T) {
	db, mock := utils.SetupMockDB(t)
	defer db.Close()

	news_service := NewNewsWidgetService(db, "test_api_key")

	mock.ExpectExec(regexp.QuoteMeta(`
		DELETE FROM "NewsWidget"
		WHERE id = $1;
	`)).
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := news_service.DeleteNewsWidget(1)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
func TestGetGoodNews(t *testing.T) {
	db, _ := utils.SetupMockDB(t)
	defer db.Close()
	err := godotenv.Load("../../../../.env")
	if err != nil {
		assert.NoError(t, err)
	}
	apiKey := os.Getenv("NEWS_API_KEY")

	sports_news_service := NewNewsWidgetService(db, apiKey)

	news_response, err := sports_news_service.GetGoodNews()

	fmt.Printf("%+v", news_response.Articles[0])
	assert.NoError(t, err)
}

func TestGetJoke(t *testing.T) {
	db, _ := utils.SetupMockDB(t)
	defer db.Close()

	sports_news_service := NewNewsWidgetService(db, "test_api_key")

	joke_response, err := sports_news_service.GetJoke()

	fmt.Printf("%s", joke_response)
	assert.NoError(t, err)
}
