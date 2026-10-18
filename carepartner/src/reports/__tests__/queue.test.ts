import { HttpError } from "../../api/request";
import { createReportQueue, QUEUE_LIMIT, RETRY_FIRST_MS, RETRY_MAX_MS, type ReportQueue } from "../queue";

/** A report that records its name when sent, or throws what `fail` says for each try. */
function report(name: string, sent: string[], group = name, fail: (attempt: number) => unknown = () => null) {
  let attempt = 0;
  return {
    group,
    send: async () => {
      const error = fail(++attempt);
      if (error) throw error;
      sent.push(name);
    },
  };
}

const settle = (ms = 0) => jest.advanceTimersByTimeAsync(ms);
const offline = new Error("Network request failed");

describe("report queue", () => {
  let queue: ReportQueue;
  let sent: string[];

  beforeEach(() => {
    jest.useFakeTimers();
    jest.spyOn(console, "warn").mockImplementation(() => {});
    queue = createReportQueue();
    sent = [];
  });

  afterEach(() => {
    queue.stop();
    jest.useRealTimers();
    jest.restoreAllMocks();
  });

  it("sends reports one at a time, in the order they were made", async () => {
    queue.add(report("a", sent));
    queue.add(report("b", sent));
    queue.add(report("c", sent));
    await settle();

    expect(sent).toEqual(["a", "b", "c"]);
    expect(queue.size()).toBe(0);
  });

  it("keeps a report that didn't get through, and everything after it, until the connection returns", async () => {
    queue.add(report("a", sent, "a", (attempt) => (attempt < 3 ? offline : null)));
    queue.add(report("b", sent));
    await settle();
    expect(sent).toEqual([]);
    expect(queue.size()).toBe(2);

    await settle(RETRY_FIRST_MS);
    expect(sent).toEqual([]);

    await settle(2 * RETRY_FIRST_MS);
    expect(sent).toEqual(["a", "b"]);
  });

  it("treats a server error like a dropped connection", async () => {
    queue.add(report("a", sent, "a", (attempt) => (attempt === 1 ? new HttpError("/playbacks", 503) : null)));
    await settle(RETRY_FIRST_MS);

    expect(sent).toEqual(["a"]);
  });

  it("waits longer each time, up to a limit", async () => {
    let tries = 0;
    queue.add({ group: "a", send: async () => { tries++; throw offline; } });
    await settle();
    expect(tries).toBe(1);

    // 2 s, 4 s, 8 s, 16 s, 32 s, then 60 s each time.
    await settle(RETRY_FIRST_MS + 4_000 + 8_000 + 16_000 + 32_000);
    expect(tries).toBe(6);
    await settle(RETRY_MAX_MS - 1);
    expect(tries).toBe(6);
    await settle(1);
    expect(tries).toBe(7);
  });

  it("drops a refused report and the rest of its group, and carries on with the others", async () => {
    queue.add(report("play1-create", sent, "play1", () => new HttpError("/playbacks", 400)));
    queue.add(report("play1-started", sent, "play1"));
    queue.add(report("play2-create", sent, "play2"));
    await settle();

    expect(sent).toEqual(["play2-create"]);
    expect(console.warn).toHaveBeenCalledWith("The server refused a report (400); dropped 2");
  });

  it(`keeps at most ${QUEUE_LIMIT} reports, dropping the oldest`, async () => {
    queue.add({ group: "stuck", send: async () => { throw offline; } });
    await settle();
    for (let i = 0; i < QUEUE_LIMIT; i++) queue.add(report(`r${i}`, sent));

    expect(queue.size()).toBe(QUEUE_LIMIT);
    await settle(RETRY_FIRST_MS);
    expect(sent[0]).toBe("r0");
    expect(sent).toHaveLength(QUEUE_LIMIT);
  });

  it("sends nothing more once stopped", async () => {
    queue.add(report("a", sent, "a", () => offline));
    await settle();
    queue.stop();
    queue.add(report("b", sent));

    await settle(RETRY_MAX_MS);

    expect(sent).toEqual([]);
  });
});
