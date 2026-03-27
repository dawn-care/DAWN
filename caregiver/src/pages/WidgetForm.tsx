import { useEffect, useState } from "react";
import { useParams, useNavigate } from 'react-router-dom';

type WidgetType = "weather" | "sports" | "good";
type SportsType = "nfl" | "nba" | "mlb" | "nhl";

type Team = {
  name: string;
  abbr: string;
};

// Create a prop interface so that Library can pass parameters to the form for editing existing widgets without relying on URL params
interface WidgetFormProps {
  moduleId?: number;
  onDeleteSuccess?: () => void;
  onUpdateSuccess?: () => void;
}

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

const API_URL = import.meta.env.VITE_API_URL;

export default function WidgetForm({ 
        moduleId: propModuleId,
        onDeleteSuccess,
        onUpdateSuccess
    }: WidgetFormProps) {
    //For Edit Mode
    const { moduleId } = useParams<{ moduleId: string }>();

    // Use either the prop for library call or the URL param for link access
    const actualModuleId = propModuleId ?? moduleId;
    const isEditMode = Boolean(actualModuleId);

    const [widgetType, setWidgetType] = useState<WidgetType>("weather");
    const [widgetId, setWidgetId] = useState<number | null>(null);
    const [sportsType, setSportsType] = useState<SportsType>("nfl");
    const [sportsTeam, setSportsTeam] = useState<Team | null>(null);
    const [zipCode, setZipCode] = useState<string>("10001");
    const [uploadStatus, setUploadStatus] = useState("");
    const navigate = useNavigate();

    useEffect(() => {
        if (!actualModuleId) return;

        async function loadWidget() {
            var res = await fetch(`${API_URL}/modules/${actualModuleId}`);
            if (!res.ok) {
                setUploadStatus("Failed to load module");
                return;
            }
            const module = await res.json();
            console.log("Loaded module for editing:", module);

            setWidgetType(module.module_type);
            setWidgetId(module.metadata_id);

            if (module.module_type === "weather") {
                res = await fetch(`${API_URL}/weather?id=${module.metadata_id}`);
                if (!res.ok) {
                    setUploadStatus("Failed to load weather data");
                    return;
                }
                const widget = await res.json();
                console.log("Loaded weather widget metadata:", widget);
                setZipCode(widget.zipcode);
            }

            if (module.module_type === "sports") {
                res = await fetch(`${API_URL}/sports?id=${module.metadata_id}`);
                if (!res.ok) {
                    setUploadStatus("Failed to load sports data");
                    return;
                }
                const widget = await res.json();
                console.log("Loaded sports widget metadata:", widget);
                const teamStr = widget.language;
                const match = teamStr.match(/\(([^)]+)\)$/);
                const teamAbbr = match ? match[1] : teamStr;

                const country = widget.country as SportsType;

                const foundTeam = SPORTS_TEAMS[country].find(
                (team) => team.abbr.toUpperCase() === teamAbbr.toUpperCase()
                );

                setSportsType(country);
                setSportsTeam(foundTeam || null);
            }
        }

        loadWidget();
    }, [actualModuleId]);

    async function handleUpload() {
        let data;
        if (widgetType === "weather") {
            data = { ZIPCODE: zipCode };
        } else if (widgetType === "sports") {
            data = { CATEGORY: "", COUNTRY: sportsType, LANGUAGE: `${sportsTeam?.name} (${sportsTeam?.abbr})` || "" };
        } else {
            data = { CATEGORY: "", COUNTRY: "", LANGUAGE: "" };
        }

        const endpoint =
            widgetType === "weather"
                ? "/weather/upload"
                : widgetType === "sports"
                ? "/sports/upload"
                : "/news/upload";

        try {
            const response = await fetch(`${API_URL}${endpoint}`, {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify(data),
            });

            if (!response.ok) {
                const text = await response.text();
                setUploadStatus(`Error: ${text}`);
                return;
            }

            // Safely parse JSON if any
            const text = await response.text();
            let widgetMetadata: any = {};
            if (text) {
                try {
                    widgetMetadata = JSON.parse(text);
                } catch {
                    console.warn("Response not valid JSON, using empty object");
                }
            }

            const metadataId = widgetMetadata.id;
            if (metadataId) {
                setUploadStatus(`Saved! Metadata ID: ${metadataId}`);
            } else {
                setUploadStatus("Saved!");
            }

            // Reset form if needed
            setSportsTeam(null);
            setZipCode("10001");
        } catch (err: any) {
            setUploadStatus("Upload failed: " + err.message);
        }
    }

    async function handleUpdate() {
        var data;
        if (widgetType === "weather") {
            data = {
                ID: widgetId,
                ZIPCODE: zipCode,
            };
        } else if (widgetType === "sports") {
            data = {
                ID: widgetId,
                CATEGORY: "",
                COUNTRY: sportsType,
                LANGUAGE: `${sportsTeam?.name} (${sportsTeam?.abbr})` || "",
            };
        } else {
            data = {
                ID: widgetId,
                CATEGORY: "",
                COUNTRY: "",
                LANGUAGE: "",
            };
        }

        const endpoint =
            widgetType === "weather"
                ? "/weather/update"
                : widgetType === "sports"
                ? "/sports/update"
                : "/news/update";
        
        try {
            const response = await fetch(`${API_URL}${endpoint}`, {
                method: "PUT",
                headers: {
                    "Content-Type": "application/json",
                },
                body: JSON.stringify(data),
            });

            if (!response.ok) {
                const text = await response.text();
                setUploadStatus(`Error: ${text}`);
                return;
            }

            setUploadStatus("Updated Successfully!");
            if (onUpdateSuccess) {
                onUpdateSuccess();
            }
        } catch (err: any) {
            setUploadStatus("Update failed: " + err.message);
        }
    }

    async function handleDelete() {
        if (!confirm("Are you sure you want to delete this widget?")) return;
        if (!actualModuleId) return;

        const moduleEndpoint = `/modules/delete?id=${actualModuleId}`;

        const widgetEndpoint =
            widgetType === "weather"
                ? `/weather/delete?id=${widgetId}`
                : widgetType === "sports"
                ? `/sports/delete?id=${widgetId}`
                : `/news/delete?id=${widgetId}`;
        try{
            const moduleResponse = await fetch(`${API_URL}${moduleEndpoint}`, {
                method: "DELETE",
            });

            if (!moduleResponse.ok) {
                const text = await moduleResponse.text();
                setUploadStatus(`Error: ${text}`);
                return;
            }

            const widgetResponse = await fetch(`${API_URL}${widgetEndpoint}`, {
                method: "DELETE",
            });

            if (!widgetResponse.ok) {
                const text = await widgetResponse.text();
                setUploadStatus(`Error: ${text}`);
                return;
            }

            setUploadStatus("Deleted!");
            if (onDeleteSuccess) {
                onDeleteSuccess();
            } else {
                navigate("/library");
            }
        } catch (err: any) {
            setUploadStatus("Delete failed: " + err.message);
        }
    }

    return (
        <div className="w-full flex flex-col items-center mt-8 px-6">
            {isEditMode ? (<h1 className="text-3xl font-semibold mb-6">Edit Widget</h1>) : 
                (<h1 className="text-3xl font-semibold mb-6">Add Widget</h1>)}
            <div className="border p-10 max-w-3xl w-full rounded-lg bg-white shadow">

                {/* Widget Type */}
                <div className="mb-6">
                    <label className="text-lg font-medium">Widget Type:</label>
                    {isEditMode ? (
                        <div className="w-full mt-2 p-3 border rounded bg-gray-100 text-gray-700">
                        {widgetType === "weather"
                            ? "Weather News"
                            : widgetType === "sports"
                            ? "Sports News"
                            : "Good News"}
                        </div>
                    ) : (
                        <select
                        value={widgetType}
                        onChange={(e) => setWidgetType(e.target.value as WidgetType)}
                        className="w-full mt-2 p-3 border rounded bg-purple-50"
                        >
                        <option value="weather">Weather News</option>
                        <option value="sports">Sports News</option>
                        <option value="good">Good News</option>
                        </select>
                    )}
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
                        value={sportsTeam?.name || ""}
                        onChange={(e) => {
                            const selected = SPORTS_TEAMS[sportsType].find(
                            (t) => t.name === e.target.value
                            );
                            setSportsTeam(selected || null);
                        }}
                        className="w-full mt-2 p-3 border rounded bg-purple-50"
                        >
                        <option value=""></option>
                        {SPORTS_TEAMS[sportsType].map((team) => (
                            <option key={team.abbr} value={team.name}>
                            {team.name}
                            </option>
                        ))}
                        </select>
                    </div>
                )}

                {isEditMode ? (
                    <>
                        {/* Edit Mode */}
                        <button
                        className="mt-4 w-full bg-teal-200 hover:bg-teal-300 text-gray-800 font-semibold py-3 rounded"
                        onClick={handleUpdate}
                        >
                        Update
                        </button>
                        <button
                        className="mt-4 w-full bg-teal-200 hover:bg-teal-300 text-gray-800 font-semibold py-3 rounded"
                        onClick={handleDelete}
                        >
                        Delete
                        </button>
                    </>
                    ) : (
                    <>
                        {/* Create Mode */}
                        <button
                        className="mt-4 w-full bg-teal-200 hover:bg-teal-300 text-gray-800 font-semibold py-3 rounded"
                        onClick={handleUpload}
                        >
                        Save
                        </button>
                    </>
                )}
                {uploadStatus && <p className="mt-2 text-sm">{uploadStatus}</p>}
            </div>
        </div>
    );
}