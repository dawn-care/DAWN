import { useEffect, useState } from "react";
import { now } from "./clock";

/**
 * The current time, re-read every `intervalMs` so a component redraws as time
 * passes. Reads the time through the app clock.
 */
export function useNow(intervalMs = 30_000): Date {
  const [current, setCurrent] = useState(now);

  useEffect(() => {
    const timer = setInterval(() => setCurrent(now()), intervalMs);
    return () => clearInterval(timer);
  }, [intervalMs]);

  return current;
}
