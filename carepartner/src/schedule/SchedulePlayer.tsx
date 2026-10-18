/**
 * SchedulePlayer.tsx
 *
 * Plays scheduled media when it is due, through the playback machine, one at
 * a time. Draws nothing, and never navigates: the screen follows the machine.
 *
 * - As each minute starts, and whenever the screen or the machine changes, plays
 *   the due play (see `duePlay`) if `canPlay` allows a scheduled play. A staff
 *   clip is never interrupted; a play still due when it ends starts then.
 * - A play counts as played once the machine puts it on screen, so a failed
 *   one doesn't loop, and one the machine refuses (a staff clip got there
 *   first) can still play later. What has played is stored on the device, so
 *   a restart doesn't repeat it.
 */

import { useEffect, useRef, useState } from "react";
import { dayOf, everyMinute, now } from "../clock/clock";
import { canPlay } from "../playback/machine";
import { usePlayback } from "../playback/PlaybackProvider";
import type { ScheduledPlay } from "../screen/screen";
import { useScreen } from "../screen/ScreenProvider";
import { loadPlayed, savePlayed } from "./playedStore";
import { duePlay, playKey } from "./schedule";

export default function SchedulePlayer() {
  const { state, play } = usePlayback();
  const schedule = useScreen().screen?.schedule;
  /** Null until loaded from storage; nothing plays before then. */
  const [played, setPlayed] = useState<Set<string> | null>(null);
  const [time, setTime] = useState(() => now());
  /** The play last asked for, until it is on screen. */
  const requested = useRef<ScheduledPlay | null>(null);

  useEffect(() => {
    let mounted = true;
    loadPlayed().then((stored) => {
      if (mounted) setPlayed(stored);
    });
    const stop = everyMinute(() => setTime(now()));
    return () => {
      mounted = false;
      stop();
    };
  }, []);

  useEffect(() => {
    if (!played) return;
    const asked = requested.current;
    // The machine has put the play asked for on screen, so it counts as played.
    if (asked && state.playing?.source === "scheduled" && state.playing.item === asked.item) {
      requested.current = null;
      const updated = new Set(played).add(playKey(asked));
      setPlayed(updated);
      savePlayed(updated, dayOf(now()));
      return;
    }
    if (!schedule || !canPlay(state, "scheduled")) return;
    const due = duePlay(schedule, now(), played);
    if (!due) return;
    requested.current = due;
    play(due.item, "scheduled", { scheduledItemId: due.scheduledItemId });
  }, [schedule, played, state, play, time]);

  return null;
}
