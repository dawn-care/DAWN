import { useEffect, useState } from "react";
import { everyMinute, now } from "./clock";

/**
 * The current time, re-read as each minute starts (or every `intervalMs`, for
 * a clock with seconds) so a component redraws as time passes. Reads the time
 * through the app clock.
 */
export function useNow(intervalMs?: number): Date {
  const [current, setCurrent] = useState(now);

  useEffect(() => {
    const update = () => setCurrent(now());
    if (intervalMs === undefined) return everyMinute(update);
    const timer = setInterval(update, intervalMs);
    return () => clearInterval(timer);
  }, [intervalMs]);

  return current;
}
