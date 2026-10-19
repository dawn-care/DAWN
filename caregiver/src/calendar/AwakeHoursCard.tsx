/**
 * AwakeHoursCard.tsx
 *
 * User: Caregiver
 *
 * The "Awake hours" card in the calendar's left column: when the carepartner
 * wakes up and goes to bed, from the screen settings. Videos and audio only
 * play between them.
 */

import { timeLabel } from "@shared";
import type { ScreenSettings } from "./api";

interface AwakeHoursCardProps {
  /** Null until the settings have loaded (or if they couldn't be) */
  settings: ScreenSettings | null;
}

export default function AwakeHoursCard({ settings }: AwakeHoursCardProps) {
  return (
    <section aria-label="Awake hours" className="flex flex-col gap-1.5 p-3 bg-white border-[1.5px] border-dawn-line rounded-xl">
      <h2 className="m-0 text-[15px] font-extrabold">Awake hours</h2>
      <p className="m-0 text-[17px] font-bold">
        {settings ? `${timeLabel(settings.wake_time)} – ${timeLabel(settings.bed_time)}` : "—"}
      </p>
      <p className="m-0 text-sm leading-snug text-dawn-muted">Videos and audio only play while the carepartner is awake.</p>
    </section>
  );
}
