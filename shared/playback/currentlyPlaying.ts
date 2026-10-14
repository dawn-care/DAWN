/**
 * currentlyPlaying.ts
 *
 * What the carepartner's screen should be showing at a given moment: her usual
 * screen, a video, audio over her usual screen, audio over one chosen photo, a
 * timed photo, or the night screen. The caregiver's preview and the tablet both
 * ask this, and the preview can pass any time to test a schedule.
 *
 * The time and the schedule are arguments; nothing here reads the clock. `now`
 * and each item's date and start are read on the same wall clock, the
 * carepartner's. An item only has to START in the awake hours. Once it is
 * playing it finishes, even past bedtime. An item that starts outside the awake
 * hours never plays.
 */

import { DEFAULT_AWAKE_HOURS, isAwakeAt, minuteOfDay } from "../display/awakeHours";
import type { AwakeHours } from "../display/awakeHours";

/**
 * What an item shows when it plays:
 * - video: full screen
 * - audio: over her usual screen, with its title
 * - audioWithPhoto: over one chosen photo, full screen
 * - photo: one photo, full screen, for a set time
 */
export type PlayableKind = "video" | "audio" | "audioWithPhoto" | "photo";

/** A scheduled item on one day, as far as deciding what plays is concerned. */
export type PlayableItem = {
  kind: PlayableKind;
  /** The day it plays, as "YYYY-MM-DD". */
  date: string;
  /** When it starts, in minutes after midnight. */
  startMinute: number;
  /** How long it plays in seconds: the media's length, or how long a photo stays up. */
  durationSeconds: number;
};

export type Playback<T extends PlayableItem = PlayableItem> =
  /** Her usual screen. */
  | { kind: "static" }
  /** Outside the awake hours with nothing playing. */
  | { kind: "night" }
  /** An item is playing; its kind says how. */
  | {
      kind: PlayableKind;
      item: T;
      /** Whole seconds since it started. */
      elapsedSeconds: number;
      /** Whole seconds until it ends. */
      remainingSeconds: number;
    };

const DATE_PATTERN = /^(\d{4})-(\d{2})-(\d{2})$/;
const SECONDS_PER_MINUTE = 60;

/**
 * Seconds from a fixed point on the wall clock, ignoring time zones and clock
 * changes, so a date and minute compare exactly with another date and minute.
 */
function wallSeconds(year: number, month: number, day: number, secondOfDay: number): number {
  return Date.UTC(year, month, day) / 1000 + secondOfDay;
}

/** The item's start on the wall clock, or null when its date or times can't be used. */
function startSeconds(item: PlayableItem): number | null {
  const match = DATE_PATTERN.exec(item.date);
  if (!match) return null;
  if (!Number.isInteger(item.startMinute) || item.startMinute < 0 || item.startMinute >= 24 * 60) {
    return null;
  }
  const year = Number(match[1]);
  const month = Number(match[2]) - 1;
  const day = Number(match[3]);
  const start = Date.UTC(year, month, day);
  // Dates like "2026-02-30" roll over into the next month; they are not real days.
  const real = new Date(start);
  if (real.getUTCFullYear() !== year || real.getUTCMonth() !== month || real.getUTCDate() !== day) {
    return null;
  }
  return start / 1000 + item.startMinute * SECONDS_PER_MINUTE;
}

/**
 * What should be on screen at `now`.
 *
 * Playing items come first: the one that started most recently if more than one
 * is under way. With nothing playing, it is the night screen outside the awake
 * hours and her usual screen inside them.
 */
export function currentlyPlaying<T extends PlayableItem>(
  now: Date,
  schedule: readonly T[],
  awakeHours: AwakeHours = DEFAULT_AWAKE_HOURS,
): Playback<T> {
  const nowSeconds = wallSeconds(
    now.getFullYear(),
    now.getMonth(),
    now.getDate(),
    minuteOfDay(now) * SECONDS_PER_MINUTE + now.getSeconds(),
  );

  let current: { item: T; start: number } | null = null;
  for (const item of schedule) {
    if (!isAwakeAt(item.startMinute, awakeHours)) continue;
    if (!(item.durationSeconds > 0)) continue;
    const start = startSeconds(item);
    if (start === null) continue;
    if (nowSeconds < start || nowSeconds >= start + item.durationSeconds) continue;
    if (current === null || start > current.start) current = { item, start };
  }

  if (current !== null) {
    const elapsedSeconds = Math.floor(nowSeconds - current.start);
    return {
      kind: current.item.kind,
      item: current.item,
      elapsedSeconds,
      remainingSeconds: Math.ceil(current.item.durationSeconds - elapsedSeconds),
    };
  }

  return isAwakeAt(minuteOfDay(now), awakeHours) ? { kind: "static" } : { kind: "night" };
}
