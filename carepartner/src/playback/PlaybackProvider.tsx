/**
 * PlaybackProvider.tsx
 *
 * Holds the playback state machine (machine.ts) for the whole app and feeds it
 * what it can't know by itself: whether the tablet is paired, and whether it
 * is night. Anything that starts or stops playback goes through `usePlayback`.
 */

import React, { createContext, useCallback, useContext, useEffect, useMemo, useReducer } from "react";
import { now } from "../clock/clock";
import { isNight } from "@shared";
import { getCarepartnerId } from "../pairing/carepartnerId";
import type { MediaItem } from "./media";
import {
  initialPlaybackState,
  playbackReducer,
  screenFor,
  type PlaybackState,
  type PlayerOutcome,
  type PlaySource,
  type Screen,
} from "./machine";

/** How often to check whether night has started or ended. */
export const NIGHT_CHECK_MS = 60_000;

type PlaybackContextValue = {
  state: PlaybackState;
  screen: Screen;
  /**
   * Asks to play an item; the machine decides whether it starts (see `canPlay`).
   * `playbackId` is the server's Playback record for it, if it has one.
   */
  play: (item: MediaItem, source: PlaySource, playbackId?: number) => void;
  /** Stops the current play, or with `playbackId`, only the play with that record. */
  stop: (playbackId?: number) => void;
  /** The player for `playId` reported finished or failed. */
  ended: (playId: number, reason: PlayerOutcome) => void;
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

  // Paired means a carepartner ID is stored; none stored (null) means unpaired.
  // The pair screen stores the ID of a newly created carepartner before a
  // caregiver has claimed it, so that also counts as paired for now (#343).
  // The machine keeps the ID, so nothing else reads it from storage.
  const checkPairing = useCallback(async () => {
    dispatch({ type: "pairingChecked", carepartnerId: await getCarepartnerId() });
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
      play: (item, source, playbackId) => dispatch({ type: "play", item, source, playbackId }),
      stop: (playbackId) => dispatch({ type: "stop", playbackId }),
      ended: (playId, reason) => dispatch({ type: "ended", playId, reason }),
      checkPairing,
    }),
    [state, checkPairing],
  );

  return <PlaybackContext.Provider value={value}>{children}</PlaybackContext.Provider>;
}
