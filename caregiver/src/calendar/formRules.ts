/**
 * formRules.ts
 *
 * User: Caregiver
 *
 * The rules behind the schedule form, with no React in them: what the form
 * starts with for a new item or an existing one, the Repeats wording for a
 * date, why a time won't play, and the request body a filled-in form becomes.
 */

import { dayName, formatClock, isAwakeAt, ordinal, parseClock, parseDate } from "@shared";import type { AwakeHours } from "@shared";
import type { Media } from "../media/api";
import type { ItemInput, Repeat, ScheduledItem } from "./api";
import { dateText, finishesByBed, minuteLabel, timeOptions } from "./time";

export type FormKind = "video" | "audio";

/** The item being edited, and the date of the occurrence it was opened from */
export interface EditTarget {
  item: ScheduledItem;
  occurrenceDate: string;
}

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

/** The form for a new item that doesn't repeat: a video by default, with the first one listed chosen unless `media` says which */
export function valuesForNew(date: string, minute: number, media?: { kind: FormKind; id: number }): FormValues {
  return { kind: media?.kind ?? "video", mediaId: media?.id ?? null, date, minute, repeat: "once", screen: "usual", photoId: null };
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
 * The request body for a filled-in form with its file chosen. Editing replaces
 * the whole series, so it keeps the series' own start date and end date while
 * the date and Repeats are both as they were. Changing either starts the item
 * again from the date in the form, with no end date: the server works out
 * weekly and monthly repeats from the start date, and a one-time item plays
 * only on it, so keeping the old start date would save a different schedule
 * than the form shows.
 */
export function toInput(values: FormValues & { mediaId: number }, carepartnerId: number, editing: EditTarget | null): ItemInput {
  const keepsSeries = editing !== null && values.date === editing.occurrenceDate && values.repeat === editing.item.repeat;
  const repeats = values.repeat !== "once";
  const startTime = formatClock(values.minute);
  if (startTime === null) throw new Error(`Can't save a start time of minute ${values.minute}`);
  return {
    carepartner_id: carepartnerId,
    media_id: values.mediaId,
    background_photo_id: values.kind === "audio" && values.screen === "photo" ? values.photoId : null,
    display_seconds: null,
    start_date: keepsSeries ? editing.item.start_date : values.date,
    start_time: startTime,
    repeat: values.repeat,
    end_date: repeats && keepsSeries ? editing.item.end_date : null,
  };
}

/** The day after a "YYYY-MM-DD" date */
function nextDate(date: string): string {
  const day = parseDate(date);
  day.setDate(day.getDate() + 1);
  return dateText(day);
}

/**
 * Where a new form starts: on `date` (or today, if `date` has passed) at the
 * first start time where something `durationSeconds` long finishes by bed.
 * Today it is the first of those from `nowMinute` on, and once none is left it
 * is tomorrow, so the form never opens on a time that has already gone by. If
 * nothing that long fits at all, it starts at wake up and the form says why it
 * won't play.
 */
export function firstOpenSlot(
  hours: AwakeHours,
  durationSeconds: number,
  date: string,
  today: string,
  nowMinute: number
): { date: string; minute: number } {
  const options = timeOptions(hours, durationSeconds);
  if (options.length === 0) return { date: date < today ? today : date, minute: hours.wakeMinute };
  if (date > today) return { date, minute: options[0] };
  const later = options.find((minute) => minute >= nowMinute);
  return later === undefined ? { date: nextDate(today), minute: options[0] } : { date: today, minute: later };
}
