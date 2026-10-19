/**
 * formRules.ts
 *
 * User: Caregiver
 *
 * The rules behind the schedule form, with no React in them: what the form
 * starts with for a new item or an existing one, the Repeats wording for a
 * date, why a time won't play, and the request body a filled-in form becomes.
 */

import { dayName, formatClock, isAwakeAt, ordinal, parseClock, parseDate } from "@shared";
import type { AwakeHours } from "@shared";
import type { Media } from "../media/api";
import type { ItemInput, Repeat, ScheduledItem } from "./api";
import { finishesByBed, minuteLabel, timeOptions } from "./time";

export type FormKind = "video" | "audio";

/** Whether an audio file shows a photo while it plays ("photo"), or just the tablet's music note ("usual") */
export type ScreenChoice = "usual" | "photo";

export interface FormValues {
  kind: FormKind;
  mediaId: number | null;
  /** "YYYY-MM-DD" */
  date: string;
  /** Minutes after midnight */
  minute: number;
  repeat: Repeat;
  screen: ScreenChoice;
  photoId: number | null;
}

/** The length of a video or audio file in seconds; 0 if unknown */
export function lengthOf(media: Pick<Media, "duration_seconds"> | undefined): number {
  return media?.duration_seconds ?? 0;
}

/** The start times the form offers: where the media fits in the awake hours, plus the one already chosen */
export function minuteOptions(hours: AwakeHours, durationSeconds: number, chosen: number): number[] {
  const options = timeOptions(hours, durationSeconds);
  if (!options.includes(chosen)) options.push(chosen);
  return options.sort((a, b) => a - b);
}

/** The label for a time in the select; one that won't play says so */
export function minuteOptionLabel(minute: number, hours: AwakeHours, durationSeconds: number): string {
  return finishesByBed(minute, durationSeconds, hours) ? minuteLabel(minute) : `${minuteLabel(minute)} (won't play)`;
}

/** Null if the time works; otherwise a sentence saying why it won't play */
export function whyWontPlay(minute: number, durationSeconds: number, hours: AwakeHours): string | null {
  if (finishesByBed(minute, durationSeconds, hours)) return null;
  const advice = "Pick a time while the carepartner is awake, or change the awake hours.";
  if (minute < hours.wakeMinute) return `This is before the carepartner wakes up, so it won't play. ${advice}`;
  if (!isAwakeAt(minute, hours)) return `This is after the carepartner's bedtime, so it won't play. ${advice}`;
  return `This would still be playing at bedtime, so it won't play. Pick an earlier time, or change the awake hours.`;
}

/** The Repeats choices for a date: the weekday and the day of the month come from it */
export function repeatOptions(date: string): { value: Repeat; label: string }[] {
  const day = parseDate(date);
  return [
    { value: "once", label: "Does not repeat" },
    { value: "daily", label: "Every day" },
    { value: "weekly", label: `Every week on ${dayName(day)}` },
    { value: "monthly", label: `Every month on the ${ordinal(day.getDate())}` },
  ];
}

/** Monthly repeats skip months that are too short for the day */
export function monthlyHint(date: string): string | null {
  const day = parseDate(date).getDate();
  return day > 28 ? `Months without a ${ordinal(day)} are skipped.` : null;
}

/** The form for a new item: a video that doesn't repeat. No media is chosen yet; the form picks the first one listed */
export function valuesForNew(date: string, minute: number): FormValues {
  return { kind: "video", mediaId: null, date, minute, repeat: "once", screen: "usual", photoId: null };
}

/** The form for an existing item, opened from its occurrence on `date`; null if it isn't a video or audio */
export function valuesForEdit(item: ScheduledItem, date: string): FormValues | null {
  const kind = item.media.kind;
  const minute = parseClock(item.start_time);
  if ((kind !== "video" && kind !== "audio") || minute === null) return null;
  return {
    kind,
    mediaId: item.media_id,
    date,
    minute,
    repeat: item.repeat,
    screen: item.background_photo_id === null ? "usual" : "photo",
    photoId: item.background_photo_id,
  };
}

/**
 * The request body for a filled-in form. Editing replaces the whole series, so
 * it keeps the series' own start date and end date unless the caregiver chose a
 * different date than the occurrence they opened.
 */
export function toInput(
  values: FormValues,
  carepartnerId: number,
  editing: { item: ScheduledItem; occurrenceDate: string } | null
): ItemInput {
  const keepsSeries = editing !== null && values.date === editing.occurrenceDate;
  const repeats = values.repeat !== "once";
  return {
    carepartner_id: carepartnerId,
    media_id: values.mediaId as number,
    background_photo_id: values.kind === "audio" && values.screen === "photo" ? values.photoId : null,
    display_seconds: null,
    start_date: keepsSeries ? editing.item.start_date : values.date,
    start_time: formatClock(values.minute) ?? "00:00",
    repeat: values.repeat,
    end_date: repeats && keepsSeries ? editing.item.end_date : null,
  };
}

/**
 * The time the Schedule button starts the form at: the next quarter hour from
 * `nowMinute` (null when the form is for another day) if that is still awake
 * with room before bed, otherwise wake up time.
 */
export function nextOpenMinute(hours: AwakeHours, nowMinute: number | null): number {
  if (nowMinute === null) return hours.wakeMinute;
  const next = Math.ceil(nowMinute / 15) * 15;
  return next >= hours.wakeMinute && next < hours.bedMinute ? next : hours.wakeMinute;
}
