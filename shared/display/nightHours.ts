/**
 * nightHours.ts
 *
 * When the tablet treats it as night: from the carepartner's bed time until
 * their wake time, which the caregiver sets as awake hours and the server sends
 * with the screen. Until the tablet has them, it uses DEFAULT_NIGHT_HOURS.
 */

/** Night as minutes after midnight. It may run past midnight (start after end). */
export type NightHours = {
  startMinute: number;
  endMinute: number;
};

/** 20:00 until 7:00. */
export const DEFAULT_NIGHT_HOURS: NightHours = {
  startMinute: 20 * 60,
  endMinute: 7 * 60,
};

export function isNight(time: Date, hours: NightHours = DEFAULT_NIGHT_HOURS): boolean {
  const minute = time.getHours() * 60 + time.getMinutes();
  const { startMinute, endMinute } = hours;
  if (startMinute <= endMinute) return minute >= startMinute && minute < endMinute;
  return minute >= startMinute || minute < endMinute;
}

/** "HH:MM" as minutes after midnight, or null if it isn't a time. */
function minuteOf(time: string): number | null {
  const match = /^(\d{1,2}):(\d{2})$/.exec(time);
  if (!match) return null;
  const [hours, minutes] = [Number(match[1]), Number(match[2])];
  return hours < 24 && minutes < 60 ? hours * 60 + minutes : null;
}

/**
 * Night hours from awake hours ("HH:MM"): night starts at bed time and ends at
 * wake time. Falls back to DEFAULT_NIGHT_HOURS if either time isn't valid.
 */
export function nightHoursFromAwake(awake: { wake: string; bed: string }): NightHours {
  const startMinute = minuteOf(awake.bed);
  const endMinute = minuteOf(awake.wake);
  if (startMinute === null || endMinute === null) return DEFAULT_NIGHT_HOURS;
  return { startMinute, endMinute };
}
