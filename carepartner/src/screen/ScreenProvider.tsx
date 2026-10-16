/**
 * ScreenProvider.tsx
 *
 * Keeps the always-on screen from the server current: everything the tablet
 * shows, including the schedule.
 *
 * - Once paired, fetches `/screen`, then again every 5 minutes, on the first
 *   check of a new day, and whenever `refresh()` is called (the socket's
 *   refresh message and the Sync button).
 * - Keeps the last good screen if a fetch fails.
 * - Live data (weather, sports, news) hides once the last good fetch is hours
 *   old (see `hideStaleWidgets`).
 */

import React, { createContext, useCallback, useContext, useEffect, useMemo, useRef, useState } from "react";
import { API_URL } from "../api/config";
import { dayOf, now } from "../clock/clock";
import { usePlayback } from "../playback/PlaybackProvider";
import { hideStaleWidgets, readScreen, type Screen } from "./screen";

export const SCREEN_REFRESH_MS = 5 * 60_000;
/** How often to check whether it is time to fetch again, and whether data has gone stale. */
export const SCREEN_CHECK_MS = 60_000;

type ScreenContextValue = {
  /** The last good screen, or null before the first one arrives. */
  screen: Screen | null;
  /** Fetches the screen again now. */
  refresh: () => Promise<void>;
};

const ScreenContext = createContext<ScreenContextValue | null>(null);

export function useScreen(): ScreenContextValue {
  const value = useContext(ScreenContext);
  if (!value) throw new Error("useScreen must be used within ScreenProvider");
  return value;
}

export function ScreenProvider({ children }: { children: React.ReactNode }) {
  const id = usePlayback().state.carepartnerId;
  const [fetched, setFetched] = useState<{ screen: Screen; at: Date } | null>(null);
  /** When the last fetch was tried, successful or not, so a failing server is tried every 5 minutes, not every check. */
  const lastTried = useRef<Date | null>(null);
  /** Re-reads the clock at each check, so stale data hides even when no fetch succeeds. */
  const [checkedAt, setCheckedAt] = useState(() => now());

  const refresh = useCallback(async () => {
    if (id == null) return;
    lastTried.current = now();
    try {
      const res = await fetch(`${API_URL}/screen?carepartner_id=${id}`);
      if (!res.ok) throw new Error(`Screen request failed (${res.status})`);
      const screen = readScreen(await res.json());
      if (!screen) throw new Error("Screen response is malformed");
      setFetched({ screen, at: now() });
    } catch (error) {
      // Keep the last good screen; the next check tries again.
      console.warn("Couldn't refresh the screen", error);
    }
  }, [id]);

  useEffect(() => {
    setFetched(null);
    if (id == null) return;
    refresh();
    const timer = setInterval(() => {
      const time = now();
      const tried = lastTried.current;
      if (!tried || time.getTime() - tried.getTime() >= SCREEN_REFRESH_MS || dayOf(time) !== dayOf(tried)) refresh();
      setCheckedAt(time);
    }, SCREEN_CHECK_MS);
    return () => clearInterval(timer);
  }, [id, refresh]);

  const screen = useMemo(() => fetched && hideStaleWidgets(fetched.screen, fetched.at, checkedAt), [fetched, checkedAt]);
  const value = useMemo(() => ({ screen, refresh }), [screen, refresh]);
  return <ScreenContext.Provider value={value}>{children}</ScreenContext.Provider>;
}
