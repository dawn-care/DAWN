/**
 * PlaybackProvider.tsx
 *
 * Holds the playback state machine (machine.ts) for the whole app and feeds it
 * what it can't know by itself: whether the tablet is paired, and whether it
 * is night. Anything that starts or stops playback goes through `usePlayback`.
 */

import React, { createContext, useCallback, useContext, useEffect, useMemo, useReducer } from "react";
import { now } from "../clock/clock";
import { isNight } from "../display/nightHours";
import { getCarepartnerId } from "../pairing/carepartnerId";
import type { MediaItem } from "./media";
import {
  initialPlaybackState,
  playbackReducer,
  screenFor,
  type PlaybackState,
  type PlaySource,
  type Screen,
} from "./machine";

/** How often to check whether night has started or ended. */
export const NIGHT_CHECK_MS = 60_000;

type PlaybackContextValue = {
  state: PlaybackState;
  screen: Screen;
  /** Asks to play an item; the machine decides whether it starts (see `canPlay`). */
  play: (item: MediaItem, source: PlaySource) => void;
  stop: () => void;
  /** The player for `playId` reported finished or failed. */
  ended: (playId: number) => void;
  /** Reads the stored pairing again, for example after pairing completes. */
  checkPairing: () => Promise<void>;
};

const PlaybackContext = createContext<PlaybackContextValue | null>(null);

export function usePlayback(): PlaybackContextValue {
  const value = useContext(PlaybackContext);
  if (!value) throw new Error("usePlayback must be used within PlaybackProvider");
  return value;
}

export function PlaybackProvider({ children }: { children: React.ReactNode }) {
  const [state, dispatch] = useReducer(playbackReducer, initialPlaybackState);

  // A stored carepartner ID counts as paired for now, although the pair screen
  // stores one before pairing completes (#343).
  const checkPairing = useCallback(async () => {
    const id = await getCarepartnerId();
    dispatch({ type: "pairingChecked", paired: id != null });
  }, []);

  useEffect(() => {
    checkPairing();
  }, [checkPairing]);

  useEffect(() => {
    const update = () => dispatch({ type: "nightChanged", night: isNight(now()) });
    update();
    const timer = setInterval(update, NIGHT_CHECK_MS);
    return () => clearInterval(timer);
  }, []);

  const value = useMemo<PlaybackContextValue>(
    () => ({
      state,
      screen: screenFor(state),
      play: (item, source) => dispatch({ type: "play", item, source }),
      stop: () => dispatch({ type: "stop" }),
      ended: (playId) => dispatch({ type: "ended", playId }),
      checkPairing,
    }),
    [state, checkPairing],
  );

  return <PlaybackContext.Provider value={value}>{children}</PlaybackContext.Provider>;
}
