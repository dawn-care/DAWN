package news

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"server/internal/models/modules/widgets"
)

type NewsService struct {
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

func NewNewsService(db *sql.DB, apiKey string) *NewsService {
	return &NewsService{
		db:     db,
		apiKey: apiKey,
		client: &http.Client{},
	}
}

// Create a new news widget module
func (s *NewsService) CreateNewsWidget(news_widget *widgets.NewsWidget) error {
	if news_widget == nil {
		return errors.New("NewsWidget cannot be nil")
	}
	if news_widget.COUNTRY == "" || news_widget.LANGUAGE == "" || news_widget.CATEGORY == "" {
		return errors.New("COUNTRY, LANGUAGE, and CATEGORY cannot be empty")
	}
	return widgets.CreateNews(s.db, news_widget)
}

// Create multiple news widgets in bulk
func (s *NewsService) CreateNewsWidgetsBulk(news_widgets []widgets.NewsWidget) error {
	if len(news_widgets) == 0 {
		return errors.New("no news widgets to create")
	}
	for _, news_widget := range news_widgets {
		if news_widget.COUNTRY == "" || news_widget.LANGUAGE == "" || news_widget.CATEGORY == "" {
			return errors.New("COUNTRY, LANGUAGE, and CATEGORY cannot be empty")
		}
	}
	return widgets.CreateNewsBulk(s.db, news_widgets)
}

// Retrieve a news widget by its ID
func (s *NewsService) GetNewsWidgetByID(id int) (*widgets.NewsWidget, error) {
	if id <= 0 {
		return nil, errors.New("invalid NewsWidget ID")
	}
	news_widget, err := widgets.GetNewsByID(s.db, id)
	if err != nil {
		return nil, fmt.Errorf("Failed to get NewsWidget by ID: %v", err)
	}
	if news_widget == nil {
		return nil, fmt.Errorf("NewsWidget with ID %d not found", id)
	}
	return news_widget, nil
}

// Retrieve all news widgets for given carepartner and caregiver
func (s *NewsService) GetNewsWidgetsByCarepartnerAndCaregiver(carepartner_id int, caregiver_id int) ([]widgets.NewsWidget, error) {
	if carepartner_id <= 0 || caregiver_id <= 0 {
		return nil, errors.New("invalid Carepartner ID or Caregiver ID")
	}
	return widgets.GetNewsByCarepartnerAndCaregiver(s.db, carepartner_id, caregiver_id)
}

// Retrieve all news widgets for a carepartner
func (s *NewsService) GetNewsWidgetsByCarepartnerIDAndDay(carepartner_id int, day, month, year int) ([]widgets.NewsWidget, error) {
	if carepartner_id <= 0 {
		return nil, errors.New("invalid Carepartner ID")
	}
	if day < 0 || month < 0 || year < 0 || day > 31 || month > 11 {
		return nil, errors.New("invalid date parameters")
	}
	return widgets.GetNewsByDayAndCarepartner(s.db, carepartner_id, day, month, year)
}

// Update an existing news widget
func (s *NewsService) UpdateNewsWidget(news_widget *widgets.NewsWidget) error {
	if news_widget == nil {
		return errors.New("NewsWidget cannot be nil")
	}
	if news_widget.ID <= 0 {
		return errors.New("invalid NewsWidget ID")
	}
	return widgets.UpdateNews(s.db, news_widget)
}

// Delete a news widget by its ID
func (s *NewsService) DeleteNewsWidget(id int) error {
	if id <= 0 {
		return errors.New("invalid NewsWidget ID")
	}
	return widgets.DeleteNews(s.db, id)
}

// Get total number of news widgets
func (s *NewsService) GetNewsWidgetStats() (int, error) {
	return widgets.GetNewsStats(s.db)
}

func (s *NewsService) GetGoodNews() (*NewsAPIResponse, error) {
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

func (s *NewsService) GetJoke() (string, error) {
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
