import React from "react";
import { act, render } from "@testing-library/react-native";
import { setClock } from "../../clock/clock";
import { getCarepartnerId } from "../../pairing/carepartnerId";
import type { MediaItem } from "../media";
import { NIGHT_CHECK_MS, PlaybackProvider, usePlayback } from "../PlaybackProvider";

jest.mock("../../pairing/carepartnerId", () => ({ getCarepartnerId: jest.fn() }));

const song: MediaItem = { kind: "audio", url: "/media/1/file" };

let playback: ReturnType<typeof usePlayback>;
function Probe() {
  playback = usePlayback();
  return null;
}

async function start(time = "2026-10-12T12:00:00") {
  setClock(new Date(time));
  await render(
    <PlaybackProvider>
      <Probe />
    </PlaybackProvider>,
  );
  await act(async () => {});
}

describe("PlaybackProvider", () => {
  beforeEach(() => {
    jest.useFakeTimers();
    jest.mocked(getCarepartnerId).mockResolvedValue(7);
  });

  afterEach(() => {
    setClock(null);
    jest.useRealTimers();
    jest.clearAllMocks();
  });

  it("counts a tablet with a stored carepartner as paired", async () => {
    await start();
    expect(playback.screen).toBe("alwaysOn");
    expect(playback.state.carepartnerId).toBe(7);
  });

  it("shows pairing when no carepartner is stored", async () => {
    jest.mocked(getCarepartnerId).mockResolvedValue(null);
    await start();
    expect(playback.screen).toBe("unpaired");
  });

  it("picks up a new pairing when asked to check again", async () => {
    jest.mocked(getCarepartnerId).mockResolvedValue(null);
    await start();

    jest.mocked(getCarepartnerId).mockResolvedValue(7);
    await act(() => playback.checkPairing());

    expect(playback.screen).toBe("alwaysOn");
  });

  it("starts at night when it is night", async () => {
    await start("2026-10-12T23:00:00");
    expect(playback.screen).toBe("night");
  });

  it("switches to night when night starts", async () => {
    await start("2026-10-12T19:59:30");
    expect(playback.screen).toBe("alwaysOn");

    await act(async () => jest.advanceTimersByTime(NIGHT_CHECK_MS));

    expect(playback.screen).toBe("night");
  });

  it("plays, ends and stops through the machine", async () => {
    await start();

    await act(async () => playback.play(song, "scheduled"));
    expect(playback.screen).toBe("playing");
    expect(playback.state.playing).toMatchObject({ item: song, source: "scheduled" });

    await act(async () => playback.ended(playback.state.playing!.playId, "finished"));
    expect(playback.screen).toBe("alwaysOn");

    await act(async () => playback.play(song, "staff"));
    await act(async () => playback.stop());
    expect(playback.screen).toBe("alwaysOn");
  });
});
