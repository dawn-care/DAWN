/**
 * screen.ts
 *
 * The always-on screen as the server sends it (`GET /screen`, see
 * docs/api/screen-and-media.md), checked once here where it enters the app.
 * Logic only.
 */

import type { GridArea } from "@shared";
import { readMedia, type MediaItem } from "../playback/media";

/** A widget and its place on the grid. `placementsFor` checks `config` and `data` for its type. */
export type ScreenWidget = GridArea & {
  id: number;
  /** "photo", "message", "reminder", "weather", "sports" or "news"; unknown types (a stored "day_time" too, since the tablet draws its own) are kept and skipped when drawn. */
  type: string;
  config: unknown;
  photo: { url: string; displayName: string } | null;
  data: unknown;
  /** "UPSTREAM_UNAVAILABLE" when an outside service is down. */
  error: string | null;
};

/** One time a scheduled item plays. */
export type ScheduledPlay = {
  scheduledItemId: number;
  /** "YYYY-MM-DD" */
  date: string;
  /** "HH:MM" */
  time: string;
  /** The caregiver moved awake hours so this falls outside them. */
  outsideAwakeHours: boolean;
  item: MediaItem;
};

export type Screen = {
  awakeHours: { wake: string; bed: string };
  /** Whether to chime before each play. Staff can switch it off per resident (#387); on unless the server says otherwise. */
  chime: boolean;
  /** In reading order. Cells no widget covers are empty. */
  widgets: ScreenWidget[];
  /** Today and tomorrow, so playback carries on past midnight. */
  schedule: ScheduledPlay[];
};

/** Widget types whose data comes from an outside service and goes out of date. */
const LIVE_DATA_TYPES = ["weather", "sports", "news"];

/** How long live data is shown after the last good fetch before its widget hides. */
export const STALE_AFTER_MS = 3 * 60 * 60_000;

const isObject = (value: unknown): value is Record<string, unknown> => typeof value === "object" && value !== null;
const HH_MM = /^\d{2}:\d{2}$/;
const DATE = /^\d{4}-\d{2}-\d{2}$/;

function readWidget(value: unknown): ScreenWidget | null {
  if (!isObject(value) || typeof value.id !== "number" || typeof value.type !== "string") return null;
  const { col, row, width, height, photo } = value;
  if (![col, row, width, height].every((n) => typeof n === "number")) return null;
  if (photo != null && (!isObject(photo) || typeof photo.url !== "string")) return null;
  return {
    id: value.id,
    type: value.type,
    col: col as number,
    row: row as number,
    width: width as number,
    height: height as number,
    config: value.config ?? null,
    photo: isObject(photo)
      ? { url: photo.url as string, displayName: typeof photo.display_name === "string" ? photo.display_name : "" }
      : null,
    data: value.data ?? null,
    error: typeof value.error === "string" ? value.error : null,
  };
}

function readScheduledPlay(value: unknown): ScheduledPlay | null {
  if (!isObject(value)) return null;
  const { scheduled_item_id, date, time, background_photo, duration_seconds } = value;
  const media = readMedia(value.media);
  if (typeof scheduled_item_id !== "number" || typeof date !== "string" || !DATE.test(date)) return null;
  if (typeof time !== "string" || !HH_MM.test(time) || !media) return null;
  const cover = isObject(background_photo) && typeof background_photo.url === "string" ? background_photo.url : undefined;
  return {
    scheduledItemId: scheduled_item_id,
    date,
    time,
    outsideAwakeHours: value.outside_awake_hours === true,
    item: {
      ...media,
      ...(media.kind === "audio" && cover ? { coverUrl: cover } : {}),
      ...(media.kind === "photo" && typeof duration_seconds === "number" ? { durationSeconds: duration_seconds } : {}),
    },
  };
}

/**
 * The screen, or null if the response isn't one. A malformed widget or
 * schedule entry is dropped, and the rest kept. Only counts are logged:
 * widgets can hold a caregiver's messages.
 */
export function readScreen(raw: unknown): Screen | null {
  if (!isObject(raw) || !Array.isArray(raw.widgets) || !Array.isArray(raw.schedule)) return null;
  const { awake_hours } = raw;
  if (!isObject(awake_hours) || typeof awake_hours.wake !== "string" || typeof awake_hours.bed !== "string") return null;

  const widgets = raw.widgets.map(readWidget).filter((widget): widget is ScreenWidget => widget !== null);
  const badWidgets = raw.widgets.length - widgets.length;
  if (badWidgets) console.warn(`Dropped ${badWidgets} malformed widgets from the screen`);

  const schedule = raw.schedule.map(readScheduledPlay).filter((play): play is ScheduledPlay => play !== null);
  const badPlays = raw.schedule.length - schedule.length;
  if (badPlays) console.warn(`Dropped ${badPlays} schedule entries the tablet can't play`);

  return {
    awakeHours: { wake: awake_hours.wake, bed: awake_hours.bed },
    chime: raw.chime !== false,
    widgets,
    schedule,
  };
}

/**
 * The screen with live-data widgets (weather, sports, news) removed once the
 * last good fetch is `STALE_AFTER_MS` old, so the screen never shows old
 * information as current.
 */
export function hideStaleWidgets(screen: Screen, fetchedAt: Date, time: Date): Screen {
  if (time.getTime() - fetchedAt.getTime() < STALE_AFTER_MS) return screen;
  return {
    ...screen,
    widgets: screen.widgets.filter((widget) => !LIVE_DATA_TYPES.includes(widget.type)),
  };
}
