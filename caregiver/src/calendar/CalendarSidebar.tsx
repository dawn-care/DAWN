/**
 * CalendarSidebar.tsx
 *
 * User: Caregiver
 *
 * The calendar's left column in its default state: the Schedule button, the
 * mini month, the awake-hours card and the legend. (The schedule form takes
 * this column's place while an item is being added or edited.)
 */

import { Plus } from "lucide-react";
import { FOCUS_RING } from "../components/ui/styles";
import type { ScreenSettings } from "./api";
import AwakeHoursCard from "./AwakeHoursCard";
import MiniMonth from "./MiniMonth";
import type { VisibleRange } from "./ScheduleCalendar";

interface CalendarSidebarProps {
  range: VisibleRange | null;
  settings: ScreenSettings | null;
  /** Moves the calendar to a date, "YYYY-MM-DD" */
  onGoToDate: (date: string) => void;
  /** Opens the schedule form; the button is off while this isn't given */
  onSchedule?: () => void;
}

// Same hatch as the asleep hours on the grid (calendar.css)
const ASLEEP_SWATCH = "repeating-linear-gradient(45deg, #E2DCD2 0 3px, #F4F0E9 3px 6px)";

export default function CalendarSidebar({ range, settings, onGoToDate, onSchedule }: CalendarSidebarProps) {
  return (
    <aside className="w-full lg:w-[290px] shrink-0 flex flex-col gap-3.5">
      <button
        type="button"
        onClick={onSchedule}
        disabled={!onSchedule}
        className={`self-start flex items-center gap-2.5 h-[52px] pl-4 pr-5 bg-white border-[1.5px] border-dawn-line rounded-2xl shadow-[0_2px_6px_rgba(30,26,21,0.15)] text-base font-bold text-dawn-ink cursor-pointer disabled:opacity-60 disabled:cursor-default ${FOCUS_RING}`}
      >
        <Plus size={22} strokeWidth={2.6} className="text-dawn-accent" aria-hidden="true" />
        Schedule
      </button>

      <MiniMonth range={range} onPick={onGoToDate} />

      <AwakeHoursCard settings={settings} />

      <ul className="m-0 p-0 pl-1 list-none flex flex-col gap-1.5 text-sm">
        <li className="flex items-center gap-2">
          <span className="w-3.5 h-3.5 rounded-[3px] bg-dawn-accent" aria-hidden="true" />
          Video, plays full screen
        </li>
        <li className="flex items-center gap-2">
          <span className="w-3.5 h-3.5 rounded-[3px] bg-dawn-audio" aria-hidden="true" />
          Audio
        </li>
        <li className="flex items-center gap-2">
          <span
            className="w-3.5 h-3.5 rounded-[3px] border border-dawn-line-strong box-border"
            style={{ background: ASLEEP_SWATCH }}
            aria-hidden="true"
          />
          Asleep, nothing plays
        </li>
      </ul>
    </aside>
  );
}
