package news

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// NewsAPIBaseURL is where News API lives; tests point a service elsewhere
const NewsAPIBaseURL = "https://newsapi.org"

// NewsService fetches good-news headlines for screen widgets.
type NewsService struct {
	apiKey  string
	client  *http.Client
	baseURL string
}

type NewsAPIResponse struct {
	Articles []struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		URLToImage  string `json:"urlToImage"`
	} `json:"articles"`
}

func NewNewsService(apiKey string) *NewsService {
	return &NewsService{
		apiKey:  apiKey,
		client:  &http.Client{Timeout: 10 * time.Second},
		baseURL: NewsAPIBaseURL,
	}
}

/*
Uses News API to search for recent stories from goodnewsnetwork.org.
A news screen widget has no settings, so every news widget uses this as is.
*/
func (s *NewsService) GetGoodNews() (*NewsAPIResponse, error) {
	url := s.baseURL + "/v2/everything?domains=goodnewsnetwork.org"
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
