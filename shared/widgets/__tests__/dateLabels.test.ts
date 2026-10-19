import { ordinal, parseDate, timeLabel } from "../dateLabels";

// In a US time zone, reading "2026-10-03" as UTC would give October 2 in the evening, so
// these only catch that mistake when the tests run in such a zone
describe("parseDate", () => {
  it.each([
    ["2026-10-03", 2026, 9, 3],
    ["2026-01-05", 2026, 0, 5],
    ["2026-12-31", 2026, 11, 31],
  ])("reads %s as a local date, not UTC", (text, year, month, day) => {
    const date = parseDate(text);
    expect(date.getFullYear()).toBe(year);
    expect(date.getMonth()).toBe(month);
    expect(date.getDate()).toBe(day);
    expect(date.getHours()).toBe(0);
  });
});

describe("timeLabel", () => {
  it.each([
    ["00:00", "12 AM"],
    ["00:30", "12:30 AM"],
    ["08:00", "8 AM"],
    ["08:05", "8:05 AM"],
    ["11:45", "11:45 AM"],
    ["12:00", "12 PM"],
    ["12:15", "12:15 PM"],
    ["18:30", "6:30 PM"],
    ["23:59", "11:59 PM"],
  ])("%s is %s", (time, expected) => {
    expect(timeLabel(time)).toBe(expected);
  });

  it("returns text that isn't a time as it is", () => {
    expect(timeLabel("noon")).toBe("noon");
  });
});

describe("ordinal", () => {
  it.each([
    [1, "1st"],
    [2, "2nd"],
    [3, "3rd"],
    [4, "4th"],
    [11, "11th"],
    [12, "12th"],
    [13, "13th"],
    [21, "21st"],
    [22, "22nd"],
    [23, "23rd"],
    [30, "30th"],
    [31, "31st"],
  ])("%i is %s", (n, expected) => {
    expect(ordinal(n)).toBe(expected);
  });
});
