import { DEFAULT_NIGHT_HOURS, isNight, nightHoursFromAwake } from "../nightHours";

describe("isNight", () => {
  it.each([
    ["2026-10-03T19:59:00", false],
    ["2026-10-03T20:00:00", true],
    ["2026-10-03T23:59:00", true],
    ["2026-10-04T00:00:00", true],
    ["2026-10-04T06:59:00", true],
    ["2026-10-04T07:00:00", false],
    ["2026-10-04T12:00:00", false],
  ])("by default, at %s is %s", (time, expected) => {
    expect(isNight(new Date(time))).toBe(expected);
  });

  it.each([
    ["2026-10-03T21:29:00", false],
    ["2026-10-03T21:30:00", true],
    ["2026-10-04T06:44:00", true],
    ["2026-10-04T06:45:00", false],
  ])("from 21:30 to 6:45, at %s is %s", (time, expected) => {
    const hours = { startMinute: 21 * 60 + 30, endMinute: 6 * 60 + 45 };
    expect(isNight(new Date(time), hours)).toBe(expected);
  });

  it.each([
    ["2026-10-03T23:59:00", false],
    ["2026-10-04T00:30:00", true],
    ["2026-10-04T06:00:00", false],
  ])("from 0:30 to 6:00, at %s is %s", (time, expected) => {
    const hours = { startMinute: 30, endMinute: 6 * 60 };
    expect(isNight(new Date(time), hours)).toBe(expected);
  });

  it("is never night when night starts and ends at the same time", () => {
    const hours = { startMinute: 20 * 60, endMinute: 20 * 60 };
    expect(isNight(new Date("2026-10-03T20:00:00"), hours)).toBe(false);
    expect(isNight(new Date("2026-10-04T03:00:00"), hours)).toBe(false);
  });
});

describe("nightHoursFromAwake", () => {
  it("runs night from bed time to wake time", () => {
    expect(nightHoursFromAwake({ wake: "07:30", bed: "21:00" })).toEqual({ startMinute: 21 * 60, endMinute: 7 * 60 + 30 });
  });

  it.each([
    ["a missing wake time", { wake: "", bed: "21:00" }],
    ["a bed time that isn't a time", { wake: "07:00", bed: "9pm" }],
    ["an hour past the day", { wake: "07:00", bed: "25:00" }],
  ])("falls back to the default for %s", (_name, awake) => {
    expect(nightHoursFromAwake(awake)).toBe(DEFAULT_NIGHT_HOURS);
  });
});
