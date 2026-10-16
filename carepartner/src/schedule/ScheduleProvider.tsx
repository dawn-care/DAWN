/**
 * ScheduleProvider.tsx
 *
 * Fetches today's schedule and asks the playback machine to play each entry
 * when it is due, one at a time. It never navigates: the screen follows the
 * machine.
 *
 * - Once paired, fetches the carepartner (for `auto_start`) and today's
 *   entries, and again every 5 minutes.
 * - Every 15 seconds, marks timed entries due at their minute, and starts the
 *   routine at the auto-start minute. On a new day it forgets what played and
 *   fetches the new schedule.
 * - Whenever nothing is playing, plays the next due entry (see `nextEntry`).
 *   An entry counts as played once it starts, so a failed one doesn't loop.
 */

import React, { createContext, useCallback, useContext, useEffect, useMemo, useState } from "react";
import { API_URL } from "../api/config";
import { now } from "../clock/clock";
import { canPlay } from "../playback/machine";
import { usePlayback } from "../playback/PlaybackProvider";
import {
  dayOf,
  minuteOf,
  nextEntry,
  readEntries,
  routineStartsAt,
  timedEntriesDueAt,
  type ScheduleEntry,
} from "./schedule";

export const SCHEDULE_REFRESH_MS = 5 * 60_000;
export const SCHEDULE_CHECK_MS = 15_000;

/** The carepartner record from the server. */
export type Carepartner = {
  id: number;
  name: string;
  /** "HH:MM:SS" */
  auto_start: string;
};

type ScheduleContextValue = {
  carepartner: Carepartner | null;
  /** Fetches the carepartner and today's schedule again. */
  refresh: () => Promise<void>;
};

const ScheduleContext = createContext<ScheduleContextValue | null>(null);

export function useSchedule(): ScheduleContextValue {
  const value = useContext(ScheduleContext);
  if (!value) throw new Error("useSchedule must be used within ScheduleProvider");
  return value;
}

export function ScheduleProvider({ children }: { children: React.ReactNode }) {
  const playback = usePlayback();
  const [carepartner, setCarepartner] = useState<Carepartner | null>(null);
  const [entries, setEntries] = useState<ScheduleEntry[]>([]);
  const [played, setPlayed] = useState<Set<number>>(new Set());
  /** Timed entries whose minute has come, waiting for nothing else to be playing. */
  const [due, setDue] = useState<Set<number>>(new Set());
  /** The day the routine was started, if it was. */
  const [routineDay, setRoutineDay] = useState<string | null>(null);
  /** The day `played` and `due` belong to. */
  const [day, setDay] = useState(() => dayOf(now()));

  // An unpaired tablet has no schedule, so fetching starts once it is paired.
  const id = playback.state.carepartnerId;
  const refresh = useCallback(async () => {
    if (id == null) return;
    try {
      const carepartnerRes = await fetch(`${API_URL}/users/carepartner/${id}`);
      if (carepartnerRes.ok) setCarepartner(await carepartnerRes.json());

      const today = now();
      const params = new URLSearchParams({
        carepartner_id: String(id),
        day: String(today.getDate()),
        month: String(today.getMonth() + 1),
        day_of_week: String(today.getDay()),
        date: dayOf(today),
      });
      const entriesRes = await fetch(`${API_URL}/configurations/day?${params}`);
      if (!entriesRes.ok) throw new Error(`Schedule request failed (${entriesRes.status})`);
      setEntries(readEntries(await entriesRes.json()));
    } catch (error) {
      // Keep the last schedule; the next refresh tries again.
      console.warn("Couldn't refresh the schedule", error);
    }
  }, [id]);

  useEffect(() => {
    if (id == null) return;
    refresh();
    const timer = setInterval(refresh, SCHEDULE_REFRESH_MS);
    return () => clearInterval(timer);
  }, [id, refresh]);

  useEffect(() => {
    const check = () => {
      const time = now();
      if (dayOf(time) !== day) {
        // A new day: forget what played and fetch the new day's schedule.
        setDay(dayOf(time));
        setPlayed(new Set());
        setDue(new Set());
        setRoutineDay(null);
        refresh();
        return;
      }
      const minute = minuteOf(time);
      const newlyDue = timedEntriesDueAt(entries, minute, played);
      setDue((current) => {
        const added = newlyDue.filter((id) => !current.has(id));
        return added.length ? new Set([...current, ...added]) : current;
      });
      if (routineStartsAt(carepartner?.auto_start, minute, entries, played)) setRoutineDay(dayOf(time));
    };
    check();
    const timer = setInterval(check, SCHEDULE_CHECK_MS);
    return () => clearInterval(timer);
  }, [entries, played, carepartner, day, refresh]);

  const { state, play } = playback;
  useEffect(() => {
    if (!canPlay(state, "scheduled")) return;
    const entry = nextEntry(entries, { played, due, routineRunning: routineDay === day });
    if (!entry) return;
    setPlayed((current) => new Set(current).add(entry.id));
    play(entry, "scheduled");
  }, [state, play, entries, played, due, routineDay, day]);

  const value = useMemo(() => ({ carepartner, refresh }), [carepartner, refresh]);
  return <ScheduleContext.Provider value={value}>{children}</ScheduleContext.Provider>;
}
