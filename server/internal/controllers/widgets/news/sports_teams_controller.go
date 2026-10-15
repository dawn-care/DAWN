package widgets

import (
	"encoding/json"
	"net/http"

	services "server/internal/services/widgets/news"
)

// SportsTeamsHandler godoc
// @Summary List the leagues and teams a sports widget can follow
// @Description A sports widget's config is {"league": <league id>, "team": <team abbr>}. Pass league to get one league.
// @Tags sports
// @Produce json
// @Param league query string false "nfl, nba, mlb or nhl"
// @Success 200 {array} news.SportsLeague
// @Failure 404 {object} map[string]string
// @Router /widgets/sports/teams [get]
func SportsTeamsHandler(w http.ResponseWriter, r *http.Request) {
	leagues := services.SportsLeagues
	if id := r.URL.Query().Get("league"); id != "" {
		leagues = nil
		for _, l := range services.SportsLeagues {
			if l.ID == id {
				leagues = []services.SportsLeague{l}
			}
		}
		if leagues == nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{"code": "NOT_FOUND", "error": "unknown league " + id})
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(leagues)
}
