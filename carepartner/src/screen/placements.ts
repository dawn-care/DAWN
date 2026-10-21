/**
 * placements.ts
 *
 * Turns the screen's widgets into what the grid draws (`WidgetPlacement`),
 * checking each type's `config` and `data` once here. Logic only.
 *
 * The day and time isn't one of the server's widgets: every screen has it in
 * the top left (`DAY_TIME_PLACEMENT`, at `DAY_TIME_AREA` in `@shared`), so the
 * server's widgets can't go there.
 *
 * A widget is left out, so its cells stay empty, when:
 * - its type is unknown to this version, or `createDrawCheck` in `@shared`
 *   drops it: it isn't its type's fixed size, doesn't fit on the grid, or
 *   overlaps the day and time or a widget before it (the same check as the
 *   caregiver's editor);
 * - its outside service is down (`error`), or the data it needs is missing or
 *   malformed.
 */

import { createDrawCheck, DAY_TIME_AREA, parseClock, type ScreenWidgetType } from "@shared";
import { resolveApiUri } from "../api/resolveApiUri";
import type { ReminderItem, WidgetPlacement } from "../widgets/types";
import type { Screen, ScreenWidget } from "./screen";

type Look = WidgetPlacement extends infer W ? (W extends WidgetPlacement ? Pick<W, "renderer" | "data"> : never) : never;

const isObject = (value: unknown): value is Record<string, unknown> => typeof value === "object" && value !== null;
const text = (value: unknown): string | null => (typeof value === "string" && value.trim() ? value.trim() : null);

/** Weather conditions (OpenWeather's main groups) as a tile icon key and a plain word. */
const WEATHER: Record<string, { icon: string; word: string }> = {
  Clear: { icon: "sunny", word: "Sunny" },
  Clouds: { icon: "cloudy", word: "Cloudy" },
  Rain: { icon: "rain", word: "Rainy" },
  Drizzle: { icon: "rain", word: "Rainy" },
  Thunderstorm: { icon: "storm", word: "Stormy" },
  Snow: { icon: "snow", word: "Snowy" },
  Mist: { icon: "fog", word: "Foggy" },
  Fog: { icon: "fog", word: "Foggy" },
  Haze: { icon: "fog", word: "Hazy" },
};

/** "48°" for a temperature in degrees Fahrenheit, or null if it isn't a number. */
const degrees = (value: unknown): string | null =>
  typeof value === "number" && Number.isFinite(value) ? `${Math.round(value)}°` : null;

const LEAGUE_ICONS: Record<string, string> = { nfl: "football", nba: "basketball", mlb: "baseball", nhl: "hockey" };
const RESULTS: Record<string, string> = { won: "Won", lost: "Lost", tied: "Tied" };

/**
 * A minute after midnight as "8:30 AM", or "8 AM" on the hour: shorter than the
 * caregiver's "8:00 am" (`timeLabel` in `@shared`), so reminders fit the tile.
 */
function reminderTime(minute: number): string {
  const hours = Math.floor(minute / 60);
  const minutes = minute % 60;
  const hour12 = hours % 12 || 12;
  const suffix = hours < 12 ? "AM" : "PM";
  return minutes ? `${hour12}:${String(minutes).padStart(2, "0")} ${suffix}` : `${hour12} ${suffix}`;
}

function readReminderItem(value: unknown): ReminderItem | null {
  if (!isObject(value)) return null;
  const minute = typeof value.time === "string" ? parseClock(value.time) : null;
  const words = text(value.text);
  if (minute == null || !words) return null;
  return { minute, timeLabel: reminderTime(minute), text: words, icon: typeof value.icon === "string" ? value.icon : "other" };
}

/** The day and time of day, in the top left of every screen. */
export const DAY_TIME_PLACEMENT: WidgetPlacement = {
  id: "day_time",
  ...DAY_TIME_AREA,
  renderer: "orientation",
  data: { showDate: false },
};

/** How each widget type is drawn, or null when it has nothing to show. */
const LOOKS: Record<Exclude<ScreenWidgetType, "day_time">, (widget: ScreenWidget) => Look | null> = {
  photo: ({ photo, data }) =>
    photo && {
      renderer: "photo",
      data: { uri: resolveApiUri(photo.url), name: (isObject(data) && text(data.caption)) || text(photo.displayName) },
    },
  message: ({ config, photo }) => {
    const words = isObject(config) ? text(config.text) : null;
    if (!isObject(config) || !words) return null;
    return { renderer: "message", data: { text: words, label: text(config.label), photoUri: resolveApiUri(photo?.url) } };
  },
  reminder: ({ config }) => {
    if (!isObject(config) || !Array.isArray(config.items)) return null;
    const items = config.items
      .map(readReminderItem)
      .filter((item): item is ReminderItem => item !== null)
      .sort((a, b) => a.minute - b.minute);
    return items.length ? { renderer: "reminder", data: { items } } : null;
  },
  weather: ({ data }) => {
    const temp = isObject(data) ? degrees(data.temp) : null;
    if (!isObject(data) || !temp) return null;
    const condition = typeof data.condition === "string" ? WEATHER[data.condition] : undefined;
    const high = degrees(data.temp_max);
    const low = degrees(data.temp_min);
    return {
      renderer: "wideTile",
      data: {
        icon: condition?.icon ?? "",
        label: condition?.word ?? "Today",
        word: temp,
        detail: high && low ? `High ${high}, low ${low}` : null,
      },
    };
  },
  sports: ({ data }) => {
    const team = isObject(data) ? text(data.team_name) : null;
    if (!isObject(data) || !team) return null;
    const icon = (typeof data.league === "string" && LEAGUE_ICONS[data.league]) || "";
    if (data.result === "upcoming") {
      return { renderer: "wideTile", data: { icon, label: team, word: "No games yet" } };
    }
    const word = typeof data.result === "string" ? RESULTS[data.result] : undefined;
    const { team_score, opponent_score } = data;
    if (!word || typeof team_score !== "number" || typeof opponent_score !== "number") return null;
    const opponent = text(data.opponent_team_name);
    return {
      renderer: "wideTile",
      data: { icon, label: team, word: `${word} ${team_score}–${opponent_score}`, detail: opponent && `vs ${opponent}` },
    };
  },
  news: ({ data }) => {
    const articles = isObject(data) && Array.isArray(data.articles) ? data.articles : [];
    const title = articles.map((article) => (isObject(article) ? text(article.title) : null)).find(Boolean);
    return title ? { renderer: "headline", data: { title } } : null;
  },
};

/**
 * What the grid draws for the server's widgets, around the day and time (which
 * the caller adds). Logs only a count of widgets it can't draw, never their content.
 */
export function placementsFor(screen: Screen): WidgetPlacement[] {
  let undrawable = 0;
  const drawn = createDrawCheck();
  const placements = screen.widgets.flatMap((widget): WidgetPlacement[] => {
    const lookFor = Object.prototype.hasOwnProperty.call(LOOKS, widget.type)
      ? LOOKS[widget.type as keyof typeof LOOKS]
      : undefined;
    if (!lookFor || !drawn(widget.type, widget)) {
      undrawable++;
      return [];
    }
    // An outside service being down is expected; the widget just isn't shown.
    if (widget.error) return [];
    const look = lookFor(widget);
    if (!look) {
      undrawable++;
      return [];
    }
    const { col, row, width, height } = widget;
    return [{ id: String(widget.id), widgetType: widget.type, col, row, width, height, ...look } as WidgetPlacement];
  });
  if (undrawable) console.warn(`Left out ${undrawable} widgets the tablet can't draw`);
  return placements;
}
