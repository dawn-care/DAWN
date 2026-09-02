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
)

// Service sturct (ESPN API does not require key)
type SportsNewsWidgetService struct {
	db     *sql.DB
	client *http.Client
}

// Struct to parse response from ESPN Sports API for team schedule
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

// Struct to represent a team from the ESPN Sports API response, used for easier handling in game summary creation
type Team struct {
	Name     string
	HomeAway string
	Score    float64
}

// Struct to represent the game summary information to be sent to the frontend for display in the sports news widget
type GameSummaryResponse struct {
	UserTeamName     string `json:"user_team_name"`
	OpponentTeamName string `json:"opponent_team_name"`
	UserScore        int    `json:"user_score"`
	OpponentScore    int    `json:"opponent_score"`
	Message          string `json:"message"`
}

var SPORTS_LEAGUES_MAP = map[string]string{"nfl": "football", "nba": "basketball", "mlb": "baseball", "nhl": "hockey"}

func NewSportsNewsWidgetService(db *sql.DB) *SportsNewsWidgetService {
	return &SportsNewsWidgetService{
		db:     db,
		client: &http.Client{},
	}
}

// CREATE OPERATIONS

/*
Create a new sports news widget module
sports_news_widget - information inputted formatted as a NewsWidget struct

	** The COUNTRY field representing the league and the LANGUAGE field
	   representing the team (formatted as "Team Name (TEAM ABBREVIATION)")
*/
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
	_, err := models.CreateNews(s.db, sports_news_widget)
	if err != nil {
		return fmt.Errorf("failed to create sports news widget: %v", err)
	}
	return nil
}

/*
Create multiple sports widgets in bulk
sports_news_widgets - List of NewsWidget structs

	** The COUNTRY field representing the league and the LANGUAGE field
	   representing the team (formatted as "Team Name (TEAM ABBREVIATION)")
*/
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

// RETRIEVE OPERATIONS

/*
Retrieve a sports widget by its ID
id - The ID of the sports widget to retrieve
*/
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

/*
Retrieve all sports widgets for given carepartner and caregiver
carepartner_id - The carepartner ID associated with the sports widgets to retrieve
caregiver_id - The caregiver ID associated with the sports widgets to retrieve
*/
func (s *SportsNewsWidgetService) GetSportsNewsWidgetsByCarepartnerAndCaregiver(carepartner_id int, caregiver_id int) ([]models.NewsWidget, error) {
	if carepartner_id <= 0 || caregiver_id <= 0 {
		return nil, errors.New("invalid Carepartner ID or Caregiver ID")
	}
	return models.GetNewsByCarepartnerAndCaregiver(s.db, carepartner_id, caregiver_id)
}

/*
Retrieve all sports widgets for a carepartner on a given day
carepartner_id - The carepartner ID associated with the sports widgets to retrieve
day, month, year - The date to retrieve sports widgets for
*/
func (s *SportsNewsWidgetService) GetSportsNewsWidgetsByCarepartnerIDAndDay(carepartner_id int, day, month, year int) ([]models.NewsWidget, error) {
	if carepartner_id <= 0 {
		return nil, errors.New("invalid Carepartner ID")
	}
	if day < 0 || month < 0 || year < 0 || day > 31 || month > 11 {
		return nil, errors.New("invalid date parameters")
	}
	return models.GetNewsByDayAndCarepartner(s.db, carepartner_id, day, month, year)
}

// UPDATE OPERATION

/*
Update an existing sports widget
sports_news_widget - information inputted formatted as a NewsWidget struct
*/
func (s *SportsNewsWidgetService) UpdateSportsNewsWidget(sports_news_widget *models.NewsWidget) error {
	if sports_news_widget == nil {
		return errors.New("NewsWidget cannot be nil")
	}
	if sports_news_widget.ID <= 0 {
		return errors.New("invalid NewsWidget ID")
	}
	if sports_news_widget.COUNTRY == "" || sports_news_widget.LANGUAGE == "" {
		return errors.New("COUNTRY, LANGUAGE, and CATEGORY cannot be empty")
	}
	if SPORTS_LEAGUES_MAP[sports_news_widget.COUNTRY] == "" {
		return errors.New("COUNTRY must be one of the valid sports leagues: nfl, nba, mlb, nhl")
	}
	return models.UpdateNews(s.db, sports_news_widget)
}

// DELETE OPERATION

/*
Delete a sports widget by its ID
id - The ID of the sports widget to delete
*/
func (s *SportsNewsWidgetService) DeleteSportsNewsWidget(id int) error {
	if id <= 0 {
		return errors.New("invalid NewsWidget ID")
	}
	return models.DeleteNews(s.db, id)
}

// ANALYTICS & STATS

/*
Get total number of sports widgets in the database for analytics purposes
*/
func (s *SportsNewsWidgetService) GetSportsNewsWidgetStats() (int, error) {
	return models.GetNewsStats(s.db)
}

// EXTERNAL API CALLS

/*
Gets the the schedule for a team then calls the helper function to parse that data into a game summary
sports_news_widget - sports news widget containing the league and team information
*/
func (s *SportsNewsWidgetService) GetTeamGame(sports_news_widget *models.NewsWidget) (*GameSummaryResponse, error) {
	if sports_news_widget.COUNTRY == "" || sports_news_widget.LANGUAGE == "" {
		return nil, errors.New("league and team cannot be empty")
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
		return nil, fmt.Errorf("failed to fetch team schedule: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ESPN API returned status %d", resp.StatusCode)
	}
	var apiResponse SportAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		return nil, fmt.Errorf("failed to parse team schedule response: %v", err)
	}

	game_summary, err := CreateGameSummary(&apiResponse, teamAbbr)
	if err != nil {
		return nil, fmt.Errorf("failed to create game summary: %v", err)
	}

	return game_summary, nil
}

/*
Parses the API response into a structured GameSummaryResponse
apiResponse - the response from the ESPN Sports API containing the team schedule and game information
team - the team abbreviation to identify the user's team in the API response
*/
func CreateGameSummary(apiResponse *SportAPIResponse, team string) (*GameSummaryResponse, error) {
	if apiResponse == nil || len(apiResponse.Events) == 0 {
		return &GameSummaryResponse{Message: "No upcoming games found."}, nil
	}

	game_index := 0
	now := time.Now().UTC()
	for i := len(apiResponse.Events) - 1; i >= 0; i-- {
		date, err := time.Parse("2006-01-02T15:04Z", apiResponse.Events[i].Competitions[0].Date)
		if err != nil {
			return nil, err
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

	var message string
	result := "won"
	if user_team.Score <= opponent_team.Score {
		result = "lost"
	}
	location := "away"
	if user_team.HomeAway == "home" {
		location = "at home"
	}
	message = fmt.Sprintf("In the last game, your team the %s %s against the %s %s, with a score of %d to %d.",
		user_team.Name, result, opponent_team.Name, location, int(user_team.Score), int(opponent_team.Score))

	return &GameSummaryResponse{
		UserTeamName:     user_team.Name,
		OpponentTeamName: opponent_team.Name,
		UserScore:        int(user_team.Score),
		OpponentScore:    int(opponent_team.Score),
		Message:          message,
	}, nil
}
