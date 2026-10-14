/**
 * schedule.ts
 *
 * Decides what the schedule wants to play next. Today's entries come from the
 * server's configurations, and there are two kinds:
 * - Timed entries play at their time.
 * - Untimed entries are the daily routine, played in order once the
 *   carepartner's auto-start time arrives.
 *
 * Pure functions only; ScheduleProvider supplies the time and the state.
 */

import { isPlayable, type MediaItem } from "../playback/media";

/** A configuration from `/configurations/day`: the fields the schedule uses. */
export type ScheduleEntry = MediaItem & {
  id: number;
  timed: boolean;
  /** "HH:MM:SS" for a timed entry. */
  time: string | null;
  /** Order within the routine. */
  position: number;
};

/**
 * Keeps the entries with the fields the schedule needs, dropping malformed
 * ones with a warning, and warns about entries the tablet can't play. A null
 * list is an empty day, since Go sends an empty list as null.
 */
export function readEntries(raw: unknown): ScheduleEntry[] {
  if (raw === null) return [];
  if (!Array.isArray(raw)) {
    console.warn("Schedule response is not a list; treating it as empty");
    return [];
  }
  const entries = raw.filter(
    (entry): entry is ScheduleEntry =>
      typeof entry === "object" &&
      entry !== null &&
      typeof entry.id === "number" &&
      typeof entry.module_type === "string" &&
      typeof entry.position === "number",
  );
  // Only the count: an entry can hold a caregiver's message text.
  const dropped = raw.length - entries.length;
  if (dropped > 0) console.warn(`Dropped ${dropped} malformed schedule entries`);
  // Kept, since `nextEntry` skips them; warned here once per fetch rather than
  // on every pass of `nextEntry`.
  const unplayable = entries.filter((entry) => !isPlayable(entry)).map((entry) => entry.id);
  if (unplayable.length) console.warn(`Skipping schedule entries the tablet can't play: ${unplayable.join(", ")}`);
  return entries;
}

/** "HH:MM" for a time of day, the form entry times and auto-start are matched in. */
export function minuteOf(time: Date): string {
  return `${String(time.getHours()).padStart(2, "0")}:${String(time.getMinutes()).padStart(2, "0")}`;
}

/** "YYYY-MM-DD" for the local day. */
export function dayOf(time: Date): string {
  const month = String(time.getMonth() + 1).padStart(2, "0");
  const day = String(time.getDate()).padStart(2, "0");
  return `${time.getFullYear()}-${month}-${day}`;
}

/** Timed entries whose time is this minute and that haven't played today. */
export function timedEntriesDueAt(entries: ScheduleEntry[], minute: string, played: Set<number>): number[] {
  return entries
    .filter((entry) => entry.timed && entry.time?.slice(0, 5) === minute && !played.has(entry.id))
    .map((entry) => entry.id);
}

/** Whether the routine should start this minute: it's auto-start time and something in it hasn't played. */
export function routineStartsAt(
  autoStart: string | undefined,
  minute: string,
  entries: ScheduleEntry[],
  played: Set<number>,
): boolean {
  if (autoStart?.slice(0, 5) !== minute) return false;
  return entries.some((entry) => !entry.timed && isPlayable(entry) && !played.has(entry.id));
}

/**
 * The entry to play next, or null. Timed entries that have come due go first,
 * earliest first; then, while the routine is running, the routine in order.
 * Entries the tablet can't play are skipped.
 */
export function nextEntry(
  entries: ScheduleEntry[],
  { played, due, routineRunning }: { played: Set<number>; due: Set<number>; routineRunning: boolean },
): ScheduleEntry | null {
  const waiting = entries.filter((entry) => !played.has(entry.id) && isPlayable(entry));

  const timed = waiting
    .filter((entry) => entry.timed && due.has(entry.id))
    .sort((a, b) => (a.time ?? "").localeCompare(b.time ?? ""));
  if (timed.length) return timed[0];

  if (!routineRunning) return null;
  const routine = waiting.filter((entry) => !entry.timed).sort((a, b) => a.position - b.position);
  return routine[0] ?? null;
}
