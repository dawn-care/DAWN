package news

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	models "server/internal/models/widgets/news"
)

type NewsWidgetService struct {
	db     *sql.DB
	apiKey string
	client *http.Client
}

type NewsAPIResponse struct {
	Articles []struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		URLToImage  string `json:"urlToImage"`
	} `json:"articles"`
}

func NewNewsWidgetService(db *sql.DB, apiKey string) *NewsWidgetService {
	return &NewsWidgetService{
		db:     db,
		apiKey: apiKey,
		client: &http.Client{},
	}
}

// Create a new news widget module
func (s *NewsWidgetService) CreateNewsWidget(news_widget *models.NewsWidget) error {
	if news_widget == nil {
		return errors.New("NewsWidget cannot be nil")
	}
	if news_widget.COUNTRY == "" || news_widget.LANGUAGE == "" || news_widget.CATEGORY == "" {
		return errors.New("COUNTRY, LANGUAGE, and CATEGORY cannot be empty")
	}
	return models.CreateNews(s.db, news_widget)
}

// Create multiple news widgets in bulk
func (s *NewsWidgetService) CreateNewsWidgetsBulk(news_widgets []models.NewsWidget) error {
	if len(news_widgets) == 0 {
		return errors.New("no news widgets to create")
	}
	for _, news_widget := range news_widgets {
		if news_widget.COUNTRY == "" || news_widget.LANGUAGE == "" || news_widget.CATEGORY == "" {
			return errors.New("COUNTRY, LANGUAGE, and CATEGORY cannot be empty")
		}
	}
	return models.CreateNewsBulk(s.db, news_widgets)
}

// Retrieve a news widget by its ID
func (s *NewsWidgetService) GetNewsWidgetByID(id int) (*models.NewsWidget, error) {
	if id <= 0 {
		return nil, errors.New("invalid NewsWidget ID")
	}
	news_widget, err := models.GetNewsByID(s.db, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get newswidget by id: %v", err)
	}
	if news_widget == nil {
		return nil, fmt.Errorf("NewsWidget with ID %d not found", id)
	}
	return news_widget, nil
}

// Retrieve all news widgets for given carepartner and caregiver
func (s *NewsWidgetService) GetNewsWidgetsByCarepartnerAndCaregiver(carepartner_id int, caregiver_id int) ([]models.NewsWidget, error) {
	if carepartner_id <= 0 || caregiver_id <= 0 {
		return nil, errors.New("invalid Carepartner ID or Caregiver ID")
	}
	return models.GetNewsByCarepartnerAndCaregiver(s.db, carepartner_id, caregiver_id)
}

// Retrieve all news widgets for a carepartner
func (s *NewsWidgetService) GetNewsWidgetsByCarepartnerIDAndDay(carepartner_id int, day, month, year int) ([]models.NewsWidget, error) {
	if carepartner_id <= 0 {
		return nil, errors.New("invalid Carepartner ID")
	}
	if day < 0 || month < 0 || year < 0 || day > 31 || month > 11 {
		return nil, errors.New("invalid date parameters")
	}
	return models.GetNewsByDayAndCarepartner(s.db, carepartner_id, day, month, year)
}

// Update an existing news widget
func (s *NewsWidgetService) UpdateNewsWidget(news_widget *models.NewsWidget) error {
	if news_widget == nil {
		return errors.New("NewsWidget cannot be nil")
	}
	if news_widget.ID <= 0 {
		return errors.New("invalid NewsWidget ID")
	}
	return models.UpdateNews(s.db, news_widget)
}

// Delete a news widget by its ID
func (s *NewsWidgetService) DeleteNewsWidget(id int) error {
	if id <= 0 {
		return errors.New("invalid NewsWidget ID")
	}
	return models.DeleteNews(s.db, id)
}

// Get total number of news widgets
func (s *NewsWidgetService) GetNewsWidgetStats() (int, error) {
	return models.GetNewsStats(s.db)
}

/*
Uses News API to search for reacent stories from goodnewsnetwork.org
*/
func (s *NewsWidgetService) GetGoodNews() (*NewsAPIResponse, error) {
	url := "https://newsapi.org/v2/everything?domains=goodnewsnetwork.org"
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed creating http request: %v", err)
	}
	request.Header.Set("X-Api-Key", s.apiKey)
	response, err := s.client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("failed news API request: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		return nil, fmt.Errorf("news API returned status %d", response.StatusCode)
	}

	var newsResponse NewsAPIResponse
	if err := json.NewDecoder(response.Body).Decode(&newsResponse); err != nil {
		return nil, fmt.Errorf("failed parsing news API response: %v", err)
	}
	return &newsResponse, nil
}

/*
tbh I thought the way I figured out how to get uplifting news was kinda lame and might not be the best feature.
And I didn't really need any of the db attributes for it, so I thought we could use the categories for
widget categories that would call different external APIs that might be uplifting based on their type.
I only had time to do the joke one tonite, but I am looking into inspirational quotes, and fun facts APIs.
ie Category could be "Sports", "News", "Joke", "Quote", etc.
idk lmk wat u think
*/

// Gets a random dad joke
func (s *NewsWidgetService) GetJoke() (string, error) {
	url := "https://icanhazdadjoke.com/"
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed creating http request: %v", err)
	}
	request.Header.Set("Accept", "text/plain")
	response, err := s.client.Do(request)
	if err != nil {
		return "", fmt.Errorf("failed joke API request: %v", err)
	}
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return "", fmt.Errorf("failed reading joke API response: %v", err)
	}
	defer response.Body.Close()

	bodyStr := string(body)
	return bodyStr, nil
}
