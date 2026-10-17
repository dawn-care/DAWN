import type { ScheduledPlay } from "../../screen/screen";
import { duePlay, GRACE_MS, playKey, startOf } from "../schedule";

const play = (scheduledItemId: number, time: string, overrides: Partial<ScheduledPlay> = {}): ScheduledPlay => ({
  scheduledItemId,
  date: "2026-10-12",
  time,
  outsideAwakeHours: false,
  item: { kind: "video", url: `/media/${scheduledItemId}/file` },
  ...overrides,
});

const at = (time: string) => new Date(`2026-10-12T${time}:00`);

describe("schedule", () => {
  it("starts a play at its local date and time", () => {
    expect(startOf(play(1, "08:15"))).toEqual(new Date(2026, 9, 12, 8, 15));
  });

  it("tells one day's play of an item from another day's", () => {
    expect(playKey(play(1, "08:00"))).toBe("1@2026-10-12");
    expect(playKey(play(1, "08:00", { date: "2026-10-13" }))).toBe("1@2026-10-13");
  });

  describe("duePlay", () => {
    const schedule = [play(1, "08:00"), play(2, "10:00")];

    it("has nothing due before a play's time", () => {
      expect(duePlay(schedule, at("07:59"), new Set())).toBeNull();
    });

    it("is due from its time until the grace window ends", () => {
      expect(duePlay(schedule, at("08:00"), new Set())?.scheduledItemId).toBe(1);
      expect(duePlay(schedule, at("08:09"), new Set())?.scheduledItemId).toBe(1);
      expect(duePlay(schedule, new Date(at("08:00").getTime() + GRACE_MS), new Set())).toBeNull();
    });

    it("is still due just after midnight for a play late the day before", () => {
      const lateNight = play(3, "23:55");
      expect(duePlay([lateNight], new Date("2026-10-13T00:02:00"), new Set())?.scheduledItemId).toBe(3);
    });

    it("isn't due early for tomorrow's play at the same time", () => {
      const tomorrow = play(1, "08:00", { date: "2026-10-13" });
      expect(duePlay([tomorrow], at("08:05"), new Set())).toBeNull();
    });

    it("isn't due once played", () => {
      expect(duePlay(schedule, at("08:01"), new Set(["1@2026-10-12"]))).toBeNull();
    });

    it("is due again on another day", () => {
      const tomorrow = play(1, "08:00", { date: "2026-10-13" });

      expect(duePlay([tomorrow], new Date("2026-10-13T08:01:00"), new Set(["1@2026-10-12"]))).toBe(tomorrow);
    });

    it("picks the earliest of several due plays", () => {
      const both = [play(2, "08:15"), play(1, "08:10")];

      expect(duePlay(both, at("08:16"), new Set())?.scheduledItemId).toBe(1);
      expect(duePlay(both, at("08:16"), new Set(["1@2026-10-12"]))?.scheduledItemId).toBe(2);
    });

    it("skips plays outside the carepartner's awake hours", () => {
      expect(duePlay([play(1, "08:00", { outsideAwakeHours: true })], at("08:01"), new Set())).toBeNull();
    });
  });
});
