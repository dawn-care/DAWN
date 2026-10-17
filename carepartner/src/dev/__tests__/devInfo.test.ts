import { API_URL } from "../../api/config";
import type { ScheduledPlay } from "../../screen/screen";
import { fetchConnected, playStatus, readMediaList, startStaffPlay } from "../devInfo";

const play = (overrides: Partial<ScheduledPlay> = {}): ScheduledPlay => ({
  scheduledItemId: 1,
  date: "2026-10-12",
  time: "10:00",
  outsideAwakeHours: false,
  item: { kind: "video", url: "/media/1/file" },
  ...overrides,
});
const at = (time: string) => new Date(`2026-10-12T${time}:00`);

describe("playStatus", () => {
  it.each([
    ["09:59", "upcoming"],
    ["10:00", "due"],
    ["10:09", "due"],
    ["10:10", "missed"],
  ])("at %s is %s", (time, status) => {
    expect(playStatus(play(), at(time), new Set())).toBe(status);
  });

  it("is played once started, whatever the time", () => {
    expect(playStatus(play(), at("10:05"), new Set(["1@2026-10-12"]))).toBe("played");
  });

  it("is asleep outside the awake hours", () => {
    expect(playStatus(play({ outsideAwakeHours: true }), at("10:05"), new Set())).toBe("asleep");
  });
});

describe("readMediaList", () => {
  it("keeps the id, kind and name, and drops entries without them", () => {
    expect(
      readMediaList([
        { id: 1, kind: "video", display_name: "Birthday", url: "/media/1/file" },
        { kind: "audio" },
        { id: 3, kind: "photo" },
      ]),
    ).toEqual([
      { id: 1, kind: "video", name: "Birthday" },
      { id: 3, kind: "photo", name: "" },
    ]);
  });

  it("reads anything else as no media", () => {
    expect(readMediaList({ error: "nope" })).toEqual([]);
  });
});

describe("server calls", () => {
  const realFetch = globalThis.fetch;
  const respond = (body: unknown) => {
    globalThis.fetch = jest.fn(async () => ({ ok: true, status: 200, json: async () => body })) as unknown as typeof fetch;
  };

  afterEach(() => {
    globalThis.fetch = realFetch;
  });

  it("reads whether the tablet's socket is connected", async () => {
    respond({ carepartner_id: 4, connected: true });

    await expect(fetchConnected(4)).resolves.toBe(true);
    expect(jest.mocked(globalThis.fetch).mock.calls[0][0]).toBe(`${API_URL}/realtime/carepartners/4/connection`);
  });

  it("starts a staff play through the server and returns its playback id", async () => {
    respond({ playback: { id: 31 }, delivered: true });

    await expect(startStaffPlay(4, 9)).resolves.toBe(31);
    expect(jest.mocked(globalThis.fetch).mock.calls[0]).toEqual([
      `${API_URL}/playbacks/play`,
      expect.objectContaining({ method: "POST", body: JSON.stringify({ carepartner_id: 4, media_id: 9 }) }),
    ]);
  });
});
