/**
 * machine.ts
 *
 * The tablet's playback state machine: one place that decides what the screen
 * shows. Playback starts because something asks the machine to play (the
 * schedule now, staff commands and taps later), never by navigating. When a
 * play ends, the idle screen comes back by itself.
 *
 * Pure functions only, so every transition is unit tested.
 */

import type { MediaItem } from "./media";

/** Who asked for a play. Matches the server's `Playback.source`. */
export type PlaySource = "scheduled" | "staff" | "tap";

export type Playing = {
  /** Unique per play, so a late report from an earlier play is ignored. */
  playId: number;
  item: MediaItem;
  source: PlaySource;
};

export type PlaybackState = {
  /** "checking" until the stored pairing has been read at startup. */
  pairing: "checking" | "paired" | "unpaired";
  night: boolean;
  playing: Playing | null;
  /** The playId the next play gets. */
  nextPlayId: number;
};

export type PlaybackEvent =
  | { type: "pairingChecked"; paired: boolean }
  | { type: "nightChanged"; night: boolean }
  | { type: "play"; item: MediaItem; source: PlaySource }
  | { type: "stop" }
  /** The player reported finished or failed. */
  | { type: "ended"; playId: number };

/** What the screen shows. */
export type Screen = "starting" | "unpaired" | "playing" | "night" | "alwaysOn";

export const initialPlaybackState: PlaybackState = {
  pairing: "checking",
  night: false,
  playing: null,
  nextPlayId: 1,
};

/**
 * Whether a play from `source` would start now. Nothing plays until the tablet
 * is paired. A staff play pre-empts anything, including another staff play
 * (the newest wins); a scheduled or tapped play only starts when nothing is
 * playing, so it never interrupts a staff clip.
 */
export function canPlay(state: PlaybackState, source: PlaySource): boolean {
  if (state.pairing !== "paired") return false;
  if (!state.playing) return true;
  return source === "staff";
}

/**
 * The next state after `event`, for `useReducer` in `PlaybackProvider`.
 *
 * - `pairingChecked`: records whether the tablet is paired. Losing pairing
 *   also ends any play.
 * - `nightChanged`: records whether it is night. A play in progress carries
 *   on; `screenFor` shows the night screen once it ends.
 * - `play`: starts the item with a new `playId` if `canPlay` allows it, and
 *   is ignored otherwise.
 * - `stop`: ends whatever is playing.
 * - `ended`: ends the play only if `playId` is the one on screen, so a late
 *   report from an earlier play is ignored.
 */
export function playbackReducer(state: PlaybackState, event: PlaybackEvent): PlaybackState {
  switch (event.type) {
    case "pairingChecked": {
      const pairing = event.paired ? "paired" : "unpaired";
      return { ...state, pairing, playing: event.paired ? state.playing : null };
    }
    case "nightChanged":
      return event.night === state.night ? state : { ...state, night: event.night };
    case "play":
      if (!canPlay(state, event.source)) return state;
      return {
        ...state,
        playing: { playId: state.nextPlayId, item: event.item, source: event.source },
        nextPlayId: state.nextPlayId + 1,
      };
    case "stop":
      return state.playing ? { ...state, playing: null } : state;
    case "ended":
      return state.playing?.playId === event.playId ? { ...state, playing: null } : state;
  }
}

/**
 * The screen for a state. Playing wins over night, so a clip that runs past
 * the start of night finishes, and the night screen follows when it ends.
 */
export function screenFor(state: PlaybackState): Screen {
  if (state.pairing === "checking") return "starting";
  if (state.pairing === "unpaired") return "unpaired";
  if (state.playing) return "playing";
  return state.night ? "night" : "alwaysOn";
}
