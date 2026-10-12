import React from "react";
import { act, render } from "@testing-library/react-native";
import { setClock } from "../../clock/clock";
import { getCarepartnerId } from "../../pairing/carepartnerId";
import { PlaybackProvider, usePlayback } from "../../playback/PlaybackProvider";
import type { ScheduleEntry } from "../schedule";
import { SCHEDULE_CHECK_MS, ScheduleProvider } from "../ScheduleProvider";

jest.mock("../../pairing/carepartnerId", () => ({ getCarepartnerId: jest.fn() }));

const entry = (id: number, overrides: Partial<ScheduleEntry> = {}): ScheduleEntry => ({
  id,
  module_type: "video",
  metadata: { id },
  duration: 30,
  timed: false,
  time: null,
  position: id,
  ...overrides,
});

/** Today's schedule: a routine of two playable items around a weather card, and one timed song. */
const schedule = [
  entry(1, { position: 1 }),
  entry(2, { position: 3, module_type: "photo" }),
  entry(3, { position: 2, module_type: "weather", metadata: { ID: 3 } }),
  entry(4, { module_type: "audio", timed: true, time: "08:05:00" }),
];

let playback: ReturnType<typeof usePlayback>;
function Probe() {
  playback = usePlayback();
  return null;
}

/** Answers the carepartner and schedule requests. */
function mockBackend(entries: ScheduleEntry[] = schedule) {
  const fetchMock = jest.fn(async (url: string) => {
    const body = url.includes("/configurations/day") ? entries : { id: 7, name: "Ada", auto_start: "08:00:00" };
    return { ok: true, status: 200, json: async () => body };
  });
  globalThis.fetch = fetchMock as unknown as typeof fetch;
  return fetchMock;
}

const flush = () => act(async () => {});
const tick = (ms = SCHEDULE_CHECK_MS) => act(async () => jest.advanceTimersByTime(ms));
const playingId = () => (playback.state.playing?.item as ScheduleEntry | undefined)?.id ?? null;
const endPlay = () => act(async () => playback.ended(playback.state.playing!.playId));

async function start(time: string) {
  setClock(new Date(time));
  await render(
    <PlaybackProvider>
      <ScheduleProvider>
        <Probe />
      </ScheduleProvider>
    </PlaybackProvider>,
  );
  await flush();
  await flush();
}

describe("ScheduleProvider", () => {
  const realFetch = globalThis.fetch;

  beforeEach(() => {
    jest.useFakeTimers();
    jest.mocked(getCarepartnerId).mockResolvedValue(7);
  });

  afterEach(() => {
    globalThis.fetch = realFetch;
    setClock(null);
    jest.useRealTimers();
    jest.clearAllMocks();
  });

  it("plays nothing before auto-start", async () => {
    mockBackend();
    await start("2026-10-12T07:58:00");
    await tick();

    expect(playingId()).toBeNull();
  });

  it("plays the routine in order from auto-start, skipping what the tablet can't play", async () => {
    mockBackend();
    await start("2026-10-12T07:59:50");

    await tick();
    expect(playingId()).toBe(1);

    await endPlay();
    expect(playingId()).toBe(2);

    await endPlay();
    expect(playingId()).toBeNull();
    expect(playback.screen).toBe("alwaysOn");
  });

  it("plays a timed entry at its time", async () => {
    mockBackend([entry(4, { module_type: "audio", timed: true, time: "08:05:00" })]);
    await start("2026-10-12T08:04:50");
    expect(playingId()).toBeNull();

    await tick();

    expect(playingId()).toBe(4);
  });

  it("holds a timed entry that comes due during another play, and plays it next", async () => {
    mockBackend();
    await start("2026-10-12T08:00:00");
    expect(playingId()).toBe(1);

    setClock(new Date("2026-10-12T08:05:00"));
    await tick();
    expect(playingId()).toBe(1);

    await endPlay();
    expect(playingId()).toBe(4);

    await endPlay();
    expect(playingId()).toBe(2);
  });

  it("doesn't play an entry again once it has played", async () => {
    mockBackend([entry(4, { module_type: "audio", timed: true, time: "08:05:00" })]);
    await start("2026-10-12T08:05:00");
    await endPlay();

    await tick();
    await tick();

    expect(playingId()).toBeNull();
  });

  it("forgets what played on a new day and fetches the new day's schedule", async () => {
    const fetchMock = mockBackend([entry(4, { module_type: "audio", timed: true, time: "08:05:00" })]);
    await start("2026-10-12T08:05:00");
    await endPlay();
    const requestsBefore = fetchMock.mock.calls.length;

    setClock(new Date("2026-10-13T08:04:50"));
    await tick();
    await flush();
    expect(fetchMock.mock.calls.length).toBeGreaterThan(requestsBefore);

    await tick();
    expect(playingId()).toBe(4);
  });

  it("fetches nothing until the tablet is paired", async () => {
    jest.mocked(getCarepartnerId).mockResolvedValue(null);
    const fetchMock = mockBackend();
    await start("2026-10-12T08:00:00");

    expect(fetchMock).not.toHaveBeenCalled();
  });
});
