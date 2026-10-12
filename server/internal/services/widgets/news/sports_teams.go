package news

// Code below was copied from the caregiver app (caregiver/src/pages/WidgetForm.tsx,
// SPORTS_TEAMS) so the team list lives on the server. Abbreviations are the ones
// the ESPN schedule API expects in GetTeamGame.

// SportsTeam is one team a sports widget can follow
type SportsTeam struct {
	Name string `json:"name"`
	Abbr string `json:"abbr"`
}

// SportsLeague is a sports league and its teams
type SportsLeague struct {
	ID    string       `json:"id"`
	Name  string       `json:"name"`
	Teams []SportsTeam `json:"teams"`
}

// SportsLeagues are the leagues and teams a sports widget can follow, in display order
var SportsLeagues = []SportsLeague{
	{ID: "nfl", Name: "NFL", Teams: []SportsTeam{
		{Name: "Arizona Cardinals", Abbr: "ARI"},
		{Name: "Atlanta Falcons", Abbr: "ATL"},
		{Name: "Baltimore Ravens", Abbr: "BAL"},
		{Name: "Buffalo Bills", Abbr: "BUF"},
		{Name: "Carolina Panthers", Abbr: "CAR"},
		{Name: "Chicago Bears", Abbr: "CHI"},
		{Name: "Cincinnati Bengals", Abbr: "CIN"},
		{Name: "Cleveland Browns", Abbr: "CLE"},
		{Name: "Dallas Cowboys", Abbr: "DAL"},
		{Name: "Denver Broncos", Abbr: "DEN"},
		{Name: "Detroit Lions", Abbr: "DET"},
		{Name: "Green Bay Packers", Abbr: "GB"},
		{Name: "Houston Texans", Abbr: "HOU"},
		{Name: "Indianapolis Colts", Abbr: "IND"},
		{Name: "Jacksonville Jaguars", Abbr: "JAX"},
		{Name: "Kansas City Chiefs", Abbr: "KC"},
		{Name: "Las Vegas Raiders", Abbr: "LV"},
		{Name: "Los Angeles Chargers", Abbr: "LAC"},
		{Name: "Los Angeles Rams", Abbr: "LAR"},
		{Name: "Miami Dolphins", Abbr: "MIA"},
		{Name: "Minnesota Vikings", Abbr: "MIN"},
		{Name: "New England Patriots", Abbr: "NE"},
		{Name: "New Orleans Saints", Abbr: "NO"},
		{Name: "New York Giants", Abbr: "NYG"},
		{Name: "New York Jets", Abbr: "NYJ"},
		{Name: "Philadelphia Eagles", Abbr: "PHI"},
		{Name: "Pittsburgh Steelers", Abbr: "PIT"},
		{Name: "San Francisco 49ers", Abbr: "SF"},
		{Name: "Seattle Seahawks", Abbr: "SEA"},
		{Name: "Tampa Bay Buccaneers", Abbr: "TB"},
		{Name: "Tennessee Titans", Abbr: "TEN"},
		{Name: "Washington Commanders", Abbr: "WSH"},
	}},
	{ID: "nba", Name: "NBA", Teams: []SportsTeam{
		{Name: "Atlanta Hawks", Abbr: "ATL"},
		{Name: "Boston Celtics", Abbr: "BOS"},
		{Name: "Brooklyn Nets", Abbr: "BKN"},
		{Name: "Charlotte Hornets", Abbr: "CHA"},
		{Name: "Chicago Bulls", Abbr: "CHI"},
		{Name: "Cleveland Cavaliers", Abbr: "CLE"},
		{Name: "Dallas Mavericks", Abbr: "DAL"},
		{Name: "Denver Nuggets", Abbr: "DEN"},
		{Name: "Detroit Pistons", Abbr: "DET"},
		{Name: "Golden State Warriors", Abbr: "GS"},
		{Name: "Houston Rockets", Abbr: "HOU"},
		{Name: "Indiana Pacers", Abbr: "IND"},
		{Name: "LA Clippers", Abbr: "LAC"},
		{Name: "Los Angeles Lakers", Abbr: "LAL"},
		{Name: "Memphis Grizzlies", Abbr: "MEM"},
		{Name: "Miami Heat", Abbr: "MIA"},
		{Name: "Milwaukee Bucks", Abbr: "MIL"},
		{Name: "Minnesota Timberwolves", Abbr: "MIN"},
		{Name: "New Orleans Pelicans", Abbr: "NO"},
		{Name: "New York Knicks", Abbr: "NY"},
		{Name: "Oklahoma City Thunder", Abbr: "OKC"},
		{Name: "Orlando Magic", Abbr: "ORL"},
		{Name: "Philadelphia 76ers", Abbr: "PHI"},
		{Name: "Phoenix Suns", Abbr: "PHX"},
		{Name: "Portland Trail Blazers", Abbr: "POR"},
		{Name: "Sacramento Kings", Abbr: "SAC"},
		{Name: "San Antonio Spurs", Abbr: "SAS"},
		{Name: "Toronto Raptors", Abbr: "TOR"},
		{Name: "Utah Jazz", Abbr: "UTAH"},
		{Name: "Washington Wizards", Abbr: "WSH"},
	}},
	{ID: "mlb", Name: "MLB", Teams: []SportsTeam{
		{Name: "Arizona Diamondbacks", Abbr: "ARI"},
		{Name: "Atlanta Braves", Abbr: "ATL"},
		{Name: "Baltimore Orioles", Abbr: "BAL"},
		{Name: "Boston Red Sox", Abbr: "BOS"},
		{Name: "Chicago Cubs", Abbr: "CHC"},
		{Name: "Chicago White Sox", Abbr: "CHW"},
		{Name: "Cincinnati Reds", Abbr: "CIN"},
		{Name: "Cleveland Guardians", Abbr: "CLE"},
		{Name: "Colorado Rockies", Abbr: "COL"},
		{Name: "Detroit Tigers", Abbr: "DET"},
		{Name: "Houston Astros", Abbr: "HOU"},
		{Name: "Kansas City Royals", Abbr: "KC"},
		{Name: "Los Angeles Angels", Abbr: "LAA"},
		{Name: "Los Angeles Dodgers", Abbr: "LAD"},
		{Name: "Miami Marlins", Abbr: "MIA"},
		{Name: "Milwaukee Brewers", Abbr: "MIL"},
		{Name: "Minnesota Twins", Abbr: "MIN"},
		{Name: "New York Mets", Abbr: "NYM"},
		{Name: "New York Yankees", Abbr: "NYY"},
		{Name: "Oakland Athletics", Abbr: "ATH"},
		{Name: "Philadelphia Phillies", Abbr: "PHI"},
		{Name: "Pittsburgh Pirates", Abbr: "PIT"},
		{Name: "San Diego Padres", Abbr: "SD"},
		{Name: "San Francisco Giants", Abbr: "SF"},
		{Name: "Seattle Mariners", Abbr: "SEA"},
		{Name: "St. Louis Cardinals", Abbr: "STL"},
		{Name: "Tampa Bay Rays", Abbr: "TB"},
		{Name: "Texas Rangers", Abbr: "TEX"},
		{Name: "Toronto Blue Jays", Abbr: "TOR"},
		{Name: "Washington Nationals", Abbr: "WSH"},
	}},
	{ID: "nhl", Name: "NHL", Teams: []SportsTeam{
		{Name: "Anaheim Ducks", Abbr: "ANA"},
		{Name: "Boston Bruins", Abbr: "BOS"},
		{Name: "Buffalo Sabres", Abbr: "BUF"},
		{Name: "Calgary Flames", Abbr: "CGY"},
		{Name: "Carolina Hurricanes", Abbr: "CAR"},
		{Name: "Chicago Blackhawks", Abbr: "CHI"},
		{Name: "Colorado Avalanche", Abbr: "COL"},
		{Name: "Columbus Blue Jackets", Abbr: "CBJ"},
		{Name: "Dallas Stars", Abbr: "DAL"},
		{Name: "Detroit Red Wings", Abbr: "DET"},
		{Name: "Edmonton Oilers", Abbr: "EDM"},
		{Name: "Florida Panthers", Abbr: "FLA"},
		{Name: "Los Angeles Kings", Abbr: "LA"},
		{Name: "Minnesota Wild", Abbr: "MIN"},
		{Name: "Montreal Canadiens", Abbr: "MTL"},
		{Name: "Nashville Predators", Abbr: "NSH"},
		{Name: "New Jersey Devils", Abbr: "NJ"},
		{Name: "New York Islanders", Abbr: "NYI"},
		{Name: "New York Rangers", Abbr: "NYR"},
		{Name: "Ottawa Senators", Abbr: "OTT"},
		{Name: "Philadelphia Flyers", Abbr: "PHI"},
		{Name: "Pittsburgh Penguins", Abbr: "PIT"},
		{Name: "San Jose Sharks", Abbr: "SJ"},
		{Name: "Seattle Kraken", Abbr: "SEA"},
		{Name: "St. Louis Blues", Abbr: "STL"},
		{Name: "Tampa Bay Lightning", Abbr: "TB"},
		{Name: "Toronto Maple Leafs", Abbr: "TOR"},
		{Name: "Utah Mammoth", Abbr: "UTA"},
		{Name: "Vancouver Canucks", Abbr: "VAN"},
		{Name: "Vegas Golden Knights", Abbr: "VGK"},
		{Name: "Washington Capitals", Abbr: "WSH"},
		{Name: "Winnipeg Jets", Abbr: "WPG"},
	}},
}

// FindTeam returns the team with this abbreviation in a league
func FindTeam(league, abbr string) (SportsTeam, bool) {
	for _, l := range SportsLeagues {
		if l.ID != league {
			continue
		}
		for _, t := range l.Teams {
			if t.Abbr == abbr {
				return t, true
			}
		}
	}
	return SportsTeam{}, false
}
