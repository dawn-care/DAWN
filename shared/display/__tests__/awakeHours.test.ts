import {
  DEFAULT_AWAKE_HOURS,
  formatClock,
  isAwake,
  isAwakeAt,
  isOnStep,
  isStartAllowed,
  isValidAwakeHours,
  itemsOutsideHours,
  minuteOfDay,
  parseClock,
  snapToStep,
  startTimeOptions,
} from "../awakeHours";

const hours = { wakeMinute: 7 * 60, bedMinute: 21 * 60 };


describe("minuteOfDay", () => {
  it("counts minutes after midnight", () => {
    expect(minuteOfDay(new Date("2026-10-03T00:00:00"))).toBe(0);
    expect(minuteOfDay(new Date("2026-10-03T07:30:00"))).toBe(450);
    expect(minuteOfDay(new Date("2026-10-03T23:59:00"))).toBe(1439);
  });
});

describe("parseClock and formatClock", () => {
  it.each([
    ["07:00", 420],
    ["7:00", 420],
    ["06:30", 390],
    ["21:00:00", 1260],
    ["00:00", 0],
    [" 23:59 ", 1439],
  ])("reads %s as %i", (text, expected) => {
    expect(parseClock(text)).toBe(expected);
  });

  it.each(["", "7", "24:00", "07:60", "7:5", "noon", "07:00pm"])("rejects %j", (text) => {
    expect(parseClock(text)).toBeNull();
  });

  it("writes minutes the way the server expects", () => {
    expect(formatClock(420)).toBe("07:00");
    expect(formatClock(1260)).toBe("21:00");
    expect(formatClock(0)).toBe("00:00");
  });

  it("round trips", () => {
    expect(parseClock(formatClock(1005))).toBe(1005);
  });
});

describe("isAwake", () => {
  it.each([
    ["2026-10-03T07:59:00", false],
    ["2026-10-03T08:00:00", true],
    ["2026-10-03T12:00:00", true],
    ["2026-10-03T21:59:00", true],
    ["2026-10-03T22:00:00", false],
    ["2026-10-04T00:00:00", false],
    ["2026-10-04T03:00:00", false],
  ])("by default, at %s is %s", (time, expected) => {
    expect(isAwake(new Date(time))).toBe(expected);
  });

  it("uses the wake and bed it is given", () => {
    const early = { wakeMinute: 6 * 60 + 30, bedMinute: 19 * 60 };
    expect(isAwake(new Date("2026-10-03T06:30:00"), early)).toBe(true);
    expect(isAwake(new Date("2026-10-03T18:59:00"), early)).toBe(true);
    expect(isAwake(new Date("2026-10-03T19:00:00"), early)).toBe(false);
  });

  it("agrees with isAwakeAt", () => {
    expect(isAwakeAt(480)).toBe(true);
    expect(isAwakeAt(479)).toBe(false);
    expect(isAwakeAt(1319)).toBe(true);
    expect(isAwakeAt(1320)).toBe(false);
  });
});

describe("isValidAwakeHours", () => {
  it("accepts the default", () => {
    expect(isValidAwakeHours(DEFAULT_AWAKE_HOURS)).toBe(true);
  });

  it("accepts any wake before bed within one day", () => {
    expect(isValidAwakeHours({ wakeMinute: 0, bedMinute: 1439 })).toBe(true);
    expect(isValidAwakeHours({ wakeMinute: 5 * 60, bedMinute: 23 * 60 + 30 })).toBe(true);
  });

  it.each([
    [600, 600],
    [900, 600],
    [-1, 1260],
    [420, 1440],
    [420.5, 1260],
    [420, 1260.5],
    [Number.NaN, 1260],
  ])("rejects wake %d with bed %d", (wakeMinute, bedMinute) => {
    expect(isValidAwakeHours({ wakeMinute, bedMinute })).toBe(false);
  });
});

describe("isOnStep and snapToStep", () => {
  it.each([0, 15, 30, 45, 60, 435])("%i is on a step", (minute) => {
    expect(isOnStep(minute)).toBe(true);
  });

  it.each([1, 7, 14, 16, 434, 7.5])("%d is not on a step", (minute) => {
    expect(isOnStep(minute)).toBe(false);
  });

  it.each([
    [420, 420],
    [421, 420],
    [427, 420],
    [428, 435],
    [435, 435],
    [442, 435],
    [443, 450],
  ])("snaps %i to %i", (minute, expected) => {
    expect(snapToStep(minute)).toBe(expected);
  });

  it("rounds halfway up", () => {
    expect(snapToStep(7 * 60 + 7.5)).toBe(7 * 60 + 15);
  });

  it("stays inside the day", () => {
    expect(snapToStep(1439)).toBe(1425);
    expect(snapToStep(-5)).toBe(0);
  });
});

describe("isStartAllowed", () => {
  it("allows a step at wake up", () => {
    expect(isStartAllowed(420, hours)).toBe(true);
  });

  it("allows the last step before bed", () => {
    expect(isStartAllowed(1245, hours)).toBe(true);
  });

  it("rejects bed time itself and later", () => {
    expect(isStartAllowed(1260, hours)).toBe(false);
    expect(isStartAllowed(1275, hours)).toBe(false);
  });

  it("rejects before wake up", () => {
    expect(isStartAllowed(405, hours)).toBe(false);
  });

  it("rejects a time off the step even when awake", () => {
    expect(isStartAllowed(610, hours)).toBe(false);
  });

  it("does not care how long the item runs", () => {
    // A start at 8:45 pm counts even if a 10 minute video runs past bed at 9 pm.
    expect(isStartAllowed(20 * 60 + 45, hours)).toBe(true);
  });
});

describe("startTimeOptions", () => {
  it("runs every 15 minutes from wake up to just before bed", () => {
    const options = startTimeOptions(hours);
    expect(options[0]).toBe(420);
    expect(options[1]).toBe(435);
    expect(options[options.length - 1]).toBe(1245);
    expect(options).toHaveLength((1260 - 420) / 15);
  });

  it("offers only allowed starts", () => {
    expect(startTimeOptions(hours).every((minute) => isStartAllowed(minute, hours))).toBe(true);
  });

  it("starts at the next step when wake is off the step", () => {
    const options = startTimeOptions({ wakeMinute: 6 * 60 + 10, bedMinute: 7 * 60 + 10 });
    expect(options).toEqual([375, 390, 405, 420]);
  });

  it("uses the default hours when none are given", () => {
    expect(startTimeOptions()).toEqual(startTimeOptions(DEFAULT_AWAKE_HOURS));
  });
});

describe("itemsOutsideHours", () => {
  const items = [
    { name: "Good morning", startMinute: 6 * 60 },
    { name: "Hymn", startMinute: 15 * 60 },
    { name: "Goodnight", startMinute: 20 * 60 + 30 },
    { name: "Late", startMinute: 21 * 60 },
  ];

  it("returns the items that start outside, in their original order", () => {
    expect(itemsOutsideHours(items, hours).map((item) => item.name)).toEqual([
      "Good morning",
      "Late",
    ]);
  });

  it("returns the full items", () => {
    expect(itemsOutsideHours(items, hours)[0]).toBe(items[0]);
  });

  it("finds more items when bed moves earlier", () => {
    const earlier = { wakeMinute: 7 * 60, bedMinute: 19 * 60 };
    expect(itemsOutsideHours(items, earlier).map((item) => item.name)).toEqual([
      "Good morning",
      "Goodnight",
      "Late",
    ]);
  });

  it("returns nothing when every item is awake", () => {
    expect(itemsOutsideHours([items[1]], hours)).toEqual([]);
    expect(itemsOutsideHours([], hours)).toEqual([]);
  });
});
