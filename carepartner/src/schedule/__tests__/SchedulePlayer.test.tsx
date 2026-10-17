import React from "react";
import { act, render } from "@testing-library/react-native";
import { setClock } from "../../clock/clock";
import { readPairedId } from "../../pairing/pairingState";
import * as playbackProvider from "../../playback/PlaybackProvider";
import { PlaybackProvider, usePlayback } from "../../playback/PlaybackProvider";
import type { ScheduledPlay } from "../../screen/screen";
import { useScreen } from "../../screen/ScreenProvider";
import { loadPlayed, savePlayed } from "../playedStore";
import SchedulePlayer, { SCHEDULE_CHECK_MS } from "../SchedulePlayer";

jest.mock("../../pairing/pairingState", () => ({ readPairedId: jest.fn() }));
jest.mock("../../screen/ScreenProvider", () => ({ useScreen: jest.fn() }));
jest.mock("../playedStore", () => ({ loadPlayed: jest.fn(), savePlayed: jest.fn() }));

const play = (scheduledItemId: number, time: string): ScheduledPlay => ({
  scheduledItemId,
  date: "2026-10-12",
  time,
  outsideAwakeHours: false,
  item: { kind: "video", url: `/media/${scheduledItemId}/file` },
});

/** A morning message at 8:00 and a song at 8:15. */
const schedule = [play(1, "08:00"), play(2, "08:15")];

let playback: ReturnType<typeof usePlayback>;
function Probe() {
  playback = usePlayback();
  return null;
}

const flush = () => act(async () => {});
const tick = (ms = SCHEDULE_CHECK_MS) => act(async () => jest.advanceTimersByTime(ms));
const playingUrl = () => playback.state.playing?.item.url ?? null;
const endPlay = () => act(async () => playback.ended(playback.state.playing!.playId, "finished"));

async function start(time: string, played: string[] = []) {
  setClock(new Date(time));
  jest.mocked(loadPlayed).mockResolvedValue(new Set(played));
  await render(
    <PlaybackProvider>
      <SchedulePlayer />
      <Probe />
    </PlaybackProvider>,
  );
  await flush();
  await flush();
}

describe("SchedulePlayer", () => {
  beforeEach(() => {
    jest.useFakeTimers();
    jest.mocked(readPairedId).mockResolvedValue(7);
    jest.mocked(useScreen).mockReturnValue({ screen: { schedule } as never, refresh: jest.fn(), fetchedAt: null });
  });

  afterEach(() => {
    jest.restoreAllMocks();
    setClock(null);
    jest.useRealTimers();
    jest.clearAllMocks();
  });

  it("plays nothing before its time", async () => {
    await start("2026-10-12T07:59:00");

    expect(playingUrl()).toBeNull();
  });

  it("plays an item when its time comes, as a scheduled play", async () => {
    await start("2026-10-12T07:59:50");
    await tick();

    expect(playingUrl()).toBe("/media/1/file");
    expect(playback.state.playing!.source).toBe("scheduled");
  });

  it("still plays an item a few minutes late, such as after a restart", async () => {
    await start("2026-10-12T08:07:00");

    expect(playingUrl()).toBe("/media/1/file");
  });

  it("skips an item once its grace window has passed", async () => {
    await start("2026-10-12T08:11:00");
    await tick();

    expect(playingUrl()).toBeNull();
  });

  it("remembers what played, on the device", async () => {
    await start("2026-10-12T08:00:30");

    expect(savePlayed).toHaveBeenCalledWith(new Set(["1@2026-10-12"]), "2026-10-12");
  });

  it("doesn't play an item again after a restart", async () => {
    await start("2026-10-12T08:03:00", ["1@2026-10-12"]);

    expect(playingUrl()).toBeNull();
  });

  it("waits for what was played to load before playing anything", async () => {
    jest.mocked(loadPlayed).mockReturnValue(new Promise(() => {}));
    setClock(new Date("2026-10-12T08:01:00"));
    await render(
      <PlaybackProvider>
        <SchedulePlayer />
        <Probe />
      </PlaybackProvider>,
    );
    await flush();
    await tick();

    expect(playingUrl()).toBeNull();
  });

  it("plays an item once, even after it ends", async () => {
    await start("2026-10-12T08:00:30");
    await endPlay();
    await tick();

    expect(playingUrl()).toBeNull();
  });

  it("never interrupts a staff clip, and plays the item after it if still in time", async () => {
    await start("2026-10-12T07:59:00");
    await act(async () => playback.play({ kind: "video", url: "/media/99/file" }, "staff", 5));
    await tick(2 * 60_000);

    expect(playingUrl()).toBe("/media/99/file");

    await endPlay();

    expect(playingUrl()).toBe("/media/1/file");
  });

  it("plays a later item at its own time", async () => {
    await start("2026-10-12T08:14:30");
    // The 8:00 item has passed its grace window; only the 8:15 one plays.
    expect(playingUrl()).toBeNull();
    await tick(2 * SCHEDULE_CHECK_MS);

    expect(playingUrl()).toBe("/media/2/file");
  });

  it("doesn't count a play the machine refused, and asks again", async () => {
    const realUsePlayback = playbackProvider.usePlayback;
    const refusedPlay = jest.fn();
    jest.spyOn(playbackProvider, "usePlayback").mockImplementation(() => ({ ...realUsePlayback(), play: refusedPlay }));
    await start("2026-10-12T08:00:30");
    await tick();

    expect(refusedPlay).toHaveBeenCalledTimes(2);
    expect(savePlayed).not.toHaveBeenCalled();
  });

  it("plays nothing without a screen yet", async () => {
    jest.mocked(useScreen).mockReturnValue({ screen: null, refresh: jest.fn(), fetchedAt: null });
    await start("2026-10-12T08:00:30");

    expect(playingUrl()).toBeNull();
  });
});
