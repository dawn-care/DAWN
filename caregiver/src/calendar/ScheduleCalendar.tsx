/**
 * ScheduleCalendar.tsx
 *
 * User: Caregiver
 *
 * The calendar grid: FullCalendar showing the carepartner's scheduled videos
 * and audio as day, week or month views. Week is the default, with rows every
 * 15 minutes from 6 am to 10 pm. Hours the carepartner is asleep are hatched.
 *
 * FullCalendar only draws what it is given: the page above passes the
 * server's occurrences for the visible dates (it asks for them when
 * `onRangeChange` says which dates are showing) and moves between dates and
 * views through the `calendarRef`, since the toolbar is the page's own.
 */

import { formatClock, parseClock, parseDate } from "@shared";
import type { AwakeHours } from "@shared";
import type { DatesSetArg, DayHeaderContentArg, EventContentArg, EventInput } from "@fullcalendar/core";
import type { DateClickArg } from "@fullcalendar/interaction";
import dayGridPlugin from "@fullcalendar/daygrid";
import interactionPlugin from "@fullcalendar/interaction";
import FullCalendar from "@fullcalendar/react";
import timeGridPlugin from "@fullcalendar/timegrid";
import { useMemo } from "react";
import type { RefObject } from "react";
import type { Occurrence } from "./api";
import "./calendar.css";
import { dateText, shortTime } from "./time";

export type CalendarViewName = "timeGridDay" | "timeGridWeek" | "dayGridMonth";

export interface VisibleRange {
  /** First and last visible day, "YYYY-MM-DD" */
  from: string;
  to: string;
  view: CalendarViewName;
  /** The range in words, e.g. "Oct 19 – 25, 2026" */
  title: string;
  /** The date the calendar is on, "YYYY-MM-DD": the day in day view, the day last moved to otherwise */
  current: string;
}

interface ScheduleCalendarProps {
  calendarRef: RefObject<FullCalendar | null>;
  occurrences: Occurrence[];
  /** Null until the awake hours have loaded; nothing is hatched meanwhile */
  awakeHours: AwakeHours | null;
  onRangeChange: (range: VisibleRange) => void;
  /** A click on an empty time (week and day) or on a day (month); `date` is that time or day */
  onDateClick: (arg: DateClickArg) => void;
}

// Events are drawn at least this long so a 30 second video is still easy to see and click
const MIN_EVENT_MINUTES = 30;

const ALL_DAYS = [0, 1, 2, 3, 4, 5, 6];

// Null if the server's time can't be read, so one bad item doesn't break the page
function toEvent(occurrence: Occurrence): EventInput | null {
  const startMinute = parseClock(occurrence.time);
  if (startMinute === null) {
    console.warn("Skipping a scheduled item with an unreadable time", occurrence);
    return null;
  }
  const start = parseDate(occurrence.date);
  start.setHours(0, startMinute);
  const end = new Date(start);
  end.setMinutes(end.getMinutes() + Math.max(MIN_EVENT_MINUTES, Math.ceil(occurrence.duration_seconds / 60)));

  const classNames = [occurrence.media.kind === "audio" ? "cal-audio" : "cal-video"];
  if (occurrence.outside_awake_hours) classNames.push("cal-outside");

  return {
    id: `${occurrence.scheduled_item_id}@${occurrence.date}`,
    title: occurrence.media.display_name,
    start,
    end,
    classNames,
    extendedProps: { startMinute },
  };
}

// "8am Title"
function renderEvent({ event }: EventContentArg) {
  return (
    <div className="dawn-event">
      <span className="dawn-event-time">{shortTime(event.extendedProps.startMinute as number)}</span>
      <span className="dawn-event-title">{event.title}</span>
    </div>
  );
}

// Week and day columns show "MON" over the date; month view's header is just the day name.
// Month view passes a placeholder date for each column, so its name has to come from `text`
function renderDayHeader({ date, isToday, text, view }: DayHeaderContentArg) {
  if (view.type === "dayGridMonth") return text.toUpperCase();
  const name = date.toLocaleDateString("en-US", { weekday: "short" }).toUpperCase();
  return (
    <div className="dawn-day-head" data-today={isToday}>
      <span className="dawn-day-name">{name}</span>
      <span className="dawn-day-number">{date.getDate()}</span>
    </div>
  );
}

export default function ScheduleCalendar({ calendarRef, occurrences, awakeHours, onRangeChange, onDateClick }: ScheduleCalendarProps) {
  const events = useMemo(() => occurrences.map(toEvent).filter((event) => event !== null), [occurrences]);

  const businessHours = awakeHours
    ? {
        daysOfWeek: ALL_DAYS,
        startTime: formatClock(awakeHours.wakeMinute) ?? "07:00",
        endTime: formatClock(awakeHours.bedMinute) ?? "21:00",
      }
    : false;

  const handleDatesSet = (arg: DatesSetArg) => {
    // `end` is the day after the last one showing
    const last = new Date(arg.end);
    last.setDate(last.getDate() - 1);
    onRangeChange({
      from: dateText(arg.start),
      to: dateText(last),
      view: arg.view.type as CalendarViewName,
      title: arg.view.title,
      current: dateText(arg.view.calendar.getDate()),
    });
  };

  return (
    <div className="dawn-calendar h-full">
    <FullCalendar
      ref={calendarRef}
      plugins={[dayGridPlugin, timeGridPlugin, interactionPlugin]}
      initialView="timeGridWeek"
      headerToolbar={false}
      height="100%"
      expandRows
      allDaySlot={false}
      // Weeks run Sunday to Saturday, whatever the browser's locale says
      firstDay={0}
      nowIndicator
      navLinks
      // A bare "day" would pick the plain day grid, which has no time rows
      navLinkDayClick="timeGridDay"
      slotMinTime="06:00:00"
      slotMaxTime="22:00:00"
      slotDuration="00:15:00"
      slotLabelInterval="01:00"
      slotLabelFormat={{ hour: "numeric", meridiem: "short" }}
      scrollTime={formatClock(awakeHours?.wakeMinute ?? 7 * 60) + ":00"}
      eventMinHeight={32}
      eventShortHeight={32}
      dayMaxEvents={3}
      moreLinkClick="timeGridDay"
      moreLinkContent={({ num }) => `+${num} more`}
      titleFormat={{ year: "numeric", month: "long", day: "numeric" }}
      views={{
        timeGridDay: { titleFormat: { weekday: "long", year: "numeric", month: "long", day: "numeric" } },
        timeGridWeek: { titleFormat: { year: "numeric", month: "short", day: "numeric" } },
        dayGridMonth: { titleFormat: { year: "numeric", month: "long" } },
      }}
      businessHours={businessHours}
      events={events}
      eventContent={renderEvent}
      dayHeaderFormat={{ weekday: "short" }}
      dayHeaderContent={renderDayHeader}
      dateClick={onDateClick}
      datesSet={handleDatesSet}
    />
    </div>
  );
}
