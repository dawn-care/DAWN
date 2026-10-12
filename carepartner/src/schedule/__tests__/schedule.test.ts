import { dayOf, minuteOf, nextEntry, readEntries, routineStartsAt, timedEntriesDueAt, type ScheduleEntry } from "../schedule";

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
const timed = (id: number, time: string, overrides: Partial<ScheduleEntry> = {}) =>
  entry(id, { timed: true, time, ...overrides });

const none = new Set<number>();

describe("readEntries", () => {
  it("keeps well-formed entries and drops malformed ones", () => {
    const good = entry(1);
    expect(readEntries([good, null, "x", { id: "2", module_type: "video", position: 1 }, { id: 3 }])).toEqual([good]);
  });

  it("treats anything but a list as no entries", () => {
    expect(readEntries({ error: "nope" })).toEqual([]);
    expect(readEntries(null)).toEqual([]);
  });
});

describe("minuteOf and dayOf", () => {
  it("formats the local minute and day", () => {
    const time = new Date("2026-03-04T07:05:59");
    expect(minuteOf(time)).toBe("07:05");
    expect(dayOf(time)).toBe("2026-03-04");
  });
});

describe("timedEntriesDueAt", () => {
  const entries = [timed(1, "08:00:00"), timed(2, "08:00:00"), timed(3, "09:00:00"), entry(4)];

  it("finds the timed entries at this minute", () => {
    expect(timedEntriesDueAt(entries, "08:00", none)).toEqual([1, 2]);
  });

  it("leaves out entries that have played", () => {
    expect(timedEntriesDueAt(entries, "08:00", new Set([1]))).toEqual([2]);
  });

  it("finds nothing at other minutes", () => {
    expect(timedEntriesDueAt(entries, "08:01", none)).toEqual([]);
  });
});

describe("routineStartsAt", () => {
  const routine = [entry(1), entry(2)];

  it("starts at the auto-start minute when something in the routine hasn't played", () => {
    expect(routineStartsAt("08:00:00", "08:00", routine, new Set([1]))).toBe(true);
  });

  it("doesn't start at other minutes", () => {
    expect(routineStartsAt("08:00:00", "08:01", routine, none)).toBe(false);
  });

  it("doesn't start when the whole routine has played", () => {
    expect(routineStartsAt("08:00:00", "08:00", routine, new Set([1, 2]))).toBe(false);
  });

  it("doesn't start when nothing in it can play", () => {
    expect(routineStartsAt("08:00:00", "08:00", [entry(1, { module_type: "weather" })], none)).toBe(false);
  });

  it("doesn't start without an auto-start time", () => {
    expect(routineStartsAt(undefined, "08:00", routine, none)).toBe(false);
  });
});

describe("nextEntry", () => {
  it("plays nothing when nothing is due and the routine isn't running", () => {
    expect(nextEntry([entry(1), timed(2, "08:00:00")], { played: none, due: none, routineRunning: false })).toBeNull();
  });

  it("plays the routine in position order", () => {
    const entries = [entry(1, { position: 3 }), entry(2, { position: 1 }), entry(3, { position: 2 })];
    expect(nextEntry(entries, { played: none, due: none, routineRunning: true })?.id).toBe(2);
    expect(nextEntry(entries, { played: new Set([2]), due: none, routineRunning: true })?.id).toBe(3);
  });

  it("plays due timed entries before the routine, earliest first", () => {
    const entries = [entry(1), timed(2, "09:00:00"), timed(3, "08:00:00")];
    expect(nextEntry(entries, { played: none, due: new Set([2, 3]), routineRunning: true })?.id).toBe(3);
  });

  it("doesn't play timed entries that aren't due yet", () => {
    expect(nextEntry([timed(1, "09:00:00")], { played: none, due: none, routineRunning: true })).toBeNull();
  });

  it("skips entries that have played", () => {
    expect(nextEntry([timed(1, "08:00:00")], { played: new Set([1]), due: new Set([1]), routineRunning: false })).toBeNull();
  });

  it("skips entries the tablet can't play", () => {
    const entries = [entry(1, { module_type: "weather" }), entry(2, { module_type: "messages" }), entry(3, { metadata: null }), entry(4)];
    expect(nextEntry(entries, { played: none, due: none, routineRunning: true })?.id).toBe(4);
  });

  it("finishes the routine once everything playable has played", () => {
    const entries = [entry(1), entry(2, { module_type: "news" })];
    expect(nextEntry(entries, { played: new Set([1]), due: none, routineRunning: true })).toBeNull();
  });
});
