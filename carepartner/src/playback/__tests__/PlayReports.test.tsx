import React from "react";
import { act, render } from "@testing-library/react-native";
import { HttpError, patchJson, postJson } from "../../api/request";
import { setClock } from "../../clock/clock";
import { readPairedId } from "../../pairing/pairingState";
import { RETRY_FIRST_MS } from "../../reports/queue";
import type { MediaItem } from "../media";
import { PlaybackProvider, usePlayback } from "../PlaybackProvider";
import PlayReports, { REPORT_DEADLINE_MS } from "../PlayReports";

jest.mock("../../pairing/pairingState", () => ({ readPairedId: jest.fn() }));
jest.mock("../../api/request", () => ({
  ...jest.requireActual("../../api/request"),
  postJson: jest.fn(),
  patchJson: jest.fn(),
}));

const hymn: MediaItem = { id: 9, kind: "audio", url: "/media/9/file", title: "Morning hymn" };

let playback: ReturnType<typeof usePlayback>;
function Probe() {
  playback = usePlayback();
  return null;
}

const settle = (ms = 0) => act(() => jest.advanceTimersByTimeAsync(ms));
const playId = () => playback.state.playing!.playId;

/** Everything sent, in order: "create" for a new record, or the status. */
const sent = () =>
  [...jest.mocked(postJson).mock.calls.map((call) => ({ call, kind: "create" })), ...jest.mocked(patchJson).mock.calls.map((call) => ({ call, kind: "status" }))]
    .sort((a, b) => order.indexOf(a.call) - order.indexOf(b.call))
    .map(({ call, kind }) => (kind === "create" ? "create" : (call[1] as { status: string }).status));
let order: unknown[][];

async function start() {
  await render(
    <PlaybackProvider>
      <PlayReports />
      <Probe />
    </PlaybackProvider>,
  );
  await settle();
}

describe("PlayReports", () => {
  beforeEach(() => {
    jest.useFakeTimers({ now: new Date("2026-10-18T15:00:00Z") });
    // The app clock is hours off, as with EXPO_PUBLIC_FAKE_TIME.
    setClock(new Date("2026-10-18T08:00:00Z"));
    jest.spyOn(console, "warn").mockImplementation(() => {});
    jest.mocked(readPairedId).mockResolvedValue(7);
    order = [];
    jest.mocked(postJson).mockImplementation(async (...call) => {
      order.push(call);
      return { id: 50 };
    });
    jest.mocked(patchJson).mockImplementation(async (...call) => {
      order.push(call);
      return {};
    });
  });

  afterEach(() => {
    setClock(null);
    jest.useRealTimers();
    jest.restoreAllMocks();
    jest.clearAllMocks();
  });

  it("records a scheduled play, then reports it acknowledged, started and finished", async () => {
    await start();

    await act(async () => playback.play(hymn, "scheduled", { scheduledItemId: 4 }));
    await settle();
    expect(postJson).toHaveBeenCalledWith("/playbacks", {
      carepartner_id: 7,
      media_id: 9,
      scheduled_item_id: 4,
      source: "scheduled",
      expires_at: new Date(Date.parse("2026-10-18T15:00:00Z") + REPORT_DEADLINE_MS).toISOString(),
    });
    expect(sent()).toEqual(["create", "acknowledged"]);

    await act(async () => playback.started(playId()));
    await act(async () => playback.ended(playId(), "finished"));
    await settle();

    expect(sent()).toEqual(["create", "acknowledged", "started", "finished"]);
    expect(patchJson).toHaveBeenLastCalledWith("/playbacks/50/status", { carepartner_id: 7, status: "finished" });
  });

  it("sends only ids and statuses, never the media's title", async () => {
    await start();
    await act(async () => playback.play(hymn, "scheduled", { scheduledItemId: 4 }));
    await settle();

    expect(JSON.stringify([jest.mocked(postJson).mock.calls, jest.mocked(patchJson).mock.calls])).not.toContain("hymn");
  });

  it("reports a clip that never loads as failed, without started", async () => {
    await start();
    await act(async () => playback.play(hymn, "scheduled", { scheduledItemId: 4 }));
    await act(async () => playback.ended(playId(), "failed"));
    await settle();

    expect(sent()).toEqual(["create", "acknowledged", "failed"]);
  });

  it("reports a play replaced by a staff clip as stopped", async () => {
    await start();
    await act(async () => playback.play(hymn, "scheduled", { scheduledItemId: 4 }));
    await act(async () => playback.play({ id: 3, kind: "video", url: "/media/3/file" }, "staff", { playbackId: 80 }));
    await settle();

    expect(sent()).toEqual(["create", "acknowledged", "stopped"]);
  });

  it("leaves staff plays to the socket", async () => {
    await start();
    await act(async () => playback.play({ id: 3, kind: "video", url: "/media/3/file" }, "staff", { playbackId: 80 }));
    await settle();

    expect(postJson).not.toHaveBeenCalled();
    expect(patchJson).not.toHaveBeenCalled();
  });

  it("holds reports while offline and sends them, in order, once the server is back", async () => {
    jest.mocked(postJson).mockRejectedValueOnce(new Error("Network request failed"));
    await start();
    await act(async () => playback.play(hymn, "scheduled", { scheduledItemId: 4 }));
    await act(async () => playback.started(playId()));
    await act(async () => playback.ended(playId(), "finished"));
    await settle();
    expect(patchJson).not.toHaveBeenCalled();

    await settle(RETRY_FIRST_MS);

    // The first try failed; the retry created the record, then the statuses followed.
    expect(postJson).toHaveBeenCalledTimes(2);
    expect(jest.mocked(patchJson).mock.calls.map(([, body]) => (body as { status: string }).status)).toEqual([
      "acknowledged",
      "started",
      "finished",
    ]);
  });

  it("sends no statuses for a play the server refused to record", async () => {
    jest.mocked(postJson).mockRejectedValueOnce(new HttpError("/playbacks", 400));
    await start();
    await act(async () => playback.play(hymn, "scheduled", { scheduledItemId: 4 }));
    await settle();
    await act(async () => playback.ended(playId(), "finished"));
    await settle();

    expect(patchJson).not.toHaveBeenCalled();
  });
});
