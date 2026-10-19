/**
 * time.ts
 *
 * User: Caregiver
 *
 * Calendar-only time helpers: the server's awake hours as @shared's
 * AwakeHours, and the server's items as @shared's PlayableItem so overlap is
 * checked by the same rule the tablet plays by. Also the server's bedtime rule
 * (an item must finish by bed time) and the time picker's options, which only
 * the caregiver needs. Parsing and start times are in @shared (parseClock,
 * formatClock, startTimeOptions).
 */

import { parseClock, startTimeOptions } from "@shared";
import type { AwakeHours, PlayableItem, PlayableKind } from "@shared";
import type { Occurrence, ScreenSettings, Shape } from "./api";

/** The server's "HH:MM" wake and bed times as minutes; null if either can't be read */
export function awakeHoursOf(settings: ScreenSettings): AwakeHours | null {
  const wakeMinute = parseClock(settings.wake_time);
  const bedMinute = parseClock(settings.bed_time);
  return wakeMinute === null || bedMinute === null ? null : { wakeMinute, bedMinute };
}

/**
 * Whether an item starting at `startMinute` and lasting `durationSeconds` is
 * allowed by the server: it starts at or after wake time and finishes by bed
 * time (`OUTSIDE_AWAKE_HOURS`). The shared rules only check the start, so this
 * is the stricter check the caregiver's save has to pass.
 */
export function finishesByBed(startMinute: number, durationSeconds: number, hours: AwakeHours): boolean {
  const startSeconds = startMinute * 60;
  return startSeconds >= hours.wakeMinute * 60 && startSeconds + durationSeconds <= hours.bedMinute * 60;
}

/**
 * The start times the form's time select offers for something that plays for
 * `durationSeconds`: each quarter hour while the carepartner is awake where it
 * still finishes by bed time.
 */
export function timeOptions(hours: AwakeHours, durationSeconds: number): number[] {
  return startTimeOptions(hours).filter((minute) => finishesByBed(minute, durationSeconds, hours));
}

// The server calls audio over a photo "audio_with_background"; @shared calls it "audioWithPhoto"
const PLAYABLE_KIND: Record<Shape, PlayableKind> = {
  video: "video",
  audio: "audio",
  audio_with_background: "audioWithPhoto",
  photo: "photo",
};

/**
 * An occurrence as @shared's overlap rule wants it; `id` is the scheduled
 * item's, so an edit can ignore itself. Null if the time can't be read.
 */
export function toPlayable(occurrence: Occurrence): (PlayableItem & { id: number }) | null {
  const startMinute = parseClock(occurrence.time);
  if (startMinute === null) return null;
  return {
    id: occurrence.scheduled_item_id,
    kind: PLAYABLE_KIND[occurrence.shape],
    date: occurrence.date,
    startMinute,
    durationSeconds: occurrence.duration_seconds,
  };
}
