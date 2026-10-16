import { GRID } from "../grid";
import {
  WIDGET_SIZES,
  areasOverlap,
  canPlace,
  hasRoom,
  isValidPlacement,
  openCells,
  widgetArea,
} from "../gridPlacement";
import type { PlacedWidget } from "../gridPlacement";
import type { ScreenWidgetType } from "../screenWidgets";

const types = Object.keys(WIDGET_SIZES) as ScreenWidgetType[];

const photo: PlacedWidget = { id: 1, type: "photo", col: 0, row: 0 };
const weather: PlacedWidget = { id: 2, type: "weather", col: 2, row: 0 };

describe("WIDGET_SIZES", () => {
  it.each(types)("%s is at least one cell and no bigger than the grid", (type) => {
    const { width, height } = WIDGET_SIZES[type];
    expect(Number.isInteger(width) && Number.isInteger(height)).toBe(true);
    expect(width).toBeGreaterThanOrEqual(1);
    expect(height).toBeGreaterThanOrEqual(1);
    expect(width).toBeLessThanOrEqual(GRID.columns);
    expect(height).toBeLessThanOrEqual(GRID.rows);
  });
});

describe("widgetArea", () => {
  it("covers the widget's size starting at its cell", () => {
    expect(widgetArea({ type: "message", col: 1, row: 2 })).toEqual({ col: 1, row: 2, width: 2, height: 1 });
  });
});

describe("areasOverlap", () => {
  const a = { col: 0, row: 0, width: 2, height: 2 };

  it("is true when areas share a cell", () => {
    expect(areasOverlap(a, { col: 1, row: 1, width: 2, height: 2 })).toBe(true);
    expect(areasOverlap(a, a)).toBe(true);
  });

  it("is false for areas that only touch", () => {
    expect(areasOverlap(a, { col: 2, row: 0, width: 1, height: 1 })).toBe(false);
    expect(areasOverlap(a, { col: 0, row: 2, width: 1, height: 1 })).toBe(false);
  });
});

describe("canPlace", () => {
  it("allows any cell on an empty grid where the widget stays on the grid", () => {
    expect(canPlace([], "weather", { col: 3, row: 2 })).toBe(true);
    expect(canPlace([], "photo", { col: 2, row: 0 })).toBe(true);
  });

  it("refuses a widget that would run off the grid", () => {
    expect(canPlace([], "message", { col: 3, row: 0 })).toBe(false);
    expect(canPlace([], "photo", { col: 0, row: 1 })).toBe(false);
  });

  it("refuses a cell that is not on the grid", () => {
    expect(canPlace([], "weather", { col: -1, row: 0 })).toBe(false);
    expect(canPlace([], "weather", { col: 4, row: 0 })).toBe(false);
    expect(canPlace([], "weather", { col: 0.5, row: 0 })).toBe(false);
  });

  it("refuses a place another widget covers, including its far cells", () => {
    expect(canPlace([photo], "weather", { col: 0, row: 0 })).toBe(false);
    expect(canPlace([photo], "weather", { col: 1, row: 2 })).toBe(false);
    expect(canPlace([photo], "weather", { col: 2, row: 2 })).toBe(true);
  });

  it("refuses a widget whose body would reach a neighbour", () => {
    expect(canPlace([weather], "message", { col: 1, row: 0 })).toBe(false);
    expect(canPlace([weather], "message", { col: 0, row: 0 })).toBe(true);
  });

  it("lets a widget being moved overlap its old place", () => {
    expect(canPlace([photo], "photo", { col: 1, row: 0 })).toBe(false);
    expect(canPlace([photo], "photo", { col: 1, row: 0 }, photo.id)).toBe(true);
  });
});

describe("openCells", () => {
  it("lists every cell for a one cell widget on an empty grid", () => {
    expect(openCells([], "weather")).toHaveLength(GRID.columns * GRID.rows);
  });

  it("lists only corners where the widget fits, in reading order", () => {
    expect(openCells([], "photo")).toEqual([
      { col: 0, row: 0 },
      { col: 1, row: 0 },
      { col: 2, row: 0 },
    ]);
  });

  it("leaves out cells under an existing widget", () => {
    const cells = openCells([photo], "weather");
    expect(cells).toHaveLength(GRID.columns * GRID.rows - 6);
    expect(cells).not.toContainEqual({ col: 1, row: 1 });
    expect(cells).toContainEqual({ col: 2, row: 0 });
  });

  it("is empty when nothing of that size fits", () => {
    const full: PlacedWidget[] = [
      photo,
      { id: 3, type: "photo", col: 2, row: 0 },
    ];
    expect(openCells(full, "message")).toEqual([]);
    expect(openCells(full, "weather")).toEqual([]);
  });
});

describe("hasRoom", () => {
  it("is true when a cell is open and false when the grid is full", () => {
    expect(hasRoom([], "photo")).toBe(true);
    const full: PlacedWidget[] = [
      { id: 1, type: "photo", col: 0, row: 0 },
      { id: 2, type: "photo", col: 2, row: 0 },
    ];
    expect(hasRoom(full, "weather")).toBe(false);
  });
});

describe("isValidPlacement", () => {
  it("accepts an empty list and widgets that fit apart", () => {
    expect(isValidPlacement([])).toBe(true);
    expect(isValidPlacement([photo, weather])).toBe(true);
  });

  it("rejects overlapping widgets", () => {
    expect(isValidPlacement([photo, { id: 9, type: "sports", col: 1, row: 1 }])).toBe(false);
  });

  it("rejects a widget off the grid", () => {
    expect(isValidPlacement([{ id: 9, type: "message", col: 3, row: 0 }])).toBe(false);
  });
});
