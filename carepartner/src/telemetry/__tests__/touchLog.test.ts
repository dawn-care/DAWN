import { postJson } from "../../api/request";
import { createReportQueue, type ReportQueue } from "../../reports/queue";
import { BATCH_LIMIT, createTouchLog, FLUSH_MS, REPEAT_WINDOW_MS, type Touch, type TouchLog } from "../touchLog";

jest.mock("../../api/request", () => ({ ...jest.requireActual("../../api/request"), postJson: jest.fn(async () => ({})) }));

const T0 = Date.parse("2026-10-18T10:00:00Z");
const weather = { type: "widget", widgetType: "weather", widgetId: "12" } as const;
const touch = (ms: number, overrides: Partial<Touch> = {}): Touch => ({
  screen: "always_on",
  target: weather,
  gesture: "tap",
  at: new Date(T0 + ms),
  ...overrides,
});

const sentEvents = () => jest.mocked(postJson).mock.calls.flatMap(([, body]) => (body as { events: unknown[] }).events);

describe("touch log", () => {
  let queue: ReportQueue;
  let log: TouchLog;

  beforeEach(() => {
    jest.useFakeTimers();
    queue = createReportQueue();
    log = createTouchLog(queue, 7);
  });

  afterEach(() => {
    log.stop();
    queue.stop();
    jest.useRealTimers();
    jest.clearAllMocks();
  });

  it("sends touches every minute, with structured fields only", async () => {
    log.record(touch(0));
    await jest.advanceTimersByTimeAsync(FLUSH_MS - 1);
    expect(postJson).not.toHaveBeenCalled();

    await jest.advanceTimersByTimeAsync(1);

    expect(postJson).toHaveBeenCalledWith("/tablet-events/touches", {
      carepartner_id: 7,
      events: [
        {
          screen: "always_on",
          target_type: "widget",
          widget_type: "weather",
          widget_id: "12",
          media_kind: null,
          gesture: "tap",
          count: 1,
          first_at: "2026-10-18T10:00:00.000Z",
          last_at: "2026-10-18T10:00:00.000Z",
        },
      ],
    });
  });

  it("counts quick repeats on the same thing as one event", async () => {
    log.record(touch(0));
    log.record(touch(800));
    log.record(touch(800 + REPEAT_WINDOW_MS));
    log.flush();
    await jest.advanceTimersByTimeAsync(0);

    expect(sentEvents()).toEqual([
      expect.objectContaining({ count: 3, first_at: "2026-10-18T10:00:00.000Z", last_at: "2026-10-18T10:00:02.800Z" }),
    ]);
  });

  it("starts a new event after a pause, or for a different target or gesture", async () => {
    log.record(touch(0));
    log.record(touch(REPEAT_WINDOW_MS + 1));
    log.record(touch(REPEAT_WINDOW_MS + 2, { target: { type: "empty" } }));
    log.record(touch(REPEAT_WINDOW_MS + 3, { target: { type: "empty" }, gesture: "swipe" }));
    log.flush();
    await jest.advanceTimersByTimeAsync(0);

    expect(sentEvents()).toEqual([
      expect.objectContaining({ target_type: "widget", count: 1 }),
      expect.objectContaining({ target_type: "widget", count: 1 }),
      expect.objectContaining({ target_type: "empty", gesture: "tap" }),
      expect.objectContaining({ target_type: "empty", gesture: "swipe" }),
    ]);
  });

  it("names the media on the playing screen, and nothing on the night screen", async () => {
    log.record(touch(0, { screen: "playing", target: { type: "media", mediaKind: "video" } }));
    log.record(touch(10, { screen: "night", target: null }));
    log.flush();
    await jest.advanceTimersByTimeAsync(0);

    expect(sentEvents()).toEqual([
      expect.objectContaining({ screen: "playing", target_type: "media", media_kind: "video", widget_type: null }),
      expect.objectContaining({ screen: "night", target_type: null, media_kind: null }),
    ]);
  });

  it(`sends straight away once ${BATCH_LIMIT} events are waiting`, async () => {
    for (let i = 0; i < BATCH_LIMIT; i++) log.record(touch(i * (REPEAT_WINDOW_MS + 1)));
    await jest.advanceTimersByTimeAsync(0);

    expect(postJson).toHaveBeenCalledTimes(1);
    expect(sentEvents()).toHaveLength(BATCH_LIMIT);
  });

  it("sends nothing when there were no touches", async () => {
    await jest.advanceTimersByTimeAsync(3 * FLUSH_MS);

    expect(postJson).not.toHaveBeenCalled();
  });
});
