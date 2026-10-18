/**
 * schedule.ts
 *
 * Decides which scheduled play is due. The schedule comes from `/screen`
 * (today and tomorrow, see `ScheduledPlay`).
 *
 * Pure functions only; SchedulePlayer supplies the time and what has played.
 */

import type { ScheduledPlay } from "../screen/screen";

/**
 * How late a scheduled play may still start. A play that couldn't start in
 * time, because the tablet was off, restarting or busy with a staff clip, is
 * skipped rather than started long after it was meant to.
 */
export const GRACE_MS = 10 * 60_000;

/** Identifies one day's play of a scheduled item, for remembering what has played. */
export function playKey(play: ScheduledPlay): string {
  return `${play.scheduledItemId}@${play.date}`;
}

/** When the play is meant to start, in the tablet's local time. */
export function startOf(play: ScheduledPlay): Date {
  const [year, month, day] = play.date.split("-").map(Number);
  const [hours, minutes] = play.time.split(":").map(Number);
  return new Date(year, month - 1, day, hours, minutes);
}

/**
 * The play to start now, or null: one whose start has come within the last
 * `GRACE_MS` and that hasn't played, earliest first. Plays outside the
 * carepartner's awake hours are skipped, since they would wake them.
 */
export function duePlay(schedule: ScheduledPlay[], time: Date, played: Set<string>): ScheduledPlay | null {
  const due = schedule
    .filter((play) => {
      const late = time.getTime() - startOf(play).getTime();
      return late >= 0 && late < GRACE_MS && !play.outsideAwakeHours && !played.has(playKey(play));
    })
    .sort((a, b) => startOf(a).getTime() - startOf(b).getTime());
  return due[0] ?? null;
}
