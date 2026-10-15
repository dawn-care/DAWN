package news

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	models "server/internal/models/widgets/news"
)

// A Yankees schedule: a 5-3 home win on Oct 1, a 2-4 away loss on Oct 3, a game on Oct 20
const yankeesSchedule = `{"events":[
	{"competitions":[{"date":"2026-10-01T23:05Z","competitors":[
		{"homeAway":"home","score":{"value":5},"team":{"displayName":"New York Yankees","abbreviation":"NYY"}},
		{"homeAway":"away","score":{"value":3},"team":{"displayName":"Boston Red Sox","abbreviation":"BOS"}}]}]},
	{"competitions":[{"date":"2026-10-03T23:05Z","competitors":[
		{"homeAway":"home","score":{"value":4},"team":{"displayName":"Toronto Blue Jays","abbreviation":"TOR"}},
		{"homeAway":"away","score":{"value":2},"team":{"displayName":"New York Yankees","abbreviation":"NYY"}}]}]},
	{"competitions":[{"date":"2026-10-20T23:05Z","competitors":[
		{"homeAway":"home","score":{"value":0},"team":{"displayName":"New York Yankees","abbreviation":"NYY"}},
		{"homeAway":"away","score":{"value":0},"team":{"displayName":"Houston Astros","abbreviation":"HOU"}}]}]}
]}`

func fakeESPN(t *testing.T, body string) (*httptest.Server, *string) {
	var path string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	return srv, &path
}

func newTestSportsService(baseURL string) *SportsNewsWidgetService {
	s := NewSportsNewsWidgetService(nil)
	s.baseURL = baseURL
	return s
}

func TestGetTeamGameFor(t *testing.T) {
	srv, path := fakeESPN(t, yankeesSchedule)

	game, err := newTestSportsService(srv.URL).GetTeamGameFor("mlb", "NYY")

	require.NoError(t, err)
	assert.Equal(t, "/apis/site/v2/sports/baseball/mlb/teams/NYY/schedule", *path)
	// Which game is "last" depends on today's date; createGameSummaryAt is tested with fixed dates
	assert.NotEmpty(t, game.Result)
}

func TestGetTeamGameFor_Errors(t *testing.T) {
	srv, _ := fakeESPN(t, yankeesSchedule)
	s := newTestSportsService(srv.URL)

	_, err := s.GetTeamGameFor("mls", "NYC")
	assert.ErrorContains(t, err, "unknown league")
	_, err = s.GetTeamGameFor("mlb", "")
	assert.Error(t, err)

	down := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer down.Close()
	_, err = newTestSportsService(down.URL).GetTeamGameFor("mlb", "NYY")
	assert.ErrorContains(t, err, "ESPN API returned status 502")
}

func TestGetTeamGame_OldWidgetFormat(t *testing.T) {
	srv, path := fakeESPN(t, `{"events":[]}`)

	_, err := newTestSportsService(srv.URL).GetTeamGame(&models.NewsWidget{COUNTRY: "nfl", LANGUAGE: "Buffalo Bills (BUF)"})

	require.NoError(t, err)
	assert.Equal(t, "/apis/site/v2/sports/football/nfl/teams/BUF/schedule", *path)
}

func parseSchedule(t *testing.T, body string) *SportAPIResponse {
	var schedule SportAPIResponse
	require.NoError(t, json.Unmarshal([]byte(body), &schedule))
	return &schedule
}

func TestCreateGameSummaryAt(t *testing.T) {
	schedule := parseSchedule(t, yankeesSchedule)
	at := func(month time.Month, day int) *GameSummaryResponse {
		g, err := createGameSummaryAt(schedule, "NYY", time.Date(2026, month, day, 12, 0, 0, 0, time.UTC))
		require.NoError(t, err)
		return g
	}

	won := at(time.October, 2)
	assert.Equal(t, ResultWon, won.Result)
	assert.Equal(t, "Boston Red Sox", won.OpponentTeamName)
	assert.Equal(t, 5, won.UserScore)
	assert.Contains(t, won.Message, "won against the Boston Red Sox at home, with a score of 5 to 3")

	lost := at(time.October, 10)
	assert.Equal(t, ResultLost, lost.Result, "picks the latest game before now, not the upcoming one")
	assert.Equal(t, 2, lost.UserScore)
	assert.Equal(t, 4, lost.OpponentScore)

	preseason := at(time.September, 30) // no game played yet
	assert.Equal(t, ResultUpcoming, preseason.Result, "used to report the next game as a 0-0 loss")
}

func TestCreateGameSummaryAt_TieAndEmpty(t *testing.T) {
	tie := parseSchedule(t, `{"events":[{"competitions":[{"date":"2026-10-01T23:05Z","competitors":[
		{"homeAway":"home","score":{"value":3},"team":{"displayName":"New York Yankees","abbreviation":"NYY"}},
		{"homeAway":"away","score":{"value":3},"team":{"displayName":"Boston Red Sox","abbreviation":"BOS"}}]}]}]}`)
	g, err := createGameSummaryAt(tie, "NYY", time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC))
	require.NoError(t, err)
	assert.Equal(t, ResultTied, g.Result, "a tie used to be reported as a loss")

	g, err = createGameSummaryAt(&SportAPIResponse{}, "NYY", time.Now())
	require.NoError(t, err)
	assert.Equal(t, ResultUpcoming, g.Result)
}

func TestGetGoodNews_FakeServer(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v2/everything", r.URL.Path)
		assert.Equal(t, "goodnewsnetwork.org", r.URL.Query().Get("domains"))
		assert.Equal(t, "news-key", r.Header.Get("X-Api-Key"))
		w.Write([]byte(`{"articles":[{"title":"Puppy rescued","description":"A happy ending"}]}`))
	}))
	defer srv.Close()
	s := NewNewsWidgetService(nil, "news-key")
	s.baseURL = srv.URL

	news, err := s.GetGoodNews()

	require.NoError(t, err)
	require.Len(t, news.Articles, 1)
	assert.Equal(t, "Puppy rescued", news.Articles[0].Title)
}
