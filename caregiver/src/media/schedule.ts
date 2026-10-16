/**
 * schedule.ts
 *
 * User: Caregiver
 *
 * Turns scheduled items into the plain-English line shown on a video's or
 * audio's detail panel, e.g. "Every day at 8:00 am". Day and date names
 * come from @shared so they match the rest of the app.
 */

import { dateLabel, dayName, ordinal, parseDate, timeLabel } from "@shared";
import type { ScheduledItem } from "./api";

/** One item, e.g. "Every week on Sunday at 6:00 pm" or "Saturday, October 3 at 10:00 am" */
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

/**
 * The line for everything scheduled for one media item: the earliest
 * schedule, plus "and N more" when there are others. Null when nothing is
 * scheduled for it.
 */
export function describeSchedule(items: ScheduledItem[], mediaId: number): string | null {
  const mine = items
    .filter((item) => item.media_id === mediaId)
    .sort((a, b) => `${a.start_date} ${a.start_time}`.localeCompare(`${b.start_date} ${b.start_time}`));

  if (mine.length === 0) return null;
  const first = describeScheduledItem(mine[0]);
  return mine.length === 1 ? first : `${first} and ${mine.length - 1} more`;
}
