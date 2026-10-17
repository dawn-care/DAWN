import { GRID } from "../grid";
import {
  DATE_TIME_AREA,
  FIXED_AREAS,
  WIDGET_SIZES,
  areasOverlap,
  canPlace,
  hasRoom,
  isValidPlacement,
  openCells,
  widgetArea,
} from "../gridPlacement";
import type { PlacedWidget } from "../gridPlacement";
import { splitKnownWidgets } from "../screenWidgets";
import type { ScreenWidgetType } from "../screenWidgets";

const types = Object.keys(WIDGET_SIZES) as ScreenWidgetType[];

/** Cells left for widgets once the date and time row is taken. */
const OPEN_CELLS = GRID.columns * (GRID.rows - DATE_TIME_AREA.height);

const photo: PlacedWidget = { id: 1, type: "photo", col: 0, row: 1 };
const weather: PlacedWidget = { id: 2, type: "weather", col: 2, row: 1 };

describe("WIDGET_SIZES", () => {
  it.each(types)("%s is at least one cell and fits below the date and time", (type) => {
    const { width, height } = WIDGET_SIZES[type];
    expect(Number.isInteger(width) && Number.isInteger(height)).toBe(true);
    expect(width).toBeGreaterThanOrEqual(1);
    expect(height).toBeGreaterThanOrEqual(1);
    expect(width).toBeLessThanOrEqual(GRID.columns);
    expect(height).toBeLessThanOrEqual(GRID.rows - DATE_TIME_AREA.height);
  });

  it("has the agreed sizes", () => {
    expect(WIDGET_SIZES).toEqual({
      photo: { width: 2, height: 2 },
      message: { width: 2, height: 1 },
      reminder: { width: 2, height: 1 },
      news: { width: 2, height: 1 },
      weather: { width: 1, height: 1 },
      sports: { width: 1, height: 1 },
    });
  });
});

describe("DATE_TIME_AREA", () => {
  it("is the whole top row and is the only fixed area", () => {
    expect(DATE_TIME_AREA).toEqual({ col: 0, row: 0, width: GRID.columns, height: 1 });
    expect(FIXED_AREAS).toEqual([DATE_TIME_AREA]);
  });
});

describe("widgetArea", () => {
  it("covers the widget's size starting at its cell", () => {
    expect(widgetArea({ type: "message", col: 1, row: 2 })).toEqual({ col: 1, row: 2, width: 2, height: 1 });
    expect(widgetArea({ type: "photo", col: 2, row: 1 })).toEqual({ col: 2, row: 1, width: 2, height: 2 });
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
  it("allows any cell below the date and time where the widget stays on the grid", () => {
    expect(canPlace([], "weather", { col: 3, row: 2 })).toBe(true);
    expect(canPlace([], "photo", { col: 2, row: 1 })).toBe(true);
  });

  it("refuses any cell in the date and time row", () => {
    for (let col = 0; col < GRID.columns; col += 1) {
      expect(canPlace([], "weather", { col, row: 0 })).toBe(false);
    }
    expect(canPlace([], "message", { col: 0, row: 0 })).toBe(false);
  });

  it("refuses a widget that would run off the grid", () => {
    expect(canPlace([], "message", { col: 3, row: 1 })).toBe(false);
    expect(canPlace([], "photo", { col: 0, row: 2 })).toBe(false);
  });

  it("refuses a cell that is not on the grid", () => {
    expect(canPlace([], "weather", { col: -1, row: 1 })).toBe(false);
    expect(canPlace([], "weather", { col: 4, row: 1 })).toBe(false);
    expect(canPlace([], "weather", { col: 0.5, row: 1 })).toBe(false);
  });

  it("refuses a place another widget covers, including its far cells", () => {
    expect(canPlace([photo], "weather", { col: 0, row: 1 })).toBe(false);
    expect(canPlace([photo], "weather", { col: 1, row: 2 })).toBe(false);
    expect(canPlace([photo], "weather", { col: 2, row: 2 })).toBe(true);
  });

  it("refuses a widget whose body would reach a neighbour", () => {
    expect(canPlace([weather], "message", { col: 1, row: 1 })).toBe(false);
    expect(canPlace([weather], "message", { col: 0, row: 1 })).toBe(true);
  });

  it("lets a widget being moved overlap its old place", () => {
    expect(canPlace([photo], "photo", { col: 1, row: 1 })).toBe(false);
    expect(canPlace([photo], "photo", { col: 1, row: 1 }, photo.id)).toBe(true);
  });
});

describe("openCells", () => {
  it("lists every cell below the date and time for a one cell widget on an empty grid", () => {
    const cells = openCells([], "weather");
    expect(cells).toHaveLength(OPEN_CELLS);
    expect(cells.every((cell) => cell.row >= 1)).toBe(true);
  });

  it("lists only corners where the widget fits, in reading order", () => {
    expect(openCells([], "photo")).toEqual([
      { col: 0, row: 1 },
      { col: 1, row: 1 },
      { col: 2, row: 1 },
    ]);
  });

  it("leaves out cells under an existing widget", () => {
    const cells = openCells([photo], "weather");
    expect(cells).toHaveLength(OPEN_CELLS - 4);
    expect(cells).not.toContainEqual({ col: 1, row: 2 });
    expect(cells).toContainEqual({ col: 2, row: 1 });
  });

  it("is empty when nothing of that size fits", () => {
    const full: PlacedWidget[] = [photo, { id: 3, type: "photo", col: 2, row: 1 }];
    expect(openCells(full, "message")).toEqual([]);
    expect(openCells(full, "weather")).toEqual([]);
  });
});

describe("hasRoom", () => {
  it("is true when a cell is open and false when the grid is full", () => {
    expect(hasRoom([], "photo")).toBe(true);
    const full: PlacedWidget[] = [
      { id: 1, type: "photo", col: 0, row: 1 },
      { id: 2, type: "photo", col: 2, row: 1 },
    ];
    expect(hasRoom(full, "weather")).toBe(false);
  });

  it("is false for a photo when only single cells are left", () => {
    const layout: PlacedWidget[] = [
      photo,
      { id: 3, type: "weather", col: 2, row: 1 },
      { id: 4, type: "message", col: 2, row: 2 },
    ];
    expect(hasRoom(layout, "photo")).toBe(false);
    expect(hasRoom(layout, "sports")).toBe(true);
  });
});

describe("isValidPlacement", () => {
  it("accepts an empty list and widgets that fit apart", () => {
    expect(isValidPlacement([])).toBe(true);
    expect(isValidPlacement([photo, weather])).toBe(true);
  });

  it("rejects overlapping widgets", () => {
    expect(isValidPlacement([photo, { id: 9, type: "sports", col: 1, row: 2 }])).toBe(false);
  });

  it("rejects a widget off the grid", () => {
    expect(isValidPlacement([{ id: 9, type: "message", col: 3, row: 1 }])).toBe(false);
  });

  it("rejects a widget in the date and time row", () => {
    expect(isValidPlacement([{ id: 9, type: "weather", col: 3, row: 0 }])).toBe(false);
    expect(isValidPlacement([{ id: 9, type: "photo", col: 0, row: 0 }])).toBe(false);
  });
});

describe("widgets from the server with an unknown type", () => {
  const fromServer = [
    { id: 1, type: "photo", col: 0, row: 1 },
    { id: 2, type: "clock", col: 2, row: 1 },
    { id: 3, type: "weather", col: 3, row: 2 },
  ];

  it("fail the whole check if passed in as they are", () => {
    expect(isValidPlacement(fromServer as unknown as PlacedWidget[])).toBe(false);
  });

  it("are checked and placed around once split off", () => {
    const { known } = splitKnownWidgets(fromServer);
    expect(isValidPlacement(known)).toBe(true);
    expect(canPlace(known, "weather", { col: 2, row: 1 })).toBe(true);
  });
});
