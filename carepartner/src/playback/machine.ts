/**
 * machine.ts
 *
 * The tablet's playback state machine: one place that decides what the screen
 * shows. Playback starts because something asks the machine to play (the
 * schedule, staff commands, and taps later), never by navigating. When a play
 * ends, the idle screen comes back by itself.
 *
 * Pure functions only, so every transition is unit tested.
 */

import type { MediaItem } from "./media";

/** Who asked for a play. Matches the server's `Playback.source`. */
export type PlaySource = "scheduled" | "staff" | "tap";

/** What a player reports when it's done. */
export type PlayerOutcome = "finished" | "failed";

export type Playing = {
  /** Unique per play, so a late report from an earlier play is ignored. */
  playId: number;
  item: MediaItem;
  source: PlaySource;
  /** The server's Playback record for this play, if it has one (staff plays for now). */
  playbackId?: number;
  /** For a scheduled play, the scheduled item it comes from. */
  scheduledItemId?: number;
  /**
   * "chime" while the chime and the pause after it play, so the resident
   * isn't startled; then "media" while the player plays the item.
   */
  phase: "chime" | "media";
  /** The player says the media is actually playing; false while it loads. */
  started: boolean;
};

/** Why a play left the screen. */
export type EndReason =
  /** The player reported it finished or failed. */
  | PlayerOutcome
  /** Stopped on request, or because the tablet became unpaired. */
  | "stopped"
  /** A staff play took over the screen. */
  | "replaced";

export type PlaybackState = {
  /** "checking" until the stored pairing has been read at startup. */
  pairing: "checking" | "paired" | "unpaired";
  /** The paired carepartner's id; null unless paired. */
  carepartnerId: number | null;
  night: boolean;
  playing: Playing | null;
  /**
   * The last play to leave the screen, and why. Whatever reports plays to the
   * server watches this rather than every way a play can end.
   */
  lastEnded: { play: Playing; reason: EndReason } | null;
  /** The playId the next play gets. */
  nextPlayId: number;
};

export type PlaybackEvent =
  /** The stored carepartner id, or null if none is stored. */
  | { type: "pairingChecked"; carepartnerId: number | null }
  | { type: "nightChanged"; night: boolean }
  | ({ type: "play"; item: MediaItem; source: PlaySource } & PlayOptions)
  /** Stops the current play, or with `playbackId`, only the play with that record. */
  | { type: "stop"; playbackId?: number }
  /** The chime and the pause after it are over, so the media can play. */
  | { type: "chimed"; playId: number }
  /** The player says the media is actually playing. */
  | { type: "started"; playId: number }
  /** The player reported finished or failed. */
  | { type: "ended"; playId: number; reason: PlayerOutcome };

/** What else a play can carry: its server record (staff), or its scheduled item (scheduled). */
export type PlayOptions = { playbackId?: number; scheduledItemId?: number };

/** What the screen shows. */
export type Screen = "starting" | "unpaired" | "playing" | "night" | "alwaysOn";

export const initialPlaybackState: PlaybackState = {
  pairing: "checking",
  carepartnerId: null,
  night: false,
  playing: null,
  lastEnded: null,
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
 * - `pairingChecked`: records whether the tablet is paired, and to which
 *   carepartner. Losing pairing also ends any play.
 * - `nightChanged`: records whether it is night. A play in progress carries
 *   on; `screenFor` shows the night screen once it ends.
 * - `play`: starts the item with a new `playId` if `canPlay` allows it, and
 *   is ignored otherwise. It starts with its chime (`phase: "chime"`). A play
 *   it replaces ends as "replaced", even during its chime.
 * - `stop`: ends whatever is playing, or with a `playbackId`, only the play
 *   with that record, so a late stop can't end a newer clip.
 * - `chimed`: moves the play on screen from its chime to its media. Like
 *   `ended`, a report for any other `playId` is ignored.
 * - `started`: records that the play on screen is actually playing. Like
 *   `ended`, a report for any other `playId` is ignored.
 * - `ended`: ends the play only if `playId` is the one on screen, so a late
 *   report from an earlier play is ignored.
 *
 * Every play that leaves the screen is recorded in `lastEnded` with its reason.
 */
export function playbackReducer(state: PlaybackState, event: PlaybackEvent): PlaybackState {
  switch (event.type) {
    case "pairingChecked": {
      const { carepartnerId } = event;
      if (carepartnerId == null) return { ...end(state, "stopped"), pairing: "unpaired", carepartnerId };
      return { ...state, pairing: "paired", carepartnerId };
    }
    case "nightChanged":
      return event.night === state.night ? state : { ...state, night: event.night };
    case "play":
      if (!canPlay(state, event.source)) return state;
      return {
        ...end(state, "replaced"),
        playing: {
          playId: state.nextPlayId,
          item: event.item,
          source: event.source,
          playbackId: event.playbackId,
          scheduledItemId: event.scheduledItemId,
          phase: "chime",
          started: false,
        },
        nextPlayId: state.nextPlayId + 1,
      };
    case "stop":
      if (event.playbackId !== undefined && state.playing?.playbackId !== event.playbackId) return state;
      return end(state, "stopped");
    case "chimed":
      if (state.playing?.playId !== event.playId || state.playing.phase !== "chime") return state;
      return { ...state, playing: { ...state.playing, phase: "media" } };
    case "started":
      if (state.playing?.playId !== event.playId || state.playing.started) return state;
      return { ...state, playing: { ...state.playing, started: true } };
    case "ended":
      return state.playing?.playId === event.playId ? end(state, event.reason) : state;
  }
}

/** Takes the current play off the screen for `reason`; the same state if nothing is playing. */
function end(state: PlaybackState, reason: EndReason): PlaybackState {
  if (!state.playing) return state;
  return { ...state, playing: null, lastEnded: { play: state.playing, reason } };
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
