import { useState } from "react";

type WidgetType = "weather" | "sports" | "good";
type RepeatType = "daily" | "weekly" | "monthly" | "yearly" | "never";
type SportsType = "nfl" | "nba" | "mlb" | "nhl";

type Team = {
  name: string;
  abbr: string;
};

const SPORTS_TEAMS: Record<SportsType, Team[]> = {
  nfl: [
    { name: "Arizona Cardinals", abbr: "ARI" },
    { name: "Atlanta Falcons", abbr: "ATL" },
    { name: "Baltimore Ravens", abbr: "BAL" },
    { name: "Buffalo Bills", abbr: "BUF" },
    { name: "Carolina Panthers", abbr: "CAR" },
    { name: "Chicago Bears", abbr: "CHI" },
    { name: "Cincinnati Bengals", abbr: "CIN" },
    { name: "Cleveland Browns", abbr: "CLE" },
    { name: "Dallas Cowboys", abbr: "DAL" },
    { name: "Denver Broncos", abbr: "DEN" },
    { name: "Detroit Lions", abbr: "DET" },
    { name: "Green Bay Packers", abbr: "GB" },
    { name: "Houston Texans", abbr: "HOU" },
    { name: "Indianapolis Colts", abbr: "IND" },
    { name: "Jacksonville Jaguars", abbr: "JAX" },
    { name: "Kansas City Chiefs", abbr: "KC" },
    { name: "Las Vegas Raiders", abbr: "LV" },
    { name: "Los Angeles Chargers", abbr: "LAC" },
    { name: "Los Angeles Rams", abbr: "LAR" },
    { name: "Miami Dolphins", abbr: "MIA" },
    { name: "Minnesota Vikings", abbr: "MIN" },
    { name: "New England Patriots", abbr: "NE" },
    { name: "New Orleans Saints", abbr: "NO" },
    { name: "New York Giants", abbr: "NYG" },
    { name: "New York Jets", abbr: "NYJ" },
    { name: "Philadelphia Eagles", abbr: "PHI" },
    { name: "Pittsburgh Steelers", abbr: "PIT" },
    { name: "San Francisco 49ers", abbr: "SF" },
    { name: "Seattle Seahawks", abbr: "SEA" },
    { name: "Tampa Bay Buccaneers", abbr: "TB" },
    { name: "Tennessee Titans", abbr: "TEN" },
    { name: "Washington Commanders", abbr: "WAS" }
  ],

  nba: [
    { name: "Atlanta Hawks", abbr: "ATL" },
    { name: "Boston Celtics", abbr: "BOS" },
    { name: "Brooklyn Nets", abbr: "BKN" },
    { name: "Charlotte Hornets", abbr: "CHA" },
    { name: "Chicago Bulls", abbr: "CHI" },
    { name: "Cleveland Cavaliers", abbr: "CLE" },
    { name: "Dallas Mavericks", abbr: "DAL" },
    { name: "Denver Nuggets", abbr: "DEN" },
    { name: "Detroit Pistons", abbr: "DET" },
    { name: "Golden State Warriors", abbr: "GSW" },
    { name: "Houston Rockets", abbr: "HOU" },
    { name: "Indiana Pacers", abbr: "IND" },
    { name: "LA Clippers", abbr: "LAC" },
    { name: "Los Angeles Lakers", abbr: "LAL" },
    { name: "Memphis Grizzlies", abbr: "MEM" },
    { name: "Miami Heat", abbr: "MIA" },
    { name: "Milwaukee Bucks", abbr: "MIL" },
    { name: "Minnesota Timberwolves", abbr: "MIN" },
    { name: "New Orleans Pelicans", abbr: "NOP" },
    { name: "New York Knicks", abbr: "NYK" },
    { name: "Oklahoma City Thunder", abbr: "OKC" },
    { name: "Orlando Magic", abbr: "ORL" },
    { name: "Philadelphia 76ers", abbr: "PHI" },
    { name: "Phoenix Suns", abbr: "PHX" },
    { name: "Portland Trail Blazers", abbr: "POR" },
    { name: "Sacramento Kings", abbr: "SAC" },
    { name: "San Antonio Spurs", abbr: "SAS" },
    { name: "Toronto Raptors", abbr: "TOR" },
    { name: "Utah Jazz", abbr: "UTA" },
    { name: "Washington Wizards", abbr: "WAS" }
  ],

  mlb: [
    { name: "Arizona Diamondbacks", abbr: "ARI" },
    { name: "Atlanta Braves", abbr: "ATL" },
    { name: "Baltimore Orioles", abbr: "BAL" },
    { name: "Boston Red Sox", abbr: "BOS" },
    { name: "Chicago Cubs", abbr: "CHC" },
    { name: "Chicago White Sox", abbr: "CWS" },
    { name: "Cincinnati Reds", abbr: "CIN" },
    { name: "Cleveland Guardians", abbr: "CLE" },
    { name: "Colorado Rockies", abbr: "COL" },
    { name: "Detroit Tigers", abbr: "DET" },
    { name: "Houston Astros", abbr: "HOU" },
    { name: "Kansas City Royals", abbr: "KC" },
    { name: "Los Angeles Angels", abbr: "LAA" },
    { name: "Los Angeles Dodgers", abbr: "LAD" },
    { name: "Miami Marlins", abbr: "MIA" },
    { name: "Milwaukee Brewers", abbr: "MIL" },
    { name: "Minnesota Twins", abbr: "MIN" },
    { name: "New York Mets", abbr: "NYM" },
    { name: "New York Yankees", abbr: "NYY" },
    { name: "Oakland Athletics", abbr: "OAK" },
    { name: "Philadelphia Phillies", abbr: "PHI" },
    { name: "Pittsburgh Pirates", abbr: "PIT" },
    { name: "San Diego Padres", abbr: "SD" },
    { name: "San Francisco Giants", abbr: "SF" },
    { name: "Seattle Mariners", abbr: "SEA" },
    { name: "St. Louis Cardinals", abbr: "STL" },
    { name: "Tampa Bay Rays", abbr: "TB" },
    { name: "Texas Rangers", abbr: "TEX" },
    { name: "Toronto Blue Jays", abbr: "TOR" },
    { name: "Washington Nationals", abbr: "WSH" }
  ],

  nhl: [
    { name: "Anaheim Ducks", abbr: "ANA" },
    { name: "Arizona Coyotes", abbr: "ARI" },
    { name: "Boston Bruins", abbr: "BOS" },
    { name: "Buffalo Sabres", abbr: "BUF" },
    { name: "Calgary Flames", abbr: "CGY" },
    { name: "Carolina Hurricanes", abbr: "CAR" },
    { name: "Chicago Blackhawks", abbr: "CHI" },
    { name: "Colorado Avalanche", abbr: "COL" },
    { name: "Columbus Blue Jackets", abbr: "CBJ" },
    { name: "Dallas Stars", abbr: "DAL" },
    { name: "Detroit Red Wings", abbr: "DET" },
    { name: "Edmonton Oilers", abbr: "EDM" },
    { name: "Florida Panthers", abbr: "FLA" },
    { name: "Los Angeles Kings", abbr: "LAK" },
    { name: "Minnesota Wild", abbr: "MIN" },
    { name: "Montreal Canadiens", abbr: "MTL" },
    { name: "Nashville Predators", abbr: "NSH" },
    { name: "New Jersey Devils", abbr: "NJD" },
    { name: "New York Islanders", abbr: "NYI" },
    { name: "New York Rangers", abbr: "NYR" },
    { name: "Ottawa Senators", abbr: "OTT" },
    { name: "Philadelphia Flyers", abbr: "PHI" },
    { name: "Pittsburgh Penguins", abbr: "PIT" },
    { name: "San Jose Sharks", abbr: "SJ" },
    { name: "Seattle Kraken", abbr: "SEA" },
    { name: "St. Louis Blues", abbr: "STL" },
    { name: "Tampa Bay Lightning", abbr: "TB" },
    { name: "Toronto Maple Leafs", abbr: "TOR" },
    { name: "Vancouver Canucks", abbr: "VAN" },
    { name: "Vegas Golden Knights", abbr: "VGK" },
    { name: "Washington Capitals", abbr: "WSH" },
    { name: "Winnipeg Jets", abbr: "WPG" }
  ]
};

const API_URL = import.meta.env.VITE_API_URL || "http://localhost:8080";

export default function CreateWidget() {
    const [widgetType, setWidgetType] = useState<WidgetType>("weather");
    const [sportsType, setSportsType] = useState<SportsType>("nfl");
    const [sportsTeam, setSportsTeam] = useState<Team | null>(null);
    const [zipCode, setZipCode] = useState<string>("10001");
    const [repeat, setRepeat] = useState<RepeatType>("daily");
    const [weeklyDay, setWeeklyDay] = useState("Monday");
    const [monthlyDay, setMonthlyDay] = useState<number>(1);
    const [yearlyDate, setYearlyDate] = useState("2025-01-01");
    const [showsFor, setShowsFor] = useState<number>(10);
    const [uploadStatus, setUploadStatus] = useState("");


    async function handleUpload() {

        const formData = new FormData();
        formData.append("id", "1"); // temp user ID
        formData.append("repeat", repeat);
        if (repeat === "weekly") formData.append("weeklyDay", weeklyDay);
        if (repeat === "monthly") formData.append("monthlyDay", monthlyDay.toString());
        if (repeat === "yearly") formData.append("yearlyDate", yearlyDate);
        if (widgetType === "weather") formData.append("zipcode", zipCode);
        if (widgetType === "sports") {
            formData.append("country", sportsType);

            if (sportsTeam) {
                formData.append("language", sportsTeam.abbr);
            }
        }

        const endpoint =
            widgetType === "weather"
                ? "/weather/upload"
                : widgetType === "sports"
                ? "/sports-news/upload"
                : "/news/upload";
        
        try {
        const response = await fetch(`${API_URL}${endpoint}`, {
            method: "POST",
            body: formData,
        });

        if (!response.ok) {
            const text = await response.text();
            setUploadStatus(`Error: ${text}`);
            return;
        }

        setUploadStatus("Saved!");
        } catch (err: any) {
        setUploadStatus("Upload failed: " + err.message);
        }
    }


    return (
        <div className="w-full flex flex-col items-center mt-8 px-6">
            <h1 className="text-3xl font-semibold mb-6">Add Media</h1>
            <div className="border p-10 max-w-3xl w-full rounded-lg bg-white shadow">

                {/* Widget Type */}
                <div className="mb-6">
                    <label className="text-lg font-medium">Widget Type:</label>
                    <select
                        value={widgetType}
                        onChange={(e) => setWidgetType(e.target.value as WidgetType)}
                        className="w-full mt-2 p-3 border rounded bg-purple-50"
                    >
                        <option value="weather">Weather News</option>
                        <option value="sports">Sports News</option>
                        <option value="good">Good News</option>
                    </select>
                </div>

                {/* Conditional Widget Type Inputs */}
                {widgetType === "weather" && (
                <div className="mb-6">
                    <label className="text-lg font-medium">Zipcode:</label>
                    <input
                        inputMode="numeric"
                        pattern="\d{5}"
                        maxLength={5}
                        value={zipCode}
                        onChange={(e) => setZipCode(e.target.value)}
                        className="w-full mt-2 p-3 border rounded bg-purple-50"
                    >
                    </input>
                </div>
                )}

                {widgetType === "sports" && (
                    <div className="mb-6">
                        <label className="text-lg font-medium">Select Sport:</label>
                        <select
                            value={sportsType}
                            onChange={(e) => setSportsType(e.target.value as SportsType)}
                            className="w-full mt-2 p-3 border rounded bg-purple-50"
                        >
                            <option value="nfl">NFL</option>
                            <option value="nba">NBA</option>
                            <option value="mlb">MLB</option>
                            <option value="nhl">NHL</option>
                        </select>
                    </div>
                )}
                {widgetType === "sports" && (
                    <div className="mb-6">
                        <label className="text-lg font-medium">Select Team:</label>
                        <select
                        value={sportsTeam?.abbr || ""}
                        onChange={(e) => {
                            const selected = SPORTS_TEAMS[sportsType].find(
                            (t) => t.abbr === e.target.value
                            );
                            setSportsTeam(selected || null);
                        }}
                        className="w-full mt-2 p-3 border rounded bg-purple-50"
                        >
                        <option value=""></option>
                        {SPORTS_TEAMS[sportsType].map((team) => (
                            <option key={team.abbr} value={team.abbr}>
                            {team.name} ({team.abbr})
                            </option>
                        ))}
                        </select>
                    </div>
                )}

                {/* Shows For Option */}
                <div className="mb-6">
                <label className="text-lg font-medium">Shows For:</label>
                <select
                    value={showsFor}
                    onChange={(e) => setShowsFor(Number(e.target.value))}
                    className="w-full mt-2 p-3 border rounded bg-purple-50"
                >
                    <option value="10">10 Seconds</option>
                    <option value="20">20 Seconds</option>
                    <option value="30">30 Seconds</option>
                    <option value="40">40 Second</option>
                    <option value="50">50 Seconds</option>
                    <option value="60">60 Seconds</option>
                </select>
                </div>

                {/* Repeat Option */}
                <div className="mb-6">
                <label className="text-lg font-medium">Repeat:</label>
                <select
                    value={repeat}
                    onChange={(e) => setRepeat(e.target.value as RepeatType)}
                    className="w-full mt-2 p-3 border rounded bg-purple-50"
                >
                    <option value="daily">Daily</option>
                    <option value="weekly">Weekly</option>
                    <option value="monthly">Monthly</option>
                    <option value="yearly">Yearly</option>
                    <option value="never">Never</option>
                </select>
                </div>

                {/* Conditional Repeat Inputs */}
                {repeat === "weekly" && (
                <div className="mb-6">
                    <label className="text-lg font-medium">Select Weekday:</label>
                    <select
                    value={weeklyDay}
                    onChange={(e) => setWeeklyDay(e.target.value)}
                    className="w-full mt-2 p-3 border rounded bg-purple-50"
                    >
                    {["Monday","Tuesday","Wednesday","Thursday","Friday","Saturday","Sunday"].map(day => (
                        <option key={day} value={day}>{day}</option>
                    ))}
                    </select>
                </div>
                )}

                {repeat === "monthly" && (
                <div className="mb-6">
                    <label className="text-lg font-medium">Select Day:</label>
                    <select
                    value={monthlyDay}
                    onChange={(e) => setMonthlyDay(parseInt(e.target.value))}
                    className="w-full mt-2 p-3 border rounded bg-purple-50"
                    >
                    {Array.from({length:28}, (_,i)=>i+1).map(day => (
                        <option key={day} value={day}>{day}</option>
                    ))}
                    </select>
                </div>
                )}

                {repeat === "yearly" && (
                <div className="mb-6">
                    <label className="text-lg font-medium">Select Date:</label>
                    <input
                    type="date"
                    value={yearlyDate}
                    onChange={(e) => setYearlyDate(e.target.value)}
                    className="w-full mt-2 p-3 border rounded bg-purple-50"
                    />
                </div>
                )}

                {/* Save */}
                <button
                className="mt-4 w-full bg-teal-200 hover:bg-teal-300 text-gray-800 font-semibold py-3 rounded"
                onClick={handleUpload}
                >
                Save
                </button>

                {uploadStatus && <p className="mt-2 text-sm">{uploadStatus}</p>}
            </div>
        </div>
    );
}