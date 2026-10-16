/**
 * SchedulePlayer.tsx
 *
 * Plays scheduled media when it is due, through the playback machine, one at
 * a time. Draws nothing, and never navigates: the screen follows the machine.
 *
 * - Every 15 seconds, and whenever the screen or the machine changes, plays
 *   the due play (see `duePlay`) if `canPlay` allows a scheduled play. A staff
 *   clip is never interrupted; a play still due when it ends starts then.
 * - A play counts as played once it starts, so a failed one doesn't loop.
 *   What has played is stored on the device, so a restart doesn't repeat it.
 */

import { useEffect, useState } from "react";
import { dayOf, now } from "../clock/clock";
import { canPlay } from "../playback/machine";
import { usePlayback } from "../playback/PlaybackProvider";
import { useScreen } from "../screen/ScreenProvider";
import { loadPlayed, savePlayed } from "./playedStore";
import { duePlay, playKey } from "./schedule";

export const SCHEDULE_CHECK_MS = 15_000;

export default function SchedulePlayer() {
  const { state, play } = usePlayback();
  const schedule = useScreen().screen?.schedule;
  /** Null until loaded from storage; nothing plays before then. */
  const [played, setPlayed] = useState<Set<string> | null>(null);
  const [time, setTime] = useState(() => now());

  useEffect(() => {
    let mounted = true;
    loadPlayed().then((stored) => {
      if (mounted) setPlayed(stored);
    });
    const timer = setInterval(() => setTime(now()), SCHEDULE_CHECK_MS);
    return () => {
      mounted = false;
      clearInterval(timer);
    };
  }, []);

  useEffect(() => {
    if (!schedule || !played || !canPlay(state, "scheduled")) return;
    const current = now();
    const due = duePlay(schedule, current, played);
    if (!due) return;
    const updated = new Set(played).add(playKey(due));
    setPlayed(updated);
    savePlayed(updated, dayOf(current));
    play(due.item, "scheduled");
  }, [schedule, played, state, play, time]);

  return null;
}
