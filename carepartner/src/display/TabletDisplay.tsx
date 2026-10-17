/**
 * TabletDisplay.tsx
 *
 * Decides when it is night, from the carepartner's awake hours on the screen
 * (night runs from bed time to wake time), or the default night hours until
 * the first screen arrives. It tells the playback machine, which shows the
 * night screen, and dims the tablet. Draws nothing.
 */

import { useEffect, useMemo } from "react";
import { DEFAULT_NIGHT_HOURS, isNight, nightHoursFromAwake } from "@shared";
import { now } from "../clock/clock";
import { usePlayback } from "../playback/PlaybackProvider";
import { useScreen } from "../screen/ScreenProvider";
import { useTabletDisplay } from "./useTabletDisplay";

/** How often to check whether night has started or ended. */
export const NIGHT_CHECK_MS = 60_000;

export default function TabletDisplay() {
  const awake = useScreen().screen?.awakeHours;
  const { setNight } = usePlayback();
  const wake = awake?.wake;
  const bed = awake?.bed;
  const hours = useMemo(
    () => (wake !== undefined && bed !== undefined ? nightHoursFromAwake({ wake, bed }) : DEFAULT_NIGHT_HOURS),
    [wake, bed],
  );

  useTabletDisplay(hours);

  useEffect(() => {
    const update = () => setNight(isNight(now(), hours));
    update();
    const timer = setInterval(update, NIGHT_CHECK_MS);
    return () => clearInterval(timer);
  }, [hours, setNight]);

  return null;
}
