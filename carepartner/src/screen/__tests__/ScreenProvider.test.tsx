import React from "react";
import { act, render } from "@testing-library/react-native";
import { API_URL } from "../../api/config";
import { setClock } from "../../clock/clock";
import { getCarepartnerId } from "../../pairing/carepartnerId";
import { PlaybackProvider } from "../../playback/PlaybackProvider";
import { STALE_AFTER_MS } from "../screen";
import { SCREEN_CHECK_MS, SCREEN_REFRESH_MS, ScreenProvider, useScreen } from "../ScreenProvider";

jest.mock("../../pairing/carepartnerId", () => ({ getCarepartnerId: jest.fn() }));

const screenResponse = (wake = "07:00") => ({
  layout_preset: "balanced",
  awake_hours: { wake, bed: "21:00", is_awake_now: true },
  spots: [{ id: 3, type: "weather", config: {}, photo: null, data: { temp: 48 }, error: null }],
  schedule: [],
});

/** Answers each screen request with the next of `responses`; the last repeats. */
function mockServer(...responses: ({ status: number } | object)[]) {
  let calls = 0;
  const fetchMock = jest.fn(async () => {
    const response = responses[Math.min(calls++, responses.length - 1)];
    if ("status" in response) return { ok: false, status: response.status, json: async () => ({}) };
    return { ok: true, status: 200, json: async () => response };
  });
  globalThis.fetch = fetchMock as unknown as typeof fetch;
  return fetchMock;
}

let current: ReturnType<typeof useScreen>;
function Probe() {
  current = useScreen();
  return null;
}

const flush = () => act(async () => {});
const tick = (ms = SCREEN_CHECK_MS) => act(async () => jest.advanceTimersByTime(ms));

async function start() {
  await render(
    <PlaybackProvider>
      <ScreenProvider>
        <Probe />
      </ScreenProvider>
    </PlaybackProvider>,
  );
  await flush();
  await flush();
}

describe("ScreenProvider", () => {
  const realFetch = globalThis.fetch;

  beforeEach(() => {
    jest.useFakeTimers();
    setClock(new Date("2026-10-12T10:00:00"));
    jest.mocked(getCarepartnerId).mockResolvedValue(7);
    jest.spyOn(console, "warn").mockImplementation(() => {});
  });

  afterEach(() => {
    globalThis.fetch = realFetch;
    setClock(null);
    jest.useRealTimers();
    jest.clearAllMocks();
    jest.mocked(console.warn).mockRestore();
  });

  it("fetches the paired carepartner's screen", async () => {
    const fetchMock = mockServer(screenResponse());
    await start();

    expect(fetchMock).toHaveBeenCalledWith(`${API_URL}/screen?carepartner_id=7`);
    expect(current.screen?.awakeHours.wake).toBe("07:00");
  });

  it("fetches nothing while unpaired", async () => {
    jest.mocked(getCarepartnerId).mockResolvedValue(null);
    const fetchMock = mockServer(screenResponse());
    await start();
    await tick(SCREEN_REFRESH_MS);

    expect(fetchMock).not.toHaveBeenCalled();
    expect(current.screen).toBeNull();
  });

  it("fetches again every 5 minutes", async () => {
    const fetchMock = mockServer(screenResponse("07:00"), screenResponse("08:00"));
    await start();
    await tick(SCREEN_REFRESH_MS - SCREEN_CHECK_MS);

    expect(fetchMock).toHaveBeenCalledTimes(1);

    await tick();

    expect(fetchMock).toHaveBeenCalledTimes(2);
    expect(current.screen?.awakeHours.wake).toBe("08:00");
  });

  it("fetches again on the first check of a new day", async () => {
    setClock(new Date("2026-10-12T23:59:30"));
    const fetchMock = mockServer(screenResponse());
    await start();
    await tick();

    expect(fetchMock).toHaveBeenCalledTimes(2);
  });

  it("fetches again when asked", async () => {
    const fetchMock = mockServer(screenResponse("07:00"), screenResponse("08:00"));
    await start();
    await act(async () => current.refresh());

    expect(fetchMock).toHaveBeenCalledTimes(2);
    expect(current.screen?.awakeHours.wake).toBe("08:00");
  });

  it.each([
    ["fails", { status: 500 }],
    ["is malformed", { spots: "none" }],
  ])("keeps the last good screen when a fetch %s", async (_name, bad) => {
    mockServer(screenResponse(), bad);
    await start();
    await tick(SCREEN_REFRESH_MS);

    expect(current.screen?.awakeHours.wake).toBe("07:00");
    expect(console.warn).toHaveBeenCalled();
  });

  it("hides live data once fetches have failed for hours", async () => {
    mockServer(screenResponse(), { status: 500 });
    await start();
    await tick(STALE_AFTER_MS - SCREEN_CHECK_MS);

    expect(current.screen?.spots[0]?.type).toBe("weather");

    await tick();

    expect(current.screen?.spots).toEqual([null]);
  });
});
