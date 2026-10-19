/**
 * Turns the server's date and time text into wording for people. Dates are
 * "YYYY-MM-DD" and times are "HH:MM" (24 hour), in the carepartner's local time.
 */

import { parseClock } from "../display/awakeHours";

/**
 * A "YYYY-MM-DD" date as a local date. `new Date("2026-10-03")` reads it as
 * UTC, which can land on the day before in the US.
 */
export function parseDate(date: string): Date {
  const [year, month, day] = date.split("-").map(Number);
  return new Date(year, month - 1, day);
}

/**
 * A 24 hour "HH:MM" as for example "8 AM" or "6:30 PM", leaving out ":00" so the
 * tablet has fewer characters to fit. Text that isn't a time is returned as it is.
 */
export function timeLabel(time: string): string {
  const minutes = parseClock(time);
  if (minutes == null) return time;
  const hour = Math.floor(minutes / 60);
  const minute = minutes % 60;
  const hour12 = hour % 12 || 12;
  const suffix = hour < 12 ? "AM" : "PM";
  return minute ? `${hour12}:${String(minute).padStart(2, "0")} ${suffix}` : `${hour12} ${suffix}`;
}

/** 1 is "1st", 2 is "2nd", 11 is "11th", 23 is "23rd". */
export function ordinal(n: number): string {
  const lastTwo = n % 100;
  if (lastTwo >= 11 && lastTwo <= 13) return `${n}th`;
  const suffix = ({ 1: "st", 2: "nd", 3: "rd" } as Record<number, string>)[n % 10] ?? "th";
  return `${n}${suffix}`;
}
