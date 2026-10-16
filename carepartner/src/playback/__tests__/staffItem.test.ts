import { API_URL } from "../../api/config";
import { fetchMediaItem, MEDIA_LOOKUP_TIMEOUT_MS } from "../staffItem";

describe("fetchMediaItem", () => {
  const realFetch = globalThis.fetch;
  let warn: jest.SpyInstance;

  const respond = (body: unknown, ok = true) => {
    globalThis.fetch = jest.fn(async () => ({ ok, status: ok ? 200 : 500, json: async () => body })) as unknown as typeof fetch;
  };

  beforeEach(() => {
    warn = jest.spyOn(console, "warn").mockImplementation(() => {});
  });

  afterEach(() => {
    globalThis.fetch = realFetch;
    warn.mockRestore();
    jest.useRealTimers();
  });

  it("looks up the media and returns it with its file", async () => {
    respond({ id: 9, carepartner_id: 2, kind: "video", display_name: "Birthday", url: "/media/9/file" });

    await expect(fetchMediaItem(9)).resolves.toEqual({ kind: "video", url: "/media/9/file", title: "Birthday" });
    expect(jest.mocked(globalThis.fetch).mock.calls[0][0]).toBe(`${API_URL}/media/9`);
  });

  it.each([
    ["an error response", () => respond({ code: "NOT_FOUND" }, false)],
    ["no file", () => respond({ id: 9, kind: "video" })],
    ["no kind", () => respond({ id: 9, url: "/media/9/file" })],
    ["a kind the tablet can't play", () => respond({ id: 9, kind: "document", url: "/media/9/file" })],
    ["null", () => respond(null)],
    [
      "a network failure",
      () => {
        globalThis.fetch = jest.fn(async () => {
          throw new Error("offline");
        }) as unknown as typeof fetch;
      },
    ],
  ])("returns null with a warning for %s", async (_name, setUp) => {
    setUp();

    await expect(fetchMediaItem(9)).resolves.toBeNull();
    expect(warn).toHaveBeenCalled();
  });

  it("gives up on a lookup that takes too long", async () => {
    jest.useFakeTimers();
    globalThis.fetch = jest.fn(
      (_url: string, init: RequestInit) =>
        new Promise((_resolve, reject) => init.signal!.addEventListener("abort", () => reject(new Error("aborted")))),
    ) as unknown as typeof fetch;

    const result = fetchMediaItem(9);
    jest.advanceTimersByTime(MEDIA_LOOKUP_TIMEOUT_MS);

    await expect(result).resolves.toBeNull();
  });
});
