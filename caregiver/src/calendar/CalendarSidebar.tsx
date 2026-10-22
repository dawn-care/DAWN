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
  carepartnerId: number;
  range: VisibleRange | null;
  settings: ScreenSettings | null;
  /** Whether the awake-hours editor is open in place of the hours */
  editingAwake: boolean;
  onEditAwake: () => void;
  onCloseAwake: () => void;
  onAwakeSaved: (saved: ScreenSettings) => void;
  /** Moves the calendar to a date, "YYYY-MM-DD" */
  onGoToDate: (date: string) => void;
  /** Opens the schedule form; the button is off while this isn't given */
  onSchedule?: () => void;
}

// A smaller, slightly darker version of the asleep hatch in calendar.css, so it still reads at legend size
const ASLEEP_SWATCH = "repeating-linear-gradient(45deg, rgb(var(--color-border)) 0 3px, rgb(var(--color-page)) 3px 6px)";

export default function CalendarSidebar({
  carepartnerId,
  range,
  settings,
  editingAwake,
  onEditAwake,
  onCloseAwake,
  onAwakeSaved,
  onGoToDate,
  onSchedule,
}: CalendarSidebarProps) {
  return (
    <aside className="w-full lg:w-[290px] shrink-0 flex flex-col gap-3.5">
      <button
        type="button"
        onClick={onSchedule}
        disabled={!onSchedule}
        className={`self-start flex items-center gap-2.5 h-[52px] pl-4 pr-5 bg-surface border-[1.5px] border-border rounded-2xl shadow-raised text-base font-bold text-ink cursor-pointer disabled:opacity-60 disabled:cursor-default ${FOCUS_RING}`}
      >
        <Plus size={22} strokeWidth={2.6} className="text-link" aria-hidden="true" />
        Schedule
      </button>

      <MiniMonth range={range} onPick={onGoToDate} />

      <AwakeHoursCard
        carepartnerId={carepartnerId}
        settings={settings}
        editing={editingAwake}
        onEdit={onEditAwake}
        onClose={onCloseAwake}
        onSaved={onAwakeSaved}
      />

      <ul className="m-0 p-0 pl-1 list-none flex flex-col gap-1.5 text-sm">
        <li className="flex items-center gap-2">
          <span className="w-3.5 h-3.5 rounded-[3px] bg-accent" aria-hidden="true" />
          Video, plays full screen
        </li>
        <li className="flex items-center gap-2">
          <span className="w-3.5 h-3.5 rounded-[3px] bg-ink" aria-hidden="true" />
          Audio
        </li>
        <li className="flex items-center gap-2">
          <span
            className="w-3.5 h-3.5 rounded-[3px] border border-border box-border"
            style={{ background: ASLEEP_SWATCH }}
            aria-hidden="true"
          />
          Asleep, nothing plays
        </li>
      </ul>
    </aside>
  );
}
