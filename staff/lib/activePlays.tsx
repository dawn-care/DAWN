/**
 * The plays started from this phone, kept for the whole app rather than one
 * screen, so staff can leave the Playing screen and still stop or rate them
 * from the mini player (components/MiniPlayer).
 *
 * One timer asks the server about every play that still has something to
 * learn (getPlayState). What each play shows is worked out from that answer and
 * the clock (phaseOf), so the Playing screen and the mini player always agree.
 *
 * Only this phone's plays are here: scheduled plays and other staff members'
 * plays aren't. The list lasts until the app closes; plays only matter for a few
 * minutes (the rating window closes five minutes after the end).
 */

import { createContext, useCallback, useContext, useEffect, useMemo, useRef, useState, type ReactNode } from 'react';

import { getClip, getPlayState, getResident, playOnTablet, rateClip, stopOnTablet, type PlayState } from '../data/server';
import type { Clip, Rating, Resident } from '../data/types';
import { logEvent } from './log';

/** How long the server waits for an offline tablet before the command expires. */
export const DELIVERY_WINDOW_MS = 30 * 1000;
/** How often to ask the server what the tablet has done with a play. */
const POLL_MS = 3 * 1000;
/** How long a clip is on screen before staff can rate it, if it hasn't ended by then. */
const RATE_AFTER_MS = 15 * 1000;
/** How many plays to remember; older ones are long over. */
const KEEP = 20;

/**
 * - playing: sent to the tablet, estimated to be playing
 * - waiting: the tablet wasn't connected; it may still start if it reconnects
 * - ended: finished, as the tablet reported or as estimated
 * - stopped: staff stopped it
 * - missed: it never reached the tablet's screen (not delivered in time, or cancelled first)
 * - failed: the tablet showed it but couldn't play it to the end
 */
export type Phase = 'playing' | 'waiting' | 'ended' | 'stopped' | 'missed' | 'failed';

export interface ActivePlay {
  playbackId: number;
  resident: Resident;
  clip: Clip;
  /** When this phone sent it (ms since epoch). */
  sentAt: number;
  /** Whether the tablet was connected to receive it right away. */
  delivered: boolean;
  /** The server's latest answer; null until the first one. */
  state: PlayState | null;
  /** How this phone's Stop ended it, before the server says so. */
  localEnd: { phase: 'ended' | 'stopped' | 'missed'; at: number } | null;
  stopping: boolean;
  stopNote: string | null;
  /** The rating being saved. */
  ratingPending: Rating | null;
  ratingNote: string | null;
  /** The rating the server saved. */
  rated: Rating | null;
  /** The server stopped accepting ratings. */
  ratingClosed: boolean;
  /** A newer play from this phone took its place on the same tablet. */
  replaced: boolean;
  /** Staff closed it in the mini player. */
  dismissed: boolean;
}

/** What the play is doing at `now`. The tablet's own reports beat the estimate. */
export function phaseOf(p: ActivePlay, now: number): Phase {
  const s = p.state;
  const started = s?.startedAt != null;
  // Received but never shown (the server marks that failed) is the same as not received
  if (s?.failed) return started ? 'failed' : 'missed';
  if (s?.stopped) return started ? 'stopped' : 'missed';
  if (p.localEnd) return p.localEnd.phase;
  if (s?.ended) return started ? 'ended' : 'missed';
  if (started && s?.startedAt != null) {
    // Estimated from the clip's length, since the tablet only reports the end
    const duration = p.clip.duration;
    return duration !== null && now - s.startedAt >= duration * 1000 ? 'ended' : 'playing';
  }
  if (p.delivered) return 'playing';
  return now - p.sentAt >= DELIVERY_WINDOW_MS ? 'missed' : 'waiting';
}

/** Seconds into the clip: the time since it reached the screen, up to its length. */
export function positionOf(p: ActivePlay, now: number): number {
  const duration = p.clip.duration;
  const startedAt = p.state?.startedAt;
  if (!duration || startedAt == null) return 0;
  if (phaseOf(p, now) === 'ended') return duration;
  const end = p.state?.endedAt ?? p.localEnd?.at ?? now;
  return Math.min(duration, Math.max(0, Math.floor((end - startedAt) / 1000)));
}

/**
 * Whether "How did it go?" can be offered: it reached the screen (the server
 * gave it a closing time), the window is still open, and it has been on for
 * RATE_AFTER_MS or is already over. A play the tablet couldn't finish isn't
 * rated, since that says something about the tablet, not the clip; nor is one
 * staff replaced with another.
 */
export function canRate(p: ActivePlay, now: number): boolean {
  const closesAt = p.state?.ratingClosesAt ?? null;
  const startedAt = p.state?.startedAt ?? null;
  if (closesAt === null || startedAt === null || now >= closesAt || p.ratingClosed || p.replaced) return false;
  const phase = phaseOf(p, now);
  if (phase === 'failed' || phase === 'missed') return false;
  return now - startedAt >= RATE_AFTER_MS || phase === 'ended' || phase === 'stopped';
}

/** When it stopped being active, for how long a missed or failed play stays in view. */
export function endedAt(p: ActivePlay): number {
  return p.state?.endedAt ?? p.localEnd?.at ?? p.sentAt + DELIVERY_WINDOW_MS;
}

/** Nothing more to learn from the server: no need to keep asking. */
function settled(p: ActivePlay, now: number): boolean {
  const phase = phaseOf(p, now);
  if (phase === 'missed' || phase === 'failed') return true;
  const closesAt = p.state?.ratingClosesAt ?? null;
  if (p.ratingClosed || (closesAt !== null && now >= closesAt)) return true;
  // Over on the tablet, and no rating left to take
  const over = p.state?.ended || p.localEnd !== null;
  return Boolean(over && (p.rated !== null || p.replaced));
}

interface ActivePlays {
  /** Newest first. */
  plays: ActivePlay[];
  /** Sends a clip to the resident's tablet and starts following it; returns the playback ID. */
  start(resident: Resident, clip: Clip): Promise<number>;
  stop(playbackId: number): Promise<void>;
  rate(playbackId: number, rating: Rating): Promise<void>;
  /** Brings the rating buttons back so a mis-tap can be fixed while the window is open. */
  changeRating(playbackId: number): void;
  dismiss(playbackId: number): void;
  /** Follows a play this phone no longer remembers (after a web reload, say). */
  adopt(residentId: number, playbackId: number, clipId: number): Promise<void>;
}

const Context = createContext<ActivePlays | null>(null);

export function ActivePlaysProvider({ children }: { children: ReactNode }) {
  const [plays, setPlays] = useState<ActivePlay[]>([]);
  // The latest list, for callbacks that shouldn't change every time it does
  const latest = useRef(plays);
  useEffect(() => {
    latest.current = plays;
  }, [plays]);

  const update = useCallback((playbackId: number, change: (p: ActivePlay) => Partial<ActivePlay>) => {
    setPlays((all) => all.map((p) => (p.playbackId === playbackId ? { ...p, ...change(p) } : p)));
  }, []);

  const add = useCallback((play: ActivePlay) => {
    setPlays((all) => [
      play,
      // The tablet plays one thing at a time, so a new play replaces this phone's earlier one there
      ...all
        .filter((p) => p.playbackId !== play.playbackId)
        .map((p) => (p.resident.id === play.resident.id && !p.replaced ? { ...p, replaced: true } : p)),
    ].slice(0, KEEP));
  }, []);

  const start = useCallback(
    async (resident: Resident, clip: Clip) => {
      logEvent({ type: 'play_requested', residentId: resident.id, clipId: clip.id });
      const { playback, delivered } = await playOnTablet(resident.id, clip.id);
      add(newPlay(playback.id, resident, clip, delivered));
      return playback.id;
    },
    [add]
  );

  const adopt = useCallback(
    async (residentId: number, playbackId: number, clipId: number) => {
      if (latest.current.some((p) => p.playbackId === playbackId)) return;
      const [resident, clip, state] = await Promise.all([
        getResident(residentId),
        getClip(clipId),
        getPlayState(residentId, playbackId),
      ]);
      if (!resident || !clip) throw new Error('play not found');
      // When it was sent isn't known any more; treat it as delivered so it isn't called missed early
      const play = newPlay(playbackId, resident, clip, true);
      setPlays((all) =>
        all.some((p) => p.playbackId === playbackId) ? all : [{ ...play, state }, ...all].slice(0, KEEP)
      );
    },
    []
  );

  const stop = useCallback(
    async (playbackId: number) => {
      const play = latest.current.find((p) => p.playbackId === playbackId);
      if (!play || play.stopping) return;
      const before = phaseOf(play, Date.now());
      update(playbackId, () => ({ stopping: true, stopNote: null }));
      try {
        const result = await stopOnTablet(play.resident.id, playbackId);
        if (result === 'unreachable') {
          update(playbackId, () => ({ stopNote: "Couldn't reach the tablet to stop it. It may still be playing." }));
        } else if (result === 'finished') {
          update(playbackId, () => ({
            stopNote: 'It had already finished on the tablet.',
            localEnd: { phase: 'ended', at: Date.now() },
          }));
        } else {
          logEvent({
            type: 'play_stopped',
            residentId: play.resident.id,
            clipId: play.clip.id,
            atSecond: positionOf(play, Date.now()),
          });
          update(playbackId, () => ({
            localEnd: { phase: before === 'waiting' ? 'missed' : 'stopped', at: Date.now() },
          }));
        }
      } catch {
        update(playbackId, () => ({ stopNote: "That didn't send. Check the connection and try again." }));
      } finally {
        update(playbackId, () => ({ stopping: false }));
      }
    },
    [update]
  );

  const rate = useCallback(
    async (playbackId: number, rating: Rating) => {
      const play = latest.current.find((p) => p.playbackId === playbackId);
      if (!play || play.ratingPending) return;
      update(playbackId, () => ({ ratingPending: rating, ratingNote: null }));
      try {
        // Closed only when tapped right as the window ends; then rating simply goes away
        if ((await rateClip(play.resident.id, playbackId, rating)) === 'closed') {
          update(playbackId, () => ({ ratingPending: null, ratingClosed: true }));
          return;
        }
      } catch {
        update(playbackId, () => ({
          ratingPending: null,
          ratingNote: "That didn't save. Check the connection and try again.",
        }));
        return;
      }
      logEvent({ type: 'clip_rated', residentId: play.resident.id, clipId: play.clip.id, rating });
      update(playbackId, () => ({ ratingPending: null, rated: rating }));
    },
    [update]
  );

  const changeRating = useCallback((playbackId: number) => update(playbackId, () => ({ rated: null })), [update]);
  const dismiss = useCallback((playbackId: number) => update(playbackId, () => ({ dismissed: true })), [update]);

  // Ask the server about every play that still has something to tell
  const hasPlays = plays.length > 0;
  useEffect(() => {
    if (!hasPlays) return;
    let active = true;
    const check = () => {
      for (const p of latest.current) {
        if (settled(p, Date.now())) continue;
        getPlayState(p.resident.id, p.playbackId)
          .then((state) => active && update(p.playbackId, () => ({ state })))
          // A missed check is retried on the next one
          .catch(() => undefined);
      }
    };
    check();
    const timer = setInterval(check, POLL_MS);
    return () => {
      active = false;
      clearInterval(timer);
    };
  }, [hasPlays, update]);

  // Note when a rating window closes on a play nobody rated
  const nextClose = plays
    .filter((p) => !p.ratingClosed && p.rated === null && p.state?.ratingClosesAt != null)
    .reduce((soonest, p) => Math.min(soonest, p.state!.ratingClosesAt!), Infinity);
  useEffect(() => {
    if (nextClose === Infinity) return;
    const timer = setTimeout(() => {
      const now = Date.now();
      for (const p of latest.current) {
        const closesAt = p.state?.ratingClosesAt;
        if (p.ratingClosed || p.rated !== null || closesAt == null || closesAt > now) continue;
        if (!p.replaced && p.state?.startedAt != null && !p.state.failed) {
          logEvent({ type: 'rating_window_expired', residentId: p.resident.id, clipId: p.clip.id });
        }
        update(p.playbackId, () => ({ ratingClosed: true }));
      }
    }, Math.max(0, nextClose - Date.now()));
    return () => clearTimeout(timer);
  }, [nextClose, update]);

  const value = useMemo(
    () => ({ plays, start, stop, rate, changeRating, dismiss, adopt }),
    [plays, start, stop, rate, changeRating, dismiss, adopt]
  );
  return <Context.Provider value={value}>{children}</Context.Provider>;
}

function newPlay(playbackId: number, resident: Resident, clip: Clip, delivered: boolean): ActivePlay {
  return {
    playbackId,
    resident,
    clip,
    sentAt: Date.now(),
    delivered,
    state: null,
    localEnd: null,
    stopping: false,
    stopNote: null,
    ratingPending: null,
    ratingNote: null,
    rated: null,
    ratingClosed: false,
    replaced: false,
    dismissed: false,
  };
}

export function useActivePlays(): ActivePlays {
  const value = useContext(Context);
  if (!value) throw new Error('useActivePlays needs ActivePlaysProvider');
  return value;
}

/** The current time, updated every `everyMs` while `on`; for progress and timers. */
export function useNow(everyMs = 1000, on = true): number {
  const [now, setNow] = useState(() => Date.now());
  useEffect(() => {
    if (!on) return;
    const tick = () => setNow(Date.now());
    // Catch up right away after being off, then keep ticking
    const first = setTimeout(tick, 0);
    const timer = setInterval(tick, everyMs);
    return () => {
      clearTimeout(first);
      clearInterval(timer);
    };
  }, [everyMs, on]);
  return now;
}

export const formatDuration = (seconds: number) =>
  `${Math.floor(seconds / 60)}:${String(Math.floor(seconds % 60)).padStart(2, '0')}`;
