import { GRID } from "../grid";
import {
  DAY_TIME_AREA,
  FIXED_AREAS,
  canPlace,
  hasRoom,
  isValidPlacement,
  openCells,
  widgetArea,
} from "../gridPlacement";
import type { PlacedWidget } from "../gridPlacement";
import { SCREEN_WIDGET_TYPES, splitKnownWidgets } from "../screenWidgets";
import { WIDGET_SIZES } from "../widgetSizes";

/*
 * The day and time always covers the top left 2 x 2, so widgets go in the eight
 * squares around it:
 *
 *   D D . .
 *   D D . .
 *   . . . .
 */
const OPEN_SQUARES = GRID.columns * GRID.rows - 4;

const photo: PlacedWidget = { id: 1, type: "photo", col: 2, row: 0 };
const weather: PlacedWidget = { id: 2, type: "weather", col: 0, row: 2 };

describe("sizes", () => {
  it.each(SCREEN_WIDGET_TYPES)("%s has a size in WIDGET_SIZES", (type) => {
    expect(WIDGET_SIZES[type]).toBeDefined();
  });
});

describe("DAY_TIME_AREA", () => {
  it("is the day and time's 2 x 2 in the top left, and the only fixed area", () => {
    expect(DAY_TIME_AREA).toEqual({ col: 0, row: 0, width: 2, height: 2 });
    expect(FIXED_AREAS).toEqual([DAY_TIME_AREA]);
  });
});

describe("widgetArea", () => {
  it("covers the widget's size starting at its cell", () => {
    expect(widgetArea({ type: "message", col: 1, row: 2 })).toEqual({ col: 1, row: 2, width: 2, height: 1 });
    expect(widgetArea({ type: "photo", col: 2, row: 1 })).toEqual({ col: 2, row: 1, width: 2, height: 2 });
    expect(widgetArea({ type: "sports", col: 0, row: 2 })).toEqual({ col: 0, row: 2, width: 2, height: 1 });
  });
});

describe("canPlace", () => {
  it("allows any cell on an empty grid where the widget stays on the grid and off the day and time", () => {
    expect(canPlace([], "weather", { col: 3, row: 2 })).toBe(true);
    expect(canPlace([], "photo", { col: 2, row: 1 })).toBe(true);
  });

  it("refuses any square the day and time covers, or a widget that would reach it", () => {
    for (const cell of [{ col: 0, row: 0 }, { col: 1, row: 0 }, { col: 0, row: 1 }, { col: 1, row: 1 }]) {
      expect(canPlace([], "weather", cell)).toBe(false);
    }
    expect(canPlace([], "photo", { col: 1, row: 0 })).toBe(false);
    expect(canPlace([], "message", { col: 1, row: 1 })).toBe(false);
  });

  it("refuses a widget that would run off the grid", () => {
    expect(canPlace([], "message", { col: 3, row: 0 })).toBe(false);
    expect(canPlace([], "sports", { col: 3, row: 2 })).toBe(false);
    expect(canPlace([], "photo", { col: 2, row: 2 })).toBe(false);
  });

  it("refuses a cell that is not on the grid", () => {
    expect(canPlace([], "weather", { col: -1, row: 2 })).toBe(false);
    expect(canPlace([], "weather", { col: 4, row: 2 })).toBe(false);
    expect(canPlace([], "weather", { col: 2.5, row: 2 })).toBe(false);
  });

  it("refuses a place another widget covers, including its far cells", () => {
    expect(canPlace([photo], "weather", { col: 2, row: 0 })).toBe(false);
    expect(canPlace([photo], "weather", { col: 3, row: 1 })).toBe(false);
    expect(canPlace([photo], "weather", { col: 3, row: 2 })).toBe(true);
  });

  it("refuses a widget whose body would reach a neighbour", () => {
    expect(canPlace([weather], "message", { col: 0, row: 2 })).toBe(false);
    expect(canPlace([weather], "message", { col: 1, row: 2 })).toBe(true);
  });

  it("lets a widget being moved overlap its old place", () => {
    expect(canPlace([photo], "photo", { col: 2, row: 1 })).toBe(false);
    expect(canPlace([photo], "photo", { col: 2, row: 1 }, photo.id)).toBe(true);
  });
});

describe("openCells", () => {
  it("lists every square around the day and time for a one cell widget on an empty grid", () => {
    const cells = openCells([], "weather");
    expect(cells).toHaveLength(OPEN_SQUARES);
    expect(cells).not.toContainEqual({ col: 0, row: 0 });
    expect(cells).not.toContainEqual({ col: 1, row: 1 });
  });

  it("lists only corners where the widget fits, in reading order", () => {
    expect(openCells([], "photo")).toEqual([
      { col: 2, row: 0 },
      { col: 2, row: 1 },
    ]);
    expect(openCells([], "message")).toEqual([
      { col: 2, row: 0 },
      { col: 2, row: 1 },
      { col: 0, row: 2 },
      { col: 1, row: 2 },
      { col: 2, row: 2 },
    ]);
  });

  it("leaves out cells under an existing widget", () => {
    const cells = openCells([photo], "weather");
    expect(cells).toHaveLength(OPEN_SQUARES - 4);
    expect(cells).not.toContainEqual({ col: 3, row: 1 });
    expect(cells).toContainEqual({ col: 2, row: 2 });
  });

  it("is empty when nothing of that size fits", () => {
    const full: PlacedWidget[] = [photo, { id: 3, type: "sports", col: 0, row: 2 }, { id: 4, type: "news", col: 2, row: 2 }];
    expect(openCells(full, "message")).toEqual([]);
    expect(openCells(full, "weather")).toEqual([]);
  });
});

describe("hasRoom", () => {
  it("is true when a cell is open and false when the grid is full", () => {
    expect(hasRoom([], "photo")).toBe(true);
    const full: PlacedWidget[] = [photo, { id: 3, type: "message", col: 0, row: 2 }, { id: 4, type: "reminder", col: 2, row: 2 }];
    expect(hasRoom(full, "weather")).toBe(false);
  });

  it("is false for a 2x1 widget when only single cells are left in each row", () => {
    const layout: PlacedWidget[] = [
      { id: 3, type: "weather", col: 2, row: 0 },
      { id: 4, type: "weather", col: 3, row: 1 },
      { id: 5, type: "weather", col: 1, row: 2 },
      { id: 6, type: "weather", col: 3, row: 2 },
    ];
    expect(hasRoom(layout, "sports")).toBe(false);
    expect(hasRoom(layout, "weather")).toBe(true);
  });
});

describe("isValidPlacement", () => {
  it("accepts an empty list and widgets that fit apart", () => {
    expect(isValidPlacement([])).toBe(true);
    expect(isValidPlacement([photo, weather])).toBe(true);
  });

  it("rejects overlapping widgets", () => {
    expect(isValidPlacement([photo, { id: 9, type: "weather", col: 3, row: 1 }])).toBe(false);
  });

  it("rejects a widget off the grid", () => {
    expect(isValidPlacement([{ id: 9, type: "message", col: 3, row: 2 }])).toBe(false);
  });

  it("rejects a widget over the day and time, including a saved day and time widget", () => {
    expect(isValidPlacement([{ id: 9, type: "weather", col: 1, row: 1 }])).toBe(false);
    expect(isValidPlacement([{ id: 9, type: "day_time", col: 0, row: 0 }])).toBe(false);
  });
});

describe("widgets from the server with an unknown type", () => {
  const fromServer = [
    { id: 1, type: "photo", col: 2, row: 0 },
    { id: 2, type: "clock", col: 0, row: 2 },
    { id: 3, type: "weather", col: 3, row: 2 },
  ];

  it("fail the whole check if passed in as they are", () => {
    expect(isValidPlacement(fromServer as unknown as PlacedWidget[])).toBe(false);
  });

  it("are checked and placed around once split off", () => {
    const { known } = splitKnownWidgets(fromServer);
    expect(isValidPlacement(known)).toBe(true);
    expect(canPlace(known, "weather", { col: 0, row: 2 })).toBe(true);
  });
});
