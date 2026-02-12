package news

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"time"

	models "server/internal/models/widgets/news"
	modules "server/internal/services/modules"
)

type SportsNewsWidgetService struct {
	db     *sql.DB
	client *http.Client
}

type SportAPIResponse struct {
	Events []struct {
		Competitions []struct {
			Date        string `json:"date"`
			Competitors []struct {
				HomeAway string `json:"homeAway"`
				Score    struct {
					Value float64 `json:"value"`
				} `json:"score"`
				Team struct {
					DisplayName  string `json:"displayName"`
					Abbreviation string `json:"abbreviation"`
				} `json:"team"`
			} `json:"competitors"`
		} `json:"competitions"`
	} `json:"events"`
}

type Team struct {
	Name     string
	HomeAway string
	Score    float64
}

var SPORTS_LEAGUES_MAP = map[string]string{"nfl": "football", "nba": "basketball", "mlb": "baseball", "nhl": "hockey"}

func NewSportsNewsWidgetService(db *sql.DB) *SportsNewsWidgetService {
	return &SportsNewsWidgetService{
		db:     db,
		client: &http.Client{},
	}
}

// Create a new sports sports widget module
func (s *SportsNewsWidgetService) CreateSportsNewsWidget(sports_news_widget *models.NewsWidget) error {
	if sports_news_widget == nil {
		return errors.New("NewsWidget cannot be nil")
	}
	if sports_news_widget.COUNTRY == "" || sports_news_widget.LANGUAGE == "" {
		return errors.New("COUNTRY and LANGUAGE cannot be empty")
	}
	if SPORTS_LEAGUES_MAP[sports_news_widget.COUNTRY] == "" {
		return errors.New("COUNTRY must be one of the valid sports leagues: nfl, nba, mlb, nhl")
	}
	metadata_id, err := models.CreateNews(s.db, sports_news_widget)
	if err != nil {
		return fmt.Errorf("failed to create sports news widget: %v", err)
	}
	moduleService := modules.NewModuleService(s.db)
	_, err = moduleService.CreateModule("sports", 1, 1, metadata_id)
	return err
}

// Create multiple sports widgets in bulk
func (s *SportsNewsWidgetService) CreateSportsNewsWidgetsBulk(sports_news_widgets []models.NewsWidget) error {
	if len(sports_news_widgets) == 0 {
		return errors.New("no news widgets to create")
	}
	for _, sports_news_widget := range sports_news_widgets {
		if sports_news_widget.COUNTRY == "" || sports_news_widget.LANGUAGE == "" || sports_news_widget.CATEGORY == "" {
			return errors.New("COUNTRY, LANGUAGE, and CATEGORY cannot be empty")
		}
		if SPORTS_LEAGUES_MAP[sports_news_widget.COUNTRY] == "" {
			return errors.New("COUNTRY must be one of the valid sports leagues: nfl, nba, mlb, nhl")
		}
	}
	return models.CreateNewsBulk(s.db, sports_news_widgets)
}

// Retrieve a sports widget by its ID
func (s *SportsNewsWidgetService) GetSportsNewsWidgetByID(id int) (*models.NewsWidget, error) {
	if id <= 0 {
		return nil, errors.New("invalid NewsWidget ID")
	}
	sports_news_widget, err := models.GetNewsByID(s.db, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get newswidget by id: %v", err)
	}
	if sports_news_widget == nil {
		return nil, fmt.Errorf("NewsWidget with ID %d not found", id)
	}
	return sports_news_widget, nil
}

// Retrieve all sports widgets for given carepartner and caregiver
func (s *SportsNewsWidgetService) GetSportsNewsWidgetsByCarepartnerAndCaregiver(carepartner_id int, caregiver_id int) ([]models.NewsWidget, error) {
	if carepartner_id <= 0 || caregiver_id <= 0 {
		return nil, errors.New("invalid Carepartner ID or Caregiver ID")
	}
	return models.GetNewsByCarepartnerAndCaregiver(s.db, carepartner_id, caregiver_id)
}

// Retrieve all sports widgets for a carepartner
func (s *SportsNewsWidgetService) GetSportsNewsWidgetsByCarepartnerIDAndDay(carepartner_id int, day, month, year int) ([]models.NewsWidget, error) {
	if carepartner_id <= 0 {
		return nil, errors.New("invalid Carepartner ID")
	}
	if day < 0 || month < 0 || year < 0 || day > 31 || month > 11 {
		return nil, errors.New("invalid date parameters")
	}
	return models.GetNewsByDayAndCarepartner(s.db, carepartner_id, day, month, year)
}

// Update an existing sports widget
func (s *SportsNewsWidgetService) UpdateSportsNewsWidget(sports_news_widget *models.NewsWidget) error {
	if sports_news_widget == nil {
		return errors.New("NewsWidget cannot be nil")
	}
	if sports_news_widget.ID <= 0 {
		return errors.New("invalid NewsWidget ID")
	}
	if sports_news_widget.COUNTRY == "" || sports_news_widget.LANGUAGE == "" || sports_news_widget.CATEGORY == "" {
		return errors.New("COUNTRY, LANGUAGE, and CATEGORY cannot be empty")
	}
	if SPORTS_LEAGUES_MAP[sports_news_widget.COUNTRY] == "" {
		return errors.New("COUNTRY must be one of the valid sports leagues: nfl, nba, mlb, nhl")
	}
	return models.UpdateNews(s.db, sports_news_widget)
}

// Delete a sports widget by its ID
func (s *SportsNewsWidgetService) DeleteSportsNewsWidget(id int) error {
	if id <= 0 {
		return errors.New("invalid NewsWidget ID")
	}
	return models.DeleteNews(s.db, id)
}

// Get total number of sports widgets
func (s *SportsNewsWidgetService) GetSportsNewsWidgetStats() (int, error) {
	return models.GetNewsStats(s.db)
}

// Gets the the schedule for a team then calls the helper function to parse that data into a game summary
func (s *SportsNewsWidgetService) GetTeamGame(sports_news_widget *models.NewsWidget) (string, error) {
	if sports_news_widget.COUNTRY == "" || sports_news_widget.LANGUAGE == "" {
		return "", errors.New("league and team cannot be empty")
	}
	sport := SPORTS_LEAGUES_MAP[sports_news_widget.COUNTRY]
	re := regexp.MustCompile(`\(([^)]+)\)$`)
	match := re.FindStringSubmatch(sports_news_widget.LANGUAGE)
	teamAbbr := ""
	if len(match) > 1 {
		teamAbbr = match[1]
	}

	url := fmt.Sprintf(
		"http://site.api.espn.com/apis/site/v2/sports/%s/%s/teams/%s/schedule",
		sport, sports_news_widget.COUNTRY, teamAbbr,
	)
	resp, err := s.client.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to fetch team schedule: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("ESPN APi returend status %d", resp.StatusCode)
	}
	var apiResponse SportAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		return "", fmt.Errorf("failed to parse team schedule response: %v", err)
	}

	game_summary, err := CreateGameSummary(&apiResponse, sports_news_widget.LANGUAGE)
	if err != nil {
		return "", fmt.Errorf("failed to create game summary: %v", err)
	}

	return game_summary, nil
}

// Parses the API response into a game_summary string
func CreateGameSummary(apiResponse *SportAPIResponse, team string) (string, error) {
	if apiResponse == nil || len(apiResponse.Events) == 0 {
		return "No upcoming games found.", nil
	}

	game_index := 0
	now := time.Now().UTC()
	for i := len(apiResponse.Events) - 1; i >= 0; i-- {
		date, err := time.Parse("2006-01-02T15:04Z", apiResponse.Events[i].Competitions[0].Date)
		if err != nil {
			return "", err
		}
		if date.Before(now) {
			game_index = i
			break
		}
	}

	var user_team, opponent_team Team
	competition := apiResponse.Events[game_index].Competitions[0]
	for _, competitor := range competition.Competitors {
		if competitor.Team.Abbreviation == team {
			user_team = Team{
				Name:     competitor.Team.DisplayName,
				HomeAway: competitor.HomeAway,
				Score:    competitor.Score.Value,
			}
		} else {
			opponent_team = Team{
				Name:     competitor.Team.DisplayName,
				HomeAway: competitor.HomeAway,
				Score:    competitor.Score.Value,
			}
		}
	}

	if user_team.Score > opponent_team.Score {
		if user_team.HomeAway == "home" {
			return fmt.Sprintf("In the last game, your team the %s won against the %s at home, with a score of %d to %d.", user_team.Name, opponent_team.Name, int(user_team.Score), int(opponent_team.Score)), nil
		} else {
			return fmt.Sprintf("In the last game, your team the %s won against the %s away, with a score of %d to %d.", user_team.Name, opponent_team.Name, int(user_team.Score), int(opponent_team.Score)), nil
		}
	} else {
		if user_team.HomeAway == "home" {
			return fmt.Sprintf("In the last game, your team the %s lost against the %s at home, with a score of %d to %d.", user_team.Name, opponent_team.Name, int(user_team.Score), int(opponent_team.Score)), nil
		} else {
			return fmt.Sprintf("In the last game, your team the %s lost against the %s away, with a score of %d to %d.", user_team.Name, opponent_team.Name, int(user_team.Score), int(opponent_team.Score)), nil
		}
	}
}
