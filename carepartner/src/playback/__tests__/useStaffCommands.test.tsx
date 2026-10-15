import React from "react";
import { act, render } from "@testing-library/react-native";
import { setClock } from "../../clock/clock";
import { getCarepartnerId } from "../../pairing/carepartnerId";
import type { MediaItem } from "../media";
import { PlaybackProvider, usePlayback } from "../PlaybackProvider";
import { connectTablet, type TabletConnection } from "../socket";
import type { ServerMessage } from "../socketMessages";
import StaffCommands from "../StaffCommands";
import { fetchModuleItem } from "../staffItem";

jest.mock("../../pairing/carepartnerId", () => ({ getCarepartnerId: jest.fn() }));
jest.mock("../staffItem", () => ({ fetchModuleItem: jest.fn() }));
jest.mock("../socket", () => ({ ...jest.requireActual("../socket"), connectTablet: jest.fn() }));
const mockRefresh = jest.fn();
jest.mock("../../schedule/ScheduleProvider", () => ({ useSchedule: () => ({ refresh: mockRefresh }) }));

const NOW = new Date("2026-10-14T10:00:00");
const video: MediaItem = { module_type: "video", metadata: { id: 12 } };
const song: MediaItem = { module_type: "audio", metadata: { id: 3 } };

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
  moduleId: 7,
  expiresAt: new Date(NOW.getTime() + secondsLeft * 1000),
});
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
    jest.mocked(fetchModuleItem).mockResolvedValue(video);
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

  it("acknowledges a play, plays it as staff and reports it started, then finished", async () => {
    await start();
    await connection.receive(playCommand(40));

    expect(fetchModuleItem).toHaveBeenCalledWith(7);
    expect(playback.state.playing).toMatchObject({ item: video, source: "staff", playbackId: 40 });
    expect(reports()).toEqual([
      [40, "acknowledged"],
      [40, "started"],
    ]);

    await act(async () => playback.ended(playback.state.playing!.playId, "finished"));

    expect(playback.screen).toBe("alwaysOn");
    expect(reports()).toContainEqual([40, "finished"]);
  });

  it("reports failed when the player fails", async () => {
    await start();
    await connection.receive(playCommand(40));

    await act(async () => playback.ended(playback.state.playing!.playId, "failed"));

    expect(reports()).toContainEqual([40, "failed"]);
  });

  it("reports a command that arrived too late as expired, and plays nothing", async () => {
    await start();
    await connection.receive(playCommand(40, 0));

    expect(reports()).toEqual([[40, "expired"]]);
    expect(fetchModuleItem).not.toHaveBeenCalled();
    expect(playback.state.playing).toBeNull();
  });

  it("reports failed when the module can't be looked up or played", async () => {
    await start();
    jest.mocked(fetchModuleItem).mockResolvedValueOnce(null);
    await connection.receive(playCommand(40));
    jest.mocked(fetchModuleItem).mockResolvedValueOnce({ module_type: "weather", metadata: { id: 1 } });
    await connection.receive(playCommand(41));

    expect(reports()).toEqual([
      [40, "acknowledged"],
      [40, "failed"],
      [41, "acknowledged"],
      [41, "failed"],
    ]);
    expect(playback.state.playing).toBeNull();
  });

  it("ignores a command it has already handled, as the server resends it on reconnect", async () => {
    await start();
    await connection.receive(playCommand(40));
    await connection.receive(playCommand(40));

    expect(fetchModuleItem).toHaveBeenCalledTimes(1);
    expect(reports().filter(([, status]) => status === "acknowledged")).toHaveLength(1);
  });

  it("interrupts a scheduled play", async () => {
    await start();
    await act(async () => playback.play(song, "scheduled"));

    await connection.receive(playCommand(40));

    expect(playback.state.playing).toMatchObject({ source: "staff", playbackId: 40 });
  });

  it("reports the earlier clip stopped when another staff play replaces it", async () => {
    await start();
    await connection.receive(playCommand(40));
    await connection.receive(playCommand(41));

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
    await connection.receive(playCommand(40));

    await connection.receive({ type: "stop", playbackId: 40 });

    expect(playback.screen).toBe("alwaysOn");
    expect(reports()).toContainEqual([40, "stopped"]);
  });

  it("ignores a stop for a clip that isn't on screen", async () => {
    await start();
    await connection.receive(playCommand(41));

    await connection.receive({ type: "stop", playbackId: 40 });

    expect(playback.state.playing?.playbackId).toBe(41);
  });

  describe("while a command is being looked up", () => {
    let finishLookup: (item: MediaItem | null) => void;
    beforeEach(() => {
      jest.mocked(fetchModuleItem).mockImplementation(
        () =>
          new Promise((resolve) => {
            finishLookup = resolve;
          }),
      );
    });

    it("cancels it on a stop, reports it stopped and never plays it", async () => {
      await start();
      await connection.receive(playCommand(40));

      await connection.receive({ type: "stop", playbackId: 40 });
      await act(async () => finishLookup(video));

      expect(reports()).toEqual([
        [40, "acknowledged"],
        [40, "stopped"],
      ]);
      expect(playback.state.playing).toBeNull();
    });

    it("doesn't play it if the server says it expired", async () => {
      await start();
      await connection.receive(playCommand(40));

      await connection.receive({ type: "status_result", playbackId: 40, status: "expired" });
      await act(async () => finishLookup(video));

      expect(playback.state.playing).toBeNull();
    });

    it("reports it expired if its deadline passes first", async () => {
      await start();
      await connection.receive(playCommand(40));

      setClock(new Date(NOW.getTime() + 31_000));
      await act(async () => finishLookup(video));

      expect(reports()).toEqual([
        [40, "acknowledged"],
        [40, "expired"],
      ]);
      expect(playback.state.playing).toBeNull();
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
