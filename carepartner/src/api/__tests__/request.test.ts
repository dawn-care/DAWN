import { API_URL } from "../config";
import { getJson, postJson, REQUEST_TIMEOUT_MS } from "../request";

describe("getJson", () => {
  const realFetch = globalThis.fetch;

  const respond = (body: unknown, status = 200) => {
    globalThis.fetch = jest.fn(async () => ({ ok: status < 400, status, json: async () => body })) as unknown as typeof fetch;
  };

  afterEach(() => {
    globalThis.fetch = realFetch;
    jest.useRealTimers();
  });

  it("gets the path from the server and returns its JSON", async () => {
    respond({ id: 9 });

    await expect(getJson("/media/9")).resolves.toEqual({ id: 9 });
    expect(jest.mocked(globalThis.fetch).mock.calls[0][0]).toBe(`${API_URL}/media/9`);
  });

  it("posts a JSON body", async () => {
    respond({ paired: true });

    await expect(postJson("/users/device-pairing/status", { carepartner_id: 7 })).resolves.toEqual({ paired: true });
    expect(jest.mocked(globalThis.fetch).mock.calls[0][1]).toMatchObject({
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ carepartner_id: 7 }),
    });
  });

  it("throws on an error status", async () => {
    respond({ code: "NOT_FOUND" }, 404);

    await expect(getJson("/media/9")).rejects.toThrow("Request for /media/9 failed (404)");
  });

  it.each([
    ["by default", undefined, REQUEST_TIMEOUT_MS],
    ["after the given time", 5_000, 5_000],
  ])("gives up on a request that takes too long, %s", async (_name, timeoutMs, waitMs) => {
    jest.useFakeTimers();
    globalThis.fetch = jest.fn(
      (_url: string, init: RequestInit) =>
        new Promise((_resolve, reject) => init.signal!.addEventListener("abort", () => reject(new Error("aborted")))),
    ) as unknown as typeof fetch;

    const result = getJson("/screen", timeoutMs === undefined ? undefined : { timeoutMs });
    jest.advanceTimersByTime(waitMs - 1);
    expect(jest.mocked(globalThis.fetch).mock.calls[0][1]?.signal?.aborted).toBe(false);
    jest.advanceTimersByTime(1);

    await expect(result).rejects.toThrow("aborted");
  });
});
