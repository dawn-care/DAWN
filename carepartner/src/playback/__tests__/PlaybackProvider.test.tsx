import React from "react";
import { act, render } from "@testing-library/react-native";
import { setClock } from "../../clock/clock";
import { readPairedId } from "../../pairing/pairingState";
import type { MediaItem } from "../media";
import { PlaybackProvider, usePlayback } from "../PlaybackProvider";

jest.mock("../../pairing/pairingState", () => ({ readPairedId: jest.fn() }));

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
    jest.mocked(readPairedId).mockResolvedValue(7);
  });

  afterEach(() => {
    setClock(null);
    jest.useRealTimers();
    jest.clearAllMocks();
  });

  it("counts a paired tablet as paired, with its carepartner", async () => {
    await start();
    expect(playback.screen).toBe("alwaysOn");
    expect(playback.state.carepartnerId).toBe(7);
  });

  it("shows pairing when the tablet isn't paired", async () => {
    jest.mocked(readPairedId).mockResolvedValue(null);
    await start();
    expect(playback.screen).toBe("unpaired");
  });

  it("picks up a new pairing when asked to check again", async () => {
    jest.mocked(readPairedId).mockResolvedValue(null);
    await start();

    jest.mocked(readPairedId).mockResolvedValue(7);
    await act(() => playback.checkPairing());

    expect(playback.screen).toBe("alwaysOn");
  });

  it("shows night when told night has started, and the day screen when it has ended", async () => {
    await start();
    expect(playback.screen).toBe("alwaysOn");

    await act(async () => playback.setNight(true));
    expect(playback.screen).toBe("night");

    await act(async () => playback.setNight(false));
    expect(playback.screen).toBe("alwaysOn");
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
