/**
 * screen.ts
 *
 * The always-on screen as the server sends it (`GET /screen`, see
 * docs/api/screen-and-media.md), checked once here where it enters the app.
 * Logic only.
 */

import { readMedia, type MediaItem } from "../playback/media";

/** A widget in one spot of the layout. Renderers check `config` and `data` for their type (#316). */
export type ScreenWidget = {
  id: number;
  /** "photo", "message", "reminder", "weather", "sports" or "news"; unknown types are kept and skipped when drawn. */
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
  layoutPreset: string;
  awakeHours: { wake: string; bed: string };
  /** One entry per spot of the layout, in reading order; null is an empty spot. */
  spots: (ScreenWidget | null)[];
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
  const { photo } = value;
  if (photo != null && (!isObject(photo) || typeof photo.url !== "string")) return null;
  return {
    id: value.id,
    type: value.type,
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
 * The screen, or null if the response isn't one. A malformed widget becomes an
 * empty spot, so the others keep their places, and a malformed schedule entry
 * is dropped. Only counts are logged: widgets can hold a caregiver's messages.
 */
export function readScreen(raw: unknown): Screen | null {
  if (!isObject(raw) || !Array.isArray(raw.spots) || !Array.isArray(raw.schedule)) return null;
  const { awake_hours } = raw;
  if (typeof raw.layout_preset !== "string" || !isObject(awake_hours)) return null;
  if (typeof awake_hours.wake !== "string" || typeof awake_hours.bed !== "string") return null;

  const rawSpots: unknown[] = raw.spots;
  const spots = rawSpots.map((spot) => (spot === null ? null : readWidget(spot)));
  const badSpots = spots.filter((spot, i) => spot === null && rawSpots[i] !== null).length;
  if (badSpots) console.warn(`Dropped ${badSpots} malformed widgets from the screen`);

  const schedule = raw.schedule.map(readScheduledPlay).filter((play): play is ScheduledPlay => play !== null);
  const badPlays = raw.schedule.length - schedule.length;
  if (badPlays) console.warn(`Dropped ${badPlays} schedule entries the tablet can't play`);

  return {
    layoutPreset: raw.layout_preset,
    awakeHours: { wake: awake_hours.wake, bed: awake_hours.bed },
    spots,
    schedule,
  };
}

/**
 * The screen with live-data widgets (weather, sports, news) emptied once the
 * last good fetch is `STALE_AFTER_MS` old, so the screen never shows old
 * information as current.
 */
export function hideStaleWidgets(screen: Screen, fetchedAt: Date, time: Date): Screen {
  if (time.getTime() - fetchedAt.getTime() < STALE_AFTER_MS) return screen;
  return {
    ...screen,
    spots: screen.spots.map((spot) => (spot && LIVE_DATA_TYPES.includes(spot.type) ? null : spot)),
  };
}
