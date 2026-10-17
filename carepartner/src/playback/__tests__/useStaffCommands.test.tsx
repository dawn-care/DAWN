import React from "react";
import { act, render } from "@testing-library/react-native";
import { setClock } from "../../clock/clock";
import { getCarepartnerId } from "../../pairing/carepartnerId";
import type { MediaItem } from "../media";
import { PlaybackProvider, usePlayback } from "../PlaybackProvider";
import { connectTablet, type TabletConnection } from "../socket";
import type { ServerMessage } from "../socketMessages";
import StaffCommands from "../StaffCommands";
import { fetchMediaItem } from "../staffItem";

jest.mock("../../pairing/carepartnerId", () => ({ getCarepartnerId: jest.fn() }));
jest.mock("../staffItem", () => ({ fetchMediaItem: jest.fn() }));
jest.mock("../socket", () => ({ ...jest.requireActual("../socket"), connectTablet: jest.fn() }));
const mockRefresh = jest.fn();
jest.mock("../../screen/ScreenProvider", () => ({ useScreen: () => ({ refresh: mockRefresh }) }));

const NOW = new Date("2026-10-14T10:00:00");
const video: MediaItem = { kind: "video", url: "/media/12/file" };
const song: MediaItem = { kind: "audio", url: "/media/3/file" };

let playback: ReturnType<typeof usePlayback>;
function Probe() {
  playback = usePlayback();
  return null;
}

/** The fake connection: what the tablet sent, and a way to deliver server messages. */
let connection: { send: jest.Mock; close: jest.Mock; url: string; receive: (message: ServerMessage) => Promise<void> };

const playCommand = (playbackId: number, secondsLeft = 30): ServerMessage => ({
  type: "play",
  playbackId,
  mediaId: 9,
  expiresAt: new Date(NOW.getTime() + secondsLeft * 1000),
});
/** The server's answer to the tablet's acknowledgement. */
const answer = (playbackId: number, status = "acknowledged"): ServerMessage => ({ type: "status_result", playbackId, status });
/** A play command the server accepts as soon as the tablet acknowledges it. */
async function playAccepted(playbackId: number) {
  await connection.receive(playCommand(playbackId));
  await connection.receive(answer(playbackId));
}
/** The statuses reported so far, as [playback id, status]. */
const reports = () =>
  connection.send.mock.calls.map(([text]) => {
    const { playback_id, status } = JSON.parse(text);
    return [playback_id, status];
  });

async function start() {
  const view = await render(
    <PlaybackProvider>
      <StaffCommands />
      <Probe />
    </PlaybackProvider>,
  );
  await act(async () => {});
  return view;
}

describe("useStaffCommands", () => {
  let warn: jest.SpyInstance;

  beforeEach(() => {
    setClock(NOW);
    warn = jest.spyOn(console, "warn").mockImplementation(() => {});
    jest.mocked(getCarepartnerId).mockResolvedValue(7);
    jest.mocked(fetchMediaItem).mockResolvedValue(video);
    jest.mocked(connectTablet).mockImplementation(({ url, onMessage }) => {
      const send = jest.fn();
      const close = jest.fn();
      // Delivers a message without waiting for the tablet to finish with it,
      // since some tests hold a lookup open.
      connection = {
        send,
        close,
        url,
        receive: (message) =>
          act(async () => {
            onMessage(message);
          }),
      };
      return { send, close } as TabletConnection;
    });
  });

  afterEach(() => {
    setClock(null);
    warn.mockRestore();
    jest.clearAllMocks();
  });

  it("connects once paired, to this tablet's socket", async () => {
    await start();
    expect(connectTablet).toHaveBeenCalledTimes(1);
    expect(connection.url).toMatch(/\/realtime\/tablet\?carepartner_id=7$/);
  });

  it("doesn't connect while unpaired", async () => {
    jest.mocked(getCarepartnerId).mockResolvedValue(null);
    await start();
    expect(connectTablet).not.toHaveBeenCalled();
  });

  it("closes the connection when it stops", async () => {
    const view = await start();
    await act(async () => view.unmount());
    expect(connection.close).toHaveBeenCalled();
  });

  it("acknowledges a play, plays it as staff once the server accepts, and reports it started, then finished", async () => {
    await start();
    await playAccepted(40);

    expect(fetchMediaItem).toHaveBeenCalledWith(9);
    expect(playback.state.playing).toMatchObject({ item: video, source: "staff", playbackId: 40 });
    expect(reports()).toEqual([
      [40, "acknowledged"],
      [40, "started"],
    ]);

    await act(async () => playback.ended(playback.state.playing!.playId, "finished"));

    expect(playback.screen).toBe("alwaysOn");
    expect(reports()).toContainEqual([40, "finished"]);
  });

  it("waits for the server to accept the acknowledgement before playing", async () => {
    await start();
    await connection.receive(playCommand(40));

    expect(playback.state.playing).toBeNull();

    await connection.receive(answer(40));
    expect(playback.state.playing?.playbackId).toBe(40);
  });

  it("leaves expiry to the server, not the tablet's clock", async () => {
    await start();
    // Already past its deadline by the tablet's clock, which may be wrong.
    await connection.receive(playCommand(40, -60));
    await connection.receive(answer(40));

    expect(playback.state.playing?.playbackId).toBe(40);
  });

  it.each(["expired", undefined])("plays nothing when the server answers %s", async (status) => {
    await start();
    await connection.receive(playCommand(40));

    await connection.receive({ type: "status_result", playbackId: 40, status, error: status ? undefined : "invalid" });
    // A later stray answer doesn't revive it.
    await connection.receive(answer(40));

    expect(playback.state.playing).toBeNull();
    expect(reports()).toEqual([[40, "acknowledged"]]);
  });

  it("reports failed when the media can't be looked up or played", async () => {
    await start();
    jest.mocked(fetchMediaItem).mockResolvedValueOnce(null);
    await playAccepted(40);

    expect(reports()).toEqual([
      [40, "acknowledged"],
      [40, "failed"],
    ]);
    expect(playback.state.playing).toBeNull();
  });

  it("ignores a command it has already handled, as the server resends it on reconnect", async () => {
    await start();
    await playAccepted(40);
    await connection.receive(playCommand(40));

    expect(fetchMediaItem).toHaveBeenCalledTimes(1);
    expect(reports().filter(([, status]) => status === "acknowledged")).toHaveLength(1);
  });

  it("acknowledges a resent command again if the server's answer never came", async () => {
    await start();
    await connection.receive(playCommand(40));
    await connection.receive(playCommand(40));

    expect(fetchMediaItem).toHaveBeenCalledTimes(1);
    expect(reports()).toEqual([
      [40, "acknowledged"],
      [40, "acknowledged"],
    ]);

    await connection.receive(answer(40));
    expect(playback.state.playing?.playbackId).toBe(40);
  });

  it("interrupts a scheduled play", async () => {
    await start();
    await act(async () => playback.play(song, "scheduled"));

    await playAccepted(40);

    expect(playback.state.playing).toMatchObject({ source: "staff", playbackId: 40 });
  });

  it("reports the earlier clip stopped when another staff play replaces it", async () => {
    await start();
    await playAccepted(40);
    await playAccepted(41);

    expect(playback.state.playing?.playbackId).toBe(41);
    expect(reports()).toEqual(
      expect.arrayContaining([
        [40, "stopped"],
        [41, "started"],
      ]),
    );
  });

  it("stops the clip on screen and reports it stopped", async () => {
    await start();
    await playAccepted(40);

    await connection.receive({ type: "stop", playbackId: 40 });

    expect(playback.screen).toBe("alwaysOn");
    expect(reports()).toContainEqual([40, "stopped"]);
  });

  it("ignores a stop for a clip that isn't on screen", async () => {
    await start();
    await playAccepted(41);

    await connection.receive({ type: "stop", playbackId: 40 });

    expect(playback.state.playing?.playbackId).toBe(41);
  });

  describe("while a command is being looked up", () => {
    const lookups = new Map<number, (item: MediaItem | null) => void>();
    beforeEach(() => {
      lookups.clear();
      jest.mocked(fetchMediaItem).mockImplementation(
        (mediaId) =>
          new Promise((resolve) => {
            lookups.set(mediaId, resolve);
          }),
      );
    });

    it("cancels it on a stop, reports it stopped and never plays it", async () => {
      await start();
      await playAccepted(40);

      await connection.receive({ type: "stop", playbackId: 40 });
      await act(async () => lookups.get(9)!(video));

      expect(reports()).toEqual([
        [40, "acknowledged"],
        [40, "stopped"],
      ]);
      expect(playback.state.playing).toBeNull();
    });

    it("cancels it when a newer command arrives, so the older one can't take over", async () => {
      await start();
      await playAccepted(40);
      await connection.receive({ ...playCommand(41), mediaId: 10 } as ServerMessage);
      await connection.receive(answer(41));

      // The newer lookup finishes first, then the older one.
      await act(async () => lookups.get(10)!(song));
      await act(async () => lookups.get(9)!(video));

      expect(playback.state.playing).toMatchObject({ item: song, playbackId: 41 });
      expect(reports()).toEqual([
        [40, "acknowledged"],
        [40, "stopped"],
        [41, "acknowledged"],
        [41, "started"],
      ]);
    });

    it("plays it once both the lookup and the server's answer are in, in either order", async () => {
      await start();
      await connection.receive(playCommand(40));
      await act(async () => lookups.get(9)!(video));

      expect(playback.state.playing).toBeNull();

      await connection.receive(answer(40));
      expect(playback.state.playing?.playbackId).toBe(40);
    });
  });

  it("fetches the schedule again when the server says to refresh", async () => {
    await start();
    await connection.receive({ type: "refresh" });
    expect(mockRefresh).toHaveBeenCalled();
  });

  it("warns when the server refuses a status", async () => {
    await start();
    await connection.receive({ type: "status_result", playbackId: 40, error: "invalid playback status transition" });

    expect(warn).toHaveBeenCalledWith("Server refused a status for playback 40: invalid playback status transition");
  });
});
