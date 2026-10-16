export type TimeOfDay = "morning" | "afternoon" | "evening";

const DAYS = ["Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"];
const MONTHS = [
  "January", "February", "March", "April", "May", "June",
  "July", "August", "September", "October", "November", "December",
];

/** Morning from 5:00, afternoon from 12:00, evening from 17:00 until 5:00. */
export function timeOfDay(date: Date): TimeOfDay {
  const hour = date.getHours();
  if (hour >= 5 && hour < 12) return "morning";
  if (hour >= 12 && hour < 17) return "afternoon";
  return "evening";
}

export function dayName(date: Date): string {
  return DAYS[date.getDay()];
}

/** For example "October 5". */
export function dateLabel(date: Date): string {
  return `${MONTHS[date.getMonth()]} ${date.getDate()}`;
}
