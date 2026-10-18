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
 *
 * Where this differs from the tablet today (carepartner/src/playback): the
 * tablet never interrupts an item that is playing, only starts a timed item in
 * its exact minute, and plays untimed routine entries in order. Here, asking for
 * any moment finds the item under way at it. The caregiver avoids the first
 * difference by refusing to schedule overlapping items (`overlappingItems`), so
 * the preview and the tablet never disagree about which item wins. The other two
 * stay until the tablet adopts this file.
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
 * When an item is on screen, in wall clock seconds from its start up to but not
 * including its end. Null for an item that never plays: one that starts outside
 * the awake hours, has no length, or has a date or start that can't be used.
 */
function playWindow(item: PlayableItem, awakeHours: AwakeHours): { start: number; end: number } | null {
  if (!isAwakeAt(item.startMinute, awakeHours)) return null;
  if (!(item.durationSeconds > 0)) return null;
  const start = startSeconds(item);
  if (start === null) return null;
  return { start, end: start + item.durationSeconds };
}

/**
 * The scheduled items that would be on screen at the same time as `candidate`.
 * Two items overlap when one starts before the other ends; an item that starts
 * the moment another ends does not overlap it. The caregiver uses this to refuse
 * a new item that would play over another. Items that never play are ignored.
 * When editing an item, pass its id as `ignoreId` so the schedule's old copy of
 * it doesn't count as overlapping itself.
 */
export function overlappingItems<T extends PlayableItem & { id?: string | number }>(
  candidate: PlayableItem,
  schedule: readonly T[],
  ignoreId?: string | number,
  awakeHours: AwakeHours = DEFAULT_AWAKE_HOURS,
): T[] {
  const mine = playWindow(candidate, awakeHours);
  if (mine === null) return [];
  return schedule.filter((item) => {
    if (ignoreId !== undefined && item.id === ignoreId) return false;
    const theirs = playWindow(item, awakeHours);
    return theirs !== null && mine.start < theirs.end && theirs.start < mine.end;
  });
}

/**
 * What should be on screen at `now`.
 *
 * Playing items come first. If more than one is under way, which the caregiver
 * prevents by scheduling (`overlappingItems`), the one that started most
 * recently wins. With nothing playing, it is the night screen outside the awake
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
    const window = playWindow(item, awakeHours);
    if (window === null) continue;
    if (nowSeconds < window.start || nowSeconds >= window.end) continue;
    if (current === null || window.start > current.start) current = { item, start: window.start };
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
