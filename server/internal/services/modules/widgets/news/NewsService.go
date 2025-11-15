package news

import (
	"database/sql"
	"errors"
	"fmt"

	"server/internal/models/modules/widgets"
)

type NewsService struct {
	db *sql.DB
}

func NewNewsService(db *sql.DB) *NewsService {
	return &NewsService{db: db}
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
