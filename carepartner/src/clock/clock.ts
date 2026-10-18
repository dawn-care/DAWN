/**
 * clock.ts
 *
 * The tablet's only source of the current time. Read the time through `now()`
 * so tests and manual checks can control it.
 *
 * Set EXPO_PUBLIC_FAKE_TIME (an ISO date-time, for example
 * "2026-10-03T21:30:00") in .env.local to start the app's clock at that time.
 * It then runs forward normally, so night mode or an 8:00 schedule can be
 * checked without waiting for the real time.
 */

type TimeSource = () => number;

const realTime: TimeSource = () => Date.now();

/** A clock that starts at `start` now and runs forward in real time. */
function startingAt(start: Date): TimeSource {
  const offset = start.getTime() - Date.now();
  return () => Date.now() + offset;
}

/** Reads a fake start time setting; null when it is unset or not a valid date. */
export function parseFakeTime(value: string | undefined): Date | null {
  if (!value) return null;
  const start = new Date(value);
  if (Number.isNaN(start.getTime())) {
    console.warn(`Ignoring EXPO_PUBLIC_FAKE_TIME: "${value}" is not a valid date`);
    return null;
  }
  return start;
}

const fakeStart = parseFakeTime(process.env.EXPO_PUBLIC_FAKE_TIME);
let source: TimeSource = fakeStart ? startingAt(fakeStart) : realTime;

/** The EXPO_PUBLIC_FAKE_TIME the app started at, or null for the real time. For developer tools. */
export function fakeStartTime(): Date | null {
  return fakeStart;
}

export function now(): Date {
  return new Date(source());
}

/**
 * Starts the clock at the given time, running forward from there, or returns
 * it to the real time when given null. For tests and developer tools only.
 */
export function setClock(start: Date | null): void {
  source = start ? startingAt(start) : realTime;
}

const MINUTE_MS = 60_000;

/**
 * Calls `tick` as each minute of the app clock starts, so checks of the time
 * (night, the schedule, the time of day) happen on the minute rather than up
 * to a minute late. Returns a function that stops it.
 */
export function everyMinute(tick: () => void): () => void {
  let timer: ReturnType<typeof setTimeout>;
  const wait = () => {
    // Worked out again each minute, so it never drifts. A timer that fires a
    // little early just ticks twice. A tick that throws doesn't stop the next.
    timer = setTimeout(() => {
      try {
        tick();
      } finally {
        wait();
      }
    }, MINUTE_MS - (source() % MINUTE_MS));
  };
  wait();
  return () => clearTimeout(timer);
}

/**
 * A deadline `ms` from now in real time, for the server to check against its
 * own clock (such as a play report's `expires_at`). The one place the app
 * deliberately ignores `EXPO_PUBLIC_FAKE_TIME`: a deadline from a fake clock
 * hours off would be refused or expire at once.
 */
export function serverDeadline(ms: number): Date {
  return new Date(Date.now() + ms);
}

/** "YYYY-MM-DD" for the local day of `time`. */
export function dayOf(time: Date): string {
  const month = String(time.getMonth() + 1).padStart(2, "0");
  const day = String(time.getDate()).padStart(2, "0");
  return `${time.getFullYear()}-${month}-${day}`;
}
