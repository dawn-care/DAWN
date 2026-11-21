package news

import (
	newsCtrlr "server/internal/controllers/widgets/news"

	"github.com/go-chi/chi/v5"
)

// RegisterSportsTeamsRoutes registers the leagues and teams screen widgets can follow.
func RegisterSportsTeamsRoutes(r chi.Router) {
	r.Get("/api/widgets/sports/teams", newsCtrlr.SportsTeamsHandler)
}
