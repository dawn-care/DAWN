/**
 * MiniMonth.tsx
 *
 * User: Caregiver
 *
 * The small month in the calendar's left column. It shows the month of the
 * date the calendar is on, marks today, shades the days the calendar is
 * showing (the week, in week view) and darkens the day in day view. Clicking
 * a date moves the calendar there, keeping its view. The arrows beside the
 * month name browse other months without moving the calendar; the month goes
 * back to the calendar's own whenever the calendar moves to another date
 * (switching views on the same date keeps the month browsed to).
 */

import { ChevronLeft, ChevronRight } from "lucide-react";
import { useState } from "react";

import type { VisibleRange } from "./ScheduleCalendar";
import { dateText } from "./time";

interface MiniMonthProps {
  range: VisibleRange | null;
  /** Called with the clicked date as "YYYY-MM-DD" */
  onPick: (date: string) => void;
}

const WEEKDAY_LETTERS = ["S", "M", "T", "W", "T", "F", "S"];

const ARROW = "w-8 h-8 flex items-center justify-center rounded-full border-0 bg-transparent text-ink cursor-pointer hover:bg-border/60 focus-visible:outline focus-visible:outline-2 focus-visible:outline-focus";

const DAY_BASE = "h-7 rounded-full border-0 text-[13px] cursor-pointer focus-visible:outline focus-visible:outline-2 focus-visible:outline-focus";

export default function MiniMonth({ range, onPick }: MiniMonthProps) {
  // Before the calendar has drawn, show the month around today
  const today = dateText(new Date());
  const calendarDate = range?.current ?? today;
  const current = new Date(`${calendarDate}T00:00:00`);

  // A month the arrows moved to, which only applies while the calendar stays on the date it was moved from
  const [browsed, setBrowsed] = useState<{ from: string; month: Date } | null>(null);
  const firstOfMonth =
    browsed && browsed.from === calendarDate ? browsed.month : new Date(current.getFullYear(), current.getMonth(), 1);
  const browse = (step: number) =>
    setBrowsed({ from: calendarDate, month: new Date(firstOfMonth.getFullYear(), firstOfMonth.getMonth() + step, 1) });

  const daysInMonth = new Date(firstOfMonth.getFullYear(), firstOfMonth.getMonth() + 1, 0).getDate();
  const cellCount = Math.ceil((firstOfMonth.getDay() + daysInMonth) / 7) * 7;
  const days = Array.from({ length: cellCount }, (_, i) => {
    const day = new Date(firstOfMonth);
    day.setDate(1 - firstOfMonth.getDay() + i);
    return day;
  });

  const title = firstOfMonth.toLocaleDateString("en-US", { month: "long", year: "numeric" });

  const styleOf = (day: Date, text: string) => {
    if (text === today) return "bg-accent text-on-accent font-bold";
    if (range?.view === "timeGridDay" && text === range.from) return "bg-ink text-page";
    if (range?.view === "timeGridWeek" && text >= range.from && text <= range.to) return "bg-border text-ink";
    if (day.getMonth() === firstOfMonth.getMonth()) return "bg-transparent text-ink";
    return "bg-transparent text-ink-muted";
  };

  return (
    <div className="flex flex-col gap-1">
      <div className="flex items-center justify-between">
        <div className="pl-1 text-[15px] font-bold">{title}</div>
        <div className="flex">
          <button type="button" onClick={() => browse(-1)} aria-label="Previous month" className={ARROW}>
            <ChevronLeft size={18} strokeWidth={2.4} aria-hidden="true" />
          </button>
          <button type="button" onClick={() => browse(1)} aria-label="Next month" className={ARROW}>
            <ChevronRight size={18} strokeWidth={2.4} aria-hidden="true" />
          </button>
        </div>
      </div>
      <div className="grid grid-cols-7 gap-px text-center">
        {WEEKDAY_LETTERS.map((letter, i) => (
          <span key={i} aria-hidden="true" className="h-[22px] leading-[22px] text-xs font-bold text-ink-muted">
            {letter}
          </span>
        ))}
        {days.map((day) => {
          const text = dateText(day);
          return (
            <button
              key={text}
              type="button"
              onClick={() => onPick(text)}
              aria-label={day.toLocaleDateString("en-US", { month: "long", day: "numeric" })}
              className={`${DAY_BASE} ${styleOf(day, text)}`}
            >
              {day.getDate()}
            </button>
          );
        })}
      </div>
    </div>
  );
}
