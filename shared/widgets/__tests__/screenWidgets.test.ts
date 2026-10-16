import { widgetName } from "../screenWidgets";

describe("widgetName", () => {
  it.each([
    ["photo", "Photo"],
    ["message", "Message"],
    ["reminder", "Reminder"],
    ["weather", "Weather"],
    ["sports", "Sports"],
    ["news", "News"],
  ] as const)("calls a %s widget %s", (type, name) => {
    expect(widgetName({ type })).toBe(name);
  });

  it("uses the team's name for a sports widget with a team", () => {
    expect(widgetName({ type: "sports", teamName: "Buffalo Bills" })).toBe("Buffalo Bills");
  });

  it("ignores a team name on any other widget", () => {
    expect(widgetName({ type: "news", teamName: "Buffalo Bills" })).toBe("News");
  });

  it("falls back to Sports when the team name is empty", () => {
    expect(widgetName({ type: "sports", teamName: "" })).toBe("Sports");
  });
});
