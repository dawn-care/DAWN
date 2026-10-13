/**
 * nightHours.ts
 *
 * When the tablet treats it as night. Night hours belong to the carepartner's
 * schedule, which the caregiver sets and the server stores. Until the server
 * sends them, the tablet uses DEFAULT_NIGHT_HOURS.
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
