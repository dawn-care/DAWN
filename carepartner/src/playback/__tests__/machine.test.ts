import type { MediaItem } from "../media";
import {
  canPlay,
  initialPlaybackState,
  playbackReducer,
  screenFor,
  type PlaybackEvent,
  type PlaybackState,
  type PlaySource,
} from "../machine";

const song: MediaItem = { module_type: "audio", metadata: { id: 1 } };
const photo: MediaItem = { module_type: "photo", metadata: { id: 2 } };

const run = (state: PlaybackState, ...events: PlaybackEvent[]) => events.reduce(playbackReducer, state);
const paired = run(initialPlaybackState, { type: "pairingChecked", paired: true });
const playing = (source: PlaySource, item = song) => run(paired, { type: "play", item, source });

describe("playback machine", () => {
  describe("startup and pairing", () => {
    it("starts out checking pairing, showing nothing", () => {
      expect(screenFor(initialPlaybackState)).toBe("starting");
    });

    it("shows pairing when the tablet isn't paired", () => {
      const state = run(initialPlaybackState, { type: "pairingChecked", paired: false });
      expect(screenFor(state)).toBe("unpaired");
    });

    it("shows the always-on screen once paired", () => {
      expect(screenFor(paired)).toBe("alwaysOn");
    });

    it.each(["scheduled", "staff", "tap"] as const)("plays nothing from %s before pairing is known or while unpaired", (source) => {
      const unpaired = run(initialPlaybackState, { type: "pairingChecked", paired: false });

      expect(run(initialPlaybackState, { type: "play", item: song, source }).playing).toBeNull();
      expect(run(unpaired, { type: "play", item: song, source }).playing).toBeNull();
    });

    it("stops playing and shows pairing when the tablet becomes unpaired", () => {
      const state = run(playing("scheduled"), { type: "pairingChecked", paired: false });

      expect(state.playing).toBeNull();
      expect(screenFor(state)).toBe("unpaired");
    });

    it("keeps playing when pairing is confirmed again", () => {
      const state = run(playing("scheduled"), { type: "pairingChecked", paired: true });
      expect(state.playing?.item).toBe(song);
    });
  });

  describe("day and night", () => {
    it("shows the night screen at night and the always-on screen by day", () => {
      const night = run(paired, { type: "nightChanged", night: true });
      expect(screenFor(night)).toBe("night");
      expect(screenFor(run(night, { type: "nightChanged", night: false }))).toBe("alwaysOn");
    });

    it("leaves the state as it is when night hasn't changed", () => {
      expect(run(paired, { type: "nightChanged", night: false })).toBe(paired);
    });

    it("keeps playing when night starts, and shows night when the play ends", () => {
      const state = run(playing("scheduled"), { type: "nightChanged", night: true });
      expect(screenFor(state)).toBe("playing");

      expect(screenFor(run(state, { type: "ended", playId: state.playing!.playId }))).toBe("night");
    });

    it("can play at night", () => {
      const state = run(paired, { type: "nightChanged", night: true }, { type: "play", item: song, source: "scheduled" });
      expect(screenFor(state)).toBe("playing");
    });
  });

  describe("playing", () => {
    it.each(["scheduled", "staff", "tap"] as const)("starts a %s play when idle", (source) => {
      const state = playing(source);

      expect(state.playing).toEqual({ playId: 1, item: song, source });
      expect(screenFor(state)).toBe("playing");
    });

    it("gives every play a new id", () => {
      const first = playing("scheduled");
      const second = run(first, { type: "ended", playId: 1 }, { type: "play", item: photo, source: "scheduled" });

      expect(second.playing?.playId).toBe(2);
    });

    it.each([
      ["staff", "scheduled", true],
      ["staff", "staff", true],
      ["staff", "tap", true],
      ["scheduled", "scheduled", false],
      ["scheduled", "staff", false],
      ["scheduled", "tap", false],
      ["tap", "scheduled", false],
      ["tap", "staff", false],
      ["tap", "tap", false],
    ] as const)("a %s play over a %s play replaces it: %s", (incoming, current, replaces) => {
      const before = playing(current);
      const after = run(before, { type: "play", item: photo, source: incoming });

      expect(canPlay(before, incoming)).toBe(replaces);
      expect(after.playing).toEqual(replaces ? { playId: 2, item: photo, source: incoming } : before.playing);
    });
  });

  describe("ending", () => {
    it("returns to the always-on screen when the play ends", () => {
      const state = run(playing("scheduled"), { type: "ended", playId: 1 });

      expect(state.playing).toBeNull();
      expect(screenFor(state)).toBe("alwaysOn");
    });

    it("ignores a late report from a play that was replaced", () => {
      const replaced = run(playing("scheduled"), { type: "play", item: photo, source: "staff" });

      expect(run(replaced, { type: "ended", playId: 1 })).toBe(replaced);
    });

    it("ignores an ended report when nothing is playing", () => {
      expect(run(paired, { type: "ended", playId: 1 })).toBe(paired);
    });

    it("returns to the always-on screen when stopped", () => {
      const state = run(playing("staff"), { type: "stop" });
      expect(screenFor(state)).toBe("alwaysOn");
    });

    it("ignores stop when nothing is playing", () => {
      expect(run(paired, { type: "stop" })).toBe(paired);
    });
  });
});
