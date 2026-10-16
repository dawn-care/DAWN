package widgets

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	services "server/internal/services/widgets/news"
)

func TestSportsTeamsHandler_All(t *testing.T) {
	w := httptest.NewRecorder()
	SportsTeamsHandler(w, httptest.NewRequest(http.MethodGet, "/api/widgets/sports/teams", nil))

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var leagues []services.SportsLeague
	if err := json.Unmarshal(w.Body.Bytes(), &leagues); err != nil {
		t.Fatal(err)
	}
	counts := map[string]int{}
	for _, l := range leagues {
		counts[l.ID] = len(l.Teams)
	}
	want := map[string]int{"nfl": 32, "nba": 30, "mlb": 30, "nhl": 32}
	for id, n := range want {
		if counts[id] != n {
			t.Errorf("%s: expected %d teams, got %d", id, n, counts[id])
		}
	}
}

func TestSportsTeamsHandler_OneLeague(t *testing.T) {
	w := httptest.NewRecorder()
	SportsTeamsHandler(w, httptest.NewRequest(http.MethodGet, "/api/widgets/sports/teams?league=mlb", nil))

	var leagues []services.SportsLeague
	_ = json.Unmarshal(w.Body.Bytes(), &leagues)
	if len(leagues) != 1 || leagues[0].ID != "mlb" {
		t.Fatalf("expected only mlb, got %+v", leagues)
	}
	if team, ok := services.FindTeam("mlb", "NYY"); !ok || team.Name != "New York Yankees" {
		t.Errorf("expected the Yankees, got %+v", team)
	}

	w = httptest.NewRecorder()
	SportsTeamsHandler(w, httptest.NewRequest(http.MethodGet, "/api/widgets/sports/teams?league=mls", nil))
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404 for an unknown league, got %d", w.Code)
	}
}
