/**
 * awakeHours.ts
 *
 * The hours a carepartner is awake. Videos and audio start only inside them,
 * and the screen is asleep outside them. The caregiver sets the wake and bed
 * times and the server stores them in the screen settings.
 *
 * Times are minutes after midnight. An item only has to START inside the awake
 * hours; once it is playing it may run past bedtime.
 */

/** Wake and bed as minutes after midnight. Awake from wake up to, but not including, bed. */
export type AwakeHours = {
  wakeMinute: number;
  bedMinute: number;
};

/** Items start on these steps. */
export const START_STEP_MINUTES = 15;

/** 8 am until 10 pm. */
export const DEFAULT_AWAKE_HOURS: AwakeHours = {
  wakeMinute: 8 * 60,
  bedMinute: 22 * 60,
};

const MINUTES_PER_DAY = 24 * 60;

/** Minutes after midnight, in the time's own clock. */
export function minuteOfDay(time: Date): number {
  return time.getHours() * 60 + time.getMinutes();
}

/**
 * Minutes after midnight for the server's time text: hours and minutes, with or
 * without a leading zero on the hour or seconds on the end.
 * Null when the text is not a time of day.
 */
export function parseClock(text: string): number | null {
  const match = /^(\d{1,2}):(\d{2})(?::\d{2})?$/.exec(text.trim());
  if (!match) return null;
  const hours = Number(match[1]);
  const minutes = Number(match[2]);
  if (hours > 23 || minutes > 59) return null;
  return hours * 60 + minutes;
}

/**
 * The server's time text for a minute after midnight: two digit hours, a colon,
 * two digit minutes. Null unless the minute is a whole number from 0 to 1439,
 * the same range `parseClock` accepts.
 */
export function formatClock(minute: number): string | null {
  if (!Number.isInteger(minute) || minute < 0 || minute > 1439) return null;
  const hours = Math.floor(minute / 60);
  const minutes = minute % 60;
  return `${String(hours).padStart(2, "0")}:${String(minutes).padStart(2, "0")}`;
}

/** Whether this minute is awake: from wake up to, but not including, bed. */
export function isAwakeAt(minute: number, hours: AwakeHours = DEFAULT_AWAKE_HOURS): boolean {
  return minute >= hours.wakeMinute && minute < hours.bedMinute;
}

export function isAwake(time: Date, hours: AwakeHours = DEFAULT_AWAKE_HOURS): boolean {
  return isAwakeAt(minuteOfDay(time), hours);
}

/** Whether the wake and bed times can be saved: whole minutes in one day, with bed after wake. */
export function isValidAwakeHours(hours: AwakeHours): boolean {
  const { wakeMinute, bedMinute } = hours;
  return (
    Number.isInteger(wakeMinute) &&
    Number.isInteger(bedMinute) &&
    wakeMinute >= 0 &&
    bedMinute < MINUTES_PER_DAY &&
    wakeMinute < bedMinute
  );
}

/** Whether a start time is on a 15 minute step. */
export function isOnStep(minute: number): boolean {
  return Number.isInteger(minute) && minute % START_STEP_MINUTES === 0;
}

/** The nearest 15 minute step, halfway rounding up, kept inside the same day. */
export function snapToStep(minute: number): number {
  const snapped = Math.round(minute / START_STEP_MINUTES) * START_STEP_MINUTES;
  return Math.min(Math.max(snapped, 0), MINUTES_PER_DAY - START_STEP_MINUTES);
}

/** Whether an item may start at this minute: on a 15 minute step and awake. */
export function isStartAllowed(minute: number, hours: AwakeHours = DEFAULT_AWAKE_HOURS): boolean {
  return isOnStep(minute) && isAwakeAt(minute, hours);
}

/** Every start time the time picker offers: each 15 minutes from wake up until just before bed. */
export function startTimeOptions(hours: AwakeHours = DEFAULT_AWAKE_HOURS): number[] {
  const options: number[] = [];
  const first = Math.ceil(hours.wakeMinute / START_STEP_MINUTES) * START_STEP_MINUTES;
  for (let minute = first; minute < hours.bedMinute; minute += START_STEP_MINUTES) {
    options.push(minute);
  }
  return options;
}

/**
 * The items that start outside the awake hours, in their original order. Used
 * for the warning before new hours are saved: those items won't play.
 */
export function itemsOutsideHours<T extends { startMinute: number }>(
  items: readonly T[],
  hours: AwakeHours,
): T[] {
  return items.filter((item) => !isAwakeAt(item.startMinute, hours));
}
