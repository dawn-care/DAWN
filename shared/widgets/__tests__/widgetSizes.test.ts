import { hasWidgetSize, widgetSize, WIDGET_SIZES } from "../widgetSizes";
import { fitsGrid } from "../grid";

describe("widget sizes", () => {
  it("gives each known type its fixed size", () => {
    expect(widgetSize("photo")).toEqual({ width: 2, height: 2 });
    expect(widgetSize("weather")).toEqual({ width: 2, height: 1 });
  });

  it("knows no size for an unknown type", () => {
    expect(widgetSize("slideshow")).toBeUndefined();
    expect(widgetSize("toString")).toBeUndefined();
  });

  it("fits every size on the grid", () => {
    for (const size of Object.values(WIDGET_SIZES)) {
      expect(fitsGrid({ col: 0, row: 0, ...size })).toBe(true);
    }
  });

  it("accepts a widget at its type's size, wherever it is", () => {
    expect(hasWidgetSize("weather", { col: 2, row: 2, width: 2, height: 1 })).toBe(true);
    expect(hasWidgetSize("news", { col: 0, row: 2, width: 2, height: 1 })).toBe(true);
  });

  it.each([
    ["a one cell weather widget", "weather", { width: 1, height: 1 }],
    ["a turned news widget", "news", { width: 1, height: 2 }],
    ["an unknown type", "slideshow", { width: 1, height: 1 }],
  ])("rejects %s", (_name, type, size) => {
    expect(hasWidgetSize(type, { col: 0, row: 0, ...size })).toBe(false);
  });
});
