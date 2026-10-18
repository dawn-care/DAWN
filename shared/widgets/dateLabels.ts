/**
 * Turns the server's date and time text into wording for people. Dates are
 * "YYYY-MM-DD" and times are "HH:MM" (24 hour), in the carepartner's local time.
 */

/**
 * A "YYYY-MM-DD" date as a local date. `new Date("2026-10-03")` reads it as
 * UTC, which can land on the day before in the US.
 */
export function parseDate(date: string): Date {
  const [year, month, day] = date.split("-").map(Number);
  return new Date(year, month - 1, day);
}

/** A 24 hour "HH:MM" as for example "8:00 am" or "6:30 pm". */
export function timeLabel(time: string): string {
  const [hour, minute] = time.split(":").map(Number);
  const hour12 = ((hour + 11) % 12) + 1;
  return `${hour12}:${String(minute).padStart(2, "0")} ${hour < 12 ? "am" : "pm"}`;
}

/** 1 is "1st", 2 is "2nd", 11 is "11th", 23 is "23rd". */
export function ordinal(n: number): string {
  const lastTwo = n % 100;
  if (lastTwo >= 11 && lastTwo <= 13) return `${n}th`;
  const suffix = ({ 1: "st", 2: "nd", 3: "rd" } as Record<number, string>)[n % 10] ?? "th";
  return `${n}${suffix}`;
}
