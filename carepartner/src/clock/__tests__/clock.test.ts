import { everyMinute, now, parseFakeTime, setClock } from "../clock";

describe("clock", () => {
  beforeEach(() => {
    jest.useFakeTimers({ now: new Date("2026-10-03T09:15:00") });
  });

  afterEach(() => {
    setClock(null);
    jest.useRealTimers();
    jest.restoreAllMocks();
  });

  it("reports the real time by default", () => {
    expect(now()).toEqual(new Date("2026-10-03T09:15:00"));
  });

  it("can be started at a set time and runs forward from there", () => {
    setClock(new Date("2026-12-25T07:59:00"));
    expect(now()).toEqual(new Date("2026-12-25T07:59:00"));

    jest.advanceTimersByTime(60_000);
    expect(now()).toEqual(new Date("2026-12-25T08:00:00"));
  });

  it("returns to the real time when cleared", () => {
    setClock(new Date("2026-12-25T07:59:00"));
    jest.advanceTimersByTime(60_000);

    setClock(null);

    expect(now()).toEqual(new Date("2026-10-03T09:16:00"));
  });

  describe("parseFakeTime", () => {
    it("reads a date-time setting", () => {
      expect(parseFakeTime("2026-10-03T21:30:00")).toEqual(new Date("2026-10-03T21:30:00"));
    });

    it("returns null when the setting is unset", () => {
      expect(parseFakeTime(undefined)).toBeNull();
      expect(parseFakeTime("")).toBeNull();
    });

    it("warns and returns null for a setting that isn't a date", () => {
      jest.spyOn(console, "warn").mockImplementation(() => {});

      expect(parseFakeTime("tomorrow-ish")).toBeNull();
      expect(console.warn).toHaveBeenCalledWith(expect.stringContaining("tomorrow-ish"));
    });
  });
});

describe("everyMinute", () => {
  beforeEach(() => jest.useFakeTimers());
  afterEach(() => {
    setClock(null);
    jest.useRealTimers();
  });

  it("ticks as each minute of the app clock starts, until stopped", () => {
    setClock(new Date("2026-10-12T08:59:45"));
    const ticks: string[] = [];
    const stop = everyMinute(() => ticks.push(now().toTimeString().slice(0, 8)));

    jest.advanceTimersByTime(14_999);
    expect(ticks).toEqual([]);
    jest.advanceTimersByTime(1 + 120_000);
    expect(ticks).toEqual(["09:00:00", "09:01:00", "09:02:00"]);

    stop();
    jest.advanceTimersByTime(60_000);
    expect(ticks).toHaveLength(3);
  });

  it("keeps ticking after a tick throws", () => {
    setClock(new Date("2026-10-12T08:59:30"));
    const tick = jest.fn(() => {
      throw new Error("boom");
    });
    everyMinute(tick);

    expect(() => jest.advanceTimersByTime(30_000)).toThrow("boom");
    expect(() => jest.advanceTimersByTime(60_000)).toThrow("boom");
    expect(tick).toHaveBeenCalledTimes(2);
  });
});
