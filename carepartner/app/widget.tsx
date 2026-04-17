/**
 * widget.tsx
 * 
 * User: Carepartner
 * 
 * Displays dynamic informational widgets (weather and sports).
 * 
 * Responsibilities:
 * - Fetch data from backend based on module type
 * - Render appropriate widget UI (weather, sports, or text)
 * - Handle loading state and fallback content
 * 
 * Navigation:
 * - Controlled by AgendaView
 * 
 * Notes:
 * - Uses POST requests for weather and sports APIs
 * - News content is aggregated and formatted as text
 * - Each widget is rendered via a dedicated subcomponent
 */

import React, { useEffect, useState } from "react";
import { ScrollView, StyleSheet, Text, View } from "react-native";
import { Stack } from "expo-router";
import {
  type ConfigurationWithMetadata,
  WeatherMetadata,
  NewsMetadata,
} from "./agendaView";
import { API_URL } from "./config";

/* ============================
   TYPES
============================ */

/**
 * Response structure for sports API
 */
type GameSummary = {
  user_team_name: string;
  opponent_team_name: string;
  user_score: number;
  opponent_score: number;
  message: string;
};

/**
 * Response structure for weather API
 */
type WeatherData = {
  location: string;
  temp: number;
  feels_like: number;
  temp_min: number;
  temp_max: number;
  humidity: number;
  wind_speed: number;
  condition: string;
  description: string;
  message: string;
};

/**
 * Props passed from AgendaView.
 */
interface WidgetViewerProps {
  configuration: ConfigurationWithMetadata;
}

/* ============================
   HELPERS
============================ */
async function fetchWeather(
  metadata: WeatherMetadata
): Promise<WeatherData | null> {
  try {
    const res = await fetch(`${API_URL}/weather/overview`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(metadata),
    });
    return (await res.json()) as WeatherData;
  } catch (err: any) {
    console.error("Failed getting weather overview: " + err.message);
    return null;
  }
}

async function fetchSports(
  metadata: NewsMetadata
): Promise<GameSummary | null> {
  try {
    const res = await fetch(`${API_URL}/sports/team-game`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(metadata),
    });
    return (await res.json()) as GameSummary;
  } catch (err: any) {
    console.error("Failed getting sports overview: " + err.message);
    return null;
  }
}

/**
 *** API BROKEN ***
 * async function fetchNews(): Promise<string> {
 *   try {
 *     const res = await fetch(`${API_URL}/news/good-news`);
 *     const data = await res.json();
 *     const articles: { title: string; description: string }[] = data.articles ?? [];
 *     if (articles.length === 0) return "No good news articles found.";
 *     return articles
 *       .filter((a) => a.title && a.description)
 *       .slice(0, 5)
 *       .map((a) => `${a.title}\n${a.description}`)
 *       .join("\n\n---\n\n");
 *   } catch (err: any) {
 *     console.error("Failed getting news: " + err.message);
 *     return "";
 *   }
 *  }
 */

/**
 * Reusable glass-style container
 */
function GlassCard({
  children,
  style,
}: {
  children: React.ReactNode;
  style?: any;
}) {
  return <View style={[styles.glassBase, style]}>{children}</View>;
}

/* ============================
   SUBCOMPONENTS
============================ */

/**
 * WeatherView Component
 * 
 * Handles:
 *  - Displaying current temperature and condition
 *  - Showing location and descriptive message
 */
function WeatherView({ weather }: { weather: WeatherData }) {
  return (
    <GlassCard style={styles.card}>
      <View style={styles.weatherLayout}>
        {/* Left: large temperature */}
        <View style={styles.weatherTempColumn}>
          <Text style={styles.weatherTemp}>{weather.temp}°</Text>
          <Text style={styles.weatherCondition}>{weather.condition}</Text>
          <Text style={styles.weatherLocation}>{weather.location}</Text>
        </View>

        {/* Right: message */}
        <View style={styles.weatherMessageColumn}>
          <ScrollView
            contentContainerStyle={styles.weatherMessageScroll}
            showsVerticalScrollIndicator={false}
          >
            <Text style={styles.weatherMessage}>{weather.message}</Text>
          </ScrollView>
        </View>
      </View>

      {/* Bottom details row */}
      <View style={styles.weatherDetailsRow}>
        <Text style={styles.weatherDetail}>Feels like {weather.feels_like}°</Text>
        <Text style={styles.weatherDetailDivider}>·</Text>
        <Text style={styles.weatherDetail}>H: {weather.temp_max}°  L: {weather.temp_min}°</Text>
        <Text style={styles.weatherDetailDivider}>·</Text>
        <Text style={styles.weatherDetail}>Humidity {weather.humidity}%</Text>
        <Text style={styles.weatherDetailDivider}>·</Text>
        <Text style={styles.weatherDetail}>Wind {weather.wind_speed} mph</Text>
      </View>
    </GlassCard>
  );
}

/**
 * SportsView Component
 * 
 * Handles:
 *  - Displaying team names and scores
 *  - Showing summary of last game
 */
function SportsView({ game }: { game: GameSummary }) {
  return (
    <GlassCard style={styles.card}>
      {/* Scoreboard */}
      <View style={styles.scoreboard}>
        {/* Left team */}
        <View style={styles.teamColumn}>
          <Text style={styles.teamName} numberOfLines={2}>
            {game.user_team_name}
          </Text>
        </View>

        {/* Scores */}
        <View style={styles.scoreColumn}>
          <View style={styles.scoreRow}>
            <Text style={styles.scoreText}>{game.user_score}</Text>
            <Text style={styles.scoreDash}>–</Text>
            <Text style={styles.scoreText}>{game.opponent_score}</Text>
          </View>
        </View>

        {/* Right team */}
        <View style={styles.teamColumn}>
          <Text style={styles.teamName} numberOfLines={2}>
            {game.opponent_team_name}
          </Text>
        </View>
      </View>

      {/* Message */}
      <View style={styles.messageContainer}>
        <Text style={styles.messageText}>{game.message}</Text>
      </View>
    </GlassCard>
  );
}

/**
 * TextContentView Component
 * 
 * Handles:
 *  - Displaying scrollable text content (used for news)
 * 
 * function TextContentView({ content }: { content: string }) {
 *  return (
 *    <GlassCard style={styles.card}>
 *       <ScrollView
 *         contentContainerStyle={styles.scrollContent}
 *         showsVerticalScrollIndicator={false}
 *       >
 *         <Text style={styles.contentText}>{content}</Text>
 *       </ScrollView>
 *     </GlassCard>
 *   );
 * }
 */


/* ============================
   COMPONENT
============================ */

/**
 * WidgetViewerScreen Component
 * 
 * Handles:
 *  - Determining widget type from configuration
 *  - Fetching appropriate backend data
 *  - Rendering correct widget view
 *  - Managing loading state
 */
export default function WidgetViewerScreen({
  configuration,
}: WidgetViewerProps) {
  /* ============================
     State
  ============================ */
  const [gameSummary, setGameSummary] = useState<GameSummary | null>(null);
  const [weatherData, setWeatherData] = useState<WeatherData | null>(null);
  const [loading, setLoading] = useState(true);

  /* ============================
     EFFECTS
  ============================ */

  /**
   * When configuration changes, fetch new data based on widget type
   */
  useEffect(() => {
    let mounted = true;

    const load = async () => {
      setLoading(true);

      if (configuration.module_type === "weather") {
        const meta = configuration.metadata as WeatherMetadata | null;
        const data = meta ? await fetchWeather(meta) : null;
        if (mounted) setWeatherData(data);
      } else if (configuration.module_type === "sports") {
        const meta = configuration.metadata as NewsMetadata | null;
        const game = meta ? await fetchSports(meta) : null;
        if (mounted) setGameSummary(game);
      }

      if (mounted) setLoading(false);
    };

    load();
    return () => {
      mounted = false;
    };
  }, [configuration]);

  /* ============================
       RENDER
    ============================ */
  return (
    <>
      <Stack.Screen options={{ headerShown: false }} />

      <View style={styles.safe}>
        <View style={styles.body}>
          {loading ? (
            <GlassCard style={styles.card}>
              <View style={styles.scrollContent}>
                <Text style={styles.contentText}>Loading…</Text>
              </View>
            </GlassCard>
          ) : configuration.module_type === "weather" && weatherData ? (
            <WeatherView weather={weatherData} />
          ) : configuration.module_type === "sports" && gameSummary ? (
            <SportsView game={gameSummary} />
          ) : (
            <></>
            // <TextContentView content={textContent} />
          )}
        </View>
      </View>
    </>
  );
}

// ── Styles ─────────────────────────────────────────────────────

const styles = StyleSheet.create({
  safe: { flex: 1, backgroundColor: "#000" },

  body: {
    flex: 1,
    padding: 18,
  },

  glassBase: {
    borderRadius: 18,
    borderWidth: 1.5,
    borderColor: "rgba(0,0,0,0.25)",
    backgroundColor: "rgba(255,255,255,0.55)",
  },

  card: {
    flex: 1,
    paddingHorizontal: 18,
    paddingVertical: 18,
  },

  /* ── Scoreboard ────────────────────── */

  scoreboard: {
    flexDirection: "row",
    alignItems: "center",
    justifyContent: "center",
    flex: 1,
    paddingHorizontal: 12,
  },

  teamColumn: {
    flex: 1,
    alignItems: "center",
    justifyContent: "center",
  },

  teamName: {
    fontSize: 32,
    fontWeight: "700",
    color: "rgba(0,0,0,0.85)",
    textAlign: "center",
  },

  scoreColumn: {
    paddingHorizontal: 20,
    alignItems: "center",
    justifyContent: "center",
  },

  scoreRow: {
    flexDirection: "row",
    alignItems: "baseline",
  },

  scoreText: {
    fontSize: 64,
    fontWeight: "800",
    color: "rgba(0,0,0,0.9)",
  },

  scoreDash: {
    fontSize: 48,
    fontWeight: "300",
    color: "rgba(0,0,0,0.5)",
    marginHorizontal: 12,
  },

  /* ── Message ───────────────────────── */

  messageContainer: {
    borderTopWidth: 1,
    borderTopColor: "rgba(0,0,0,0.15)",
    paddingTop: 14,
    marginTop: 8,
  },

  messageText: {
    fontSize: 22,
    lineHeight: 30,
    color: "rgba(0,0,0,0.72)",
    fontWeight: "500",
    textAlign: "center",
  },

  /* ── Weather ────────────────────────── */

  weatherLayout: {
    flex: 1,
    flexDirection: "row",
    alignItems: "center",
  },

  weatherTempColumn: {
    alignItems: "center",
    justifyContent: "center",
    paddingRight: 20,
    borderRightWidth: 1,
    borderRightColor: "rgba(0,0,0,0.15)",
    minWidth: 160,
  },

  weatherTemp: {
    fontSize: 80,
    fontWeight: "800",
    color: "rgba(0,0,0,0.9)",
  },

  weatherCondition: {
    fontSize: 24,
    fontWeight: "600",
    color: "rgba(0,0,0,0.7)",
    marginTop: 2,
  },

  weatherLocation: {
    fontSize: 18,
    fontWeight: "400",
    color: "rgba(0,0,0,0.5)",
    marginTop: 4,
  },

  weatherMessageColumn: {
    flex: 1,
    paddingLeft: 20,
    justifyContent: "center",
  },

  weatherMessageScroll: {
    flexGrow: 1,
    justifyContent: "center",
  },

  weatherMessage: {
    fontSize: 22,
    lineHeight: 30,
    color: "rgba(0,0,0,0.78)",
    fontWeight: "500",
  },

  weatherDetailsRow: {
    flexDirection: "row",
    flexWrap: "wrap",
    justifyContent: "center",
    alignItems: "center",
    borderTopWidth: 1,
    borderTopColor: "rgba(0,0,0,0.15)",
    paddingTop: 12,
    marginTop: 10,
  },

  weatherDetail: {
    fontSize: 16,
    color: "rgba(0,0,0,0.6)",
    fontWeight: "500",
  },

  weatherDetailDivider: {
    fontSize: 16,
    color: "rgba(0,0,0,0.3)",
    marginHorizontal: 8,
  },

  /* ── Generic text (weather / news) ── */

  scrollContent: {
    flexGrow: 1,
    justifyContent: "center",
    alignItems: "center",
    paddingBottom: 20,
  },

  contentText: {
    fontSize: 26,
    lineHeight: 36,
    color: "rgba(0,0,0,0.82)",
    fontWeight: "500",
    textAlign: "center",
  },
});
