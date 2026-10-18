import {
  LOCKED_WIDGET_TYPES,
  SCREEN_WIDGET_TYPES,
  isLockedWidgetType,
  isScreenWidgetType,
  splitKnownWidgets,
  widgetName,
} from "../screenWidgets";

describe("widgetName", () => {
  it.each([
    ["day_time", "Day and time"],
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

describe("isScreenWidgetType", () => {
  it.each(SCREEN_WIDGET_TYPES)("knows %s", (type) => {
    expect(isScreenWidgetType(type)).toBe(true);
  });

  it.each(["", "Photo", "calendar", "toString", "__proto__"])("does not know %j", (type) => {
    expect(isScreenWidgetType(type)).toBe(false);
  });
});

describe("locked widgets", () => {
  it("locks only the day and time", () => {
    expect(LOCKED_WIDGET_TYPES).toEqual(["day_time"]);
    expect(isLockedWidgetType("day_time")).toBe(true);
    expect(isLockedWidgetType("photo")).toBe(false);
  });
});

describe("splitKnownWidgets", () => {
  it("separates widgets with a known type from the rest, keeping order", () => {
    const widgets = [
      { id: 1, type: "photo" },
      { id: 2, type: "clock" },
      { id: 3, type: "news" },
      { id: 4, type: "map" },
      { id: 5, type: "day_time" },
    ];
    const { known, unknown } = splitKnownWidgets(widgets);
    expect(known.map((w) => w.id)).toEqual([1, 3, 5]);
    expect(unknown.map((w) => w.id)).toEqual([2, 4]);
  });

  it("returns empty lists for no widgets", () => {
    expect(splitKnownWidgets([])).toEqual({ known: [], unknown: [] });
  });
});
