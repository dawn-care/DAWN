/**
 * PlaybackProvider.tsx
 *
 * Holds the playback state machine (machine.ts) for the whole app and feeds it
 * what it can't know by itself: whether the tablet is paired, and (from
 * `TabletDisplay`) whether it is night. Anything that starts or stops playback
 * goes through `usePlayback`.
 */

import React, { createContext, useCallback, useContext, useEffect, useMemo, useReducer } from "react";
import { readPairedId } from "../pairing/pairingState";
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
  /** Whether it is night now, by the carepartner's night hours. */
  setNight: (night: boolean) => void;
};

const PlaybackContext = createContext<PlaybackContextValue | null>(null);

export function usePlayback(): PlaybackContextValue {
  const value = useContext(PlaybackContext);
  if (!value) throw new Error("usePlayback must be used within PlaybackProvider");
  return value;
}

export function PlaybackProvider({ children }: { children: React.ReactNode }) {
  const [state, dispatch] = useReducer(playbackReducer, initialPlaybackState);

  // Paired means a caregiver has claimed the stored carepartner (see
  // `readPairedId`). The machine keeps the ID, so nothing else reads it from
  // storage.
  const checkPairing = useCallback(async () => {
    dispatch({ type: "pairingChecked", carepartnerId: await readPairedId() });
  }, []);

  useEffect(() => {
    checkPairing();
  }, [checkPairing]);

  const setNight = useCallback((night: boolean) => dispatch({ type: "nightChanged", night }), []);

  const value = useMemo<PlaybackContextValue>(
    () => ({
      state,
      screen: screenFor(state),
      play: (item, source, playbackId) => dispatch({ type: "play", item, source, playbackId }),
      stop: (playbackId) => dispatch({ type: "stop", playbackId }),
      ended: (playId, reason) => dispatch({ type: "ended", playId, reason }),
      checkPairing,
      setNight,
    }),
    [state, checkPairing, setNight],
  );

  return <PlaybackContext.Provider value={value}>{children}</PlaybackContext.Provider>;
}
