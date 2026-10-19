/**
 * schedule.ts
 *
 * User: Caregiver
 *
 * Turns scheduled items into the plain-English line shown on a media item's
 * detail panel, e.g. "Every day at 8 AM". Day and date names come from
 * @shared so they match the rest of the app.
 */

import { dateLabel, dayName, ordinal, parseDate, timeLabel } from "@shared";
import type { ScheduledItem } from "./api";

/** One item, e.g. "Every week on Sunday at 6 PM" or "Saturday, October 3 at 10 AM" */
export function describeScheduledItem(item: ScheduledItem): string {
  const start = parseDate(item.start_date);
  const time = timeLabel(item.start_time);

  let text: string;
  switch (item.repeat) {
    case "daily":
      text = `Every day at ${time}`;
      break;
    case "weekly":
      text = `Every week on ${dayName(start)} at ${time}`;
      break;
    case "monthly":
      text = `Every month on the ${ordinal(start.getDate())} at ${time}`;
      break;
    default:
      text = `${dayName(start)}, ${dateLabel(start)} at ${time}`;
  }

  if (item.repeat !== "once" && item.end_date) {
    text += `, until ${dateLabel(parseDate(item.end_date))}`;
  }
  return text;
}

// "YYYY-MM-DD" in local time, comparable with the server's date text
function todayText(now: Date): string {
  const month = String(now.getMonth() + 1).padStart(2, "0");
  const day = String(now.getDate()).padStart(2, "0");
  return `${now.getFullYear()}-${month}-${day}`;
}

// Whether the item can still play: a one-time item that has passed or been skipped,
// or a repeating one whose end date has passed, is finished
function isUpcoming(item: ScheduledItem, today: string): boolean {
  if (item.repeat === "once") return item.start_date >= today && !item.skip_dates.includes(item.start_date);
  return item.end_date === null || item.end_date >= today;
}

/**
 * The line for everything still scheduled for one media item: the earliest
 * schedule, plus "and N more" when there are others. Finished schedules are
 * left out. Null when nothing is scheduled for it. `now` is only for telling
 * what has finished.
 */
export function describeSchedule(items: ScheduledItem[], mediaId: number, now: Date = new Date()): string | null {
  const today = todayText(now);
  const mine = items
    .filter((item) => item.media_id === mediaId && isUpcoming(item, today))
    .sort((a, b) => {
      const first = `${a.start_date} ${a.start_time}`;
      const second = `${b.start_date} ${b.start_time}`;
      return first < second ? -1 : first > second ? 1 : 0;
    });

  if (mine.length === 0) return null;
  const first = describeScheduledItem(mine[0]);
  return mine.length === 1 ? first : `${first} and ${mine.length - 1} more`;
}
