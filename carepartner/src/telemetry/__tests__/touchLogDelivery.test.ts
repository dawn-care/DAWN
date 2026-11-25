import { createReportQueue, type ReportQueue } from "../../reports/queue";
import { createTouchLog, FLUSH_MS, type TouchLog } from "../touchLog";

/*
 * The real request code against a server that answers as POST /tablet-events/touches
 * does: 204 with no body. touchLog.test.ts mocks postJson, so it can't see how the
 * answer is read.
 */

const T0 = Date.parse("2026-10-18T10:00:00Z");
const tap = (ms: number) => ({
  screen: "always_on" as const,
  target: { type: "widget" as const, widgetType: "weather", widgetId: "12" },
  gesture: "tap" as const,
  at: new Date(T0 + ms),
});

describe("touch log delivery", () => {
  const realFetch = globalThis.fetch;
  let queue: ReportQueue;
  let log: TouchLog;

  beforeEach(() => {
    jest.useFakeTimers();
    globalThis.fetch = jest.fn(async () => ({
      ok: true,
      status: 204,
      json: async () => {
        throw new SyntaxError("Unexpected end of JSON input");
      },
    })) as unknown as typeof fetch;
    queue = createReportQueue();
    log = createTouchLog(queue, 7);
  });

  afterEach(() => {
    log.stop();
    queue.stop();
    globalThis.fetch = realFetch;
    jest.useRealTimers();
  });

  const sentBatches = () =>
    jest.mocked(globalThis.fetch).mock.calls.map(([, init]) => JSON.parse(String(init?.body)).events.length);

  it("sends each new batch after the server answers 204, and never resends one", async () => {
    log.record(tap(0));
    await jest.advanceTimersByTimeAsync(FLUSH_MS);
    log.record(tap(FLUSH_MS + 10_000));
    log.record(tap(FLUSH_MS + 20_000));
    await jest.advanceTimersByTimeAsync(FLUSH_MS);

    // One batch of 1, then one of 2: no retries of the first
    expect(sentBatches()).toEqual([1, 2]);
  });
});
