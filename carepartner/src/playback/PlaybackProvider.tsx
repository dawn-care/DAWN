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
  type PlayOptions,
  type PlaySource,
  type Screen,
} from "./machine";

/** How long to wait before reading pairing again after storage fails at startup. */
export const PAIRING_RETRY_MS = 5_000;

type PlaybackContextValue = {
  state: PlaybackState;
  screen: Screen;
  /**
   * Asks to play an item; the machine decides whether it starts (see `canPlay`).
   * `playbackId` is the server's Playback record for it, if it has one, and
   * `scheduledItemId` the scheduled item a scheduled play comes from.
   */
  play: (item: MediaItem, source: PlaySource, options?: PlayOptions) => void;
  /** Stops the current play, or with `playbackId`, only the play with that record. */
  stop: (playbackId?: number) => void;
  /** The chime before `playId` and its pause are over. */
  chimed: (playId: number) => void;
  /** The player for `playId` says its media is actually playing. */
  started: (playId: number) => void;
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

  // Reading storage can fail on Android; until it works the screen stays
  // blank, so keep trying rather than give up or guess.
  useEffect(() => {
    let timer: ReturnType<typeof setTimeout> | undefined;
    let cancelled = false;
    const attempt = () =>
      checkPairing().catch((error) => {
        console.warn("Couldn't read whether this tablet is paired, trying again", error);
        if (!cancelled) timer = setTimeout(attempt, PAIRING_RETRY_MS);
      });
    attempt();
    return () => {
      cancelled = true;
      clearTimeout(timer);
    };
  }, [checkPairing]);

  const setNight = useCallback((night: boolean) => dispatch({ type: "nightChanged", night }), []);

  const value = useMemo<PlaybackContextValue>(
    () => ({
      state,
      screen: screenFor(state),
      play: (item, source, options) => dispatch({ type: "play", item, source, ...options }),
      stop: (playbackId) => dispatch({ type: "stop", playbackId }),
      chimed: (playId) => dispatch({ type: "chimed", playId }),
      started: (playId) => dispatch({ type: "started", playId }),
      ended: (playId, reason) => dispatch({ type: "ended", playId, reason }),
      checkPairing,
      setNight,
    }),
    [state, checkPairing, setNight],
  );

  return <PlaybackContext.Provider value={value}>{children}</PlaybackContext.Provider>;
}
