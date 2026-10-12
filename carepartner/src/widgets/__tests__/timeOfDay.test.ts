import { dateLabel, dayName, timeOfDay } from "../timeOfDay";

describe("timeOfDay", () => {
  it.each([
    ["04:59", "evening"],
    ["05:00", "morning"],
    ["11:59", "morning"],
    ["12:00", "afternoon"],
    ["16:59", "afternoon"],
    ["17:00", "evening"],
    ["23:30", "evening"],
  ])("at %s it is %s", (time, expected) => {
    expect(timeOfDay(new Date(`2026-10-05T${time}:00`))).toBe(expected);
  });
});

describe("dayName and dateLabel", () => {
  it("name the day and date", () => {
    const date = new Date("2026-10-05T09:00:00");
    expect(dayName(date)).toBe("Monday");
    expect(dateLabel(date)).toBe("October 5");
  });
});
