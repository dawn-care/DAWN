import {
  LAYOUTS,
  firstEmptySpot,
  isSpotInLayout,
  layoutLabel,
  layoutShape,
} from "../layoutPresets";
import type { LayoutName, LayoutPreset } from "../layoutPresets";

const simple: LayoutPreset = { name: "simple", spots: 2 };
const balanced: LayoutPreset = { name: "balanced", spots: 4 };
const grid: LayoutPreset = { name: "grid", spots: 4 };
const busier: LayoutPreset = { name: "busier", spots: 6 };

const layoutNames = Object.keys(LAYOUTS) as LayoutName[];

describe("LAYOUTS", () => {
  it.each([
    ["simple", 2],
    ["balanced", 4],
    ["grid", 4],
    ["busier", 6],
  ] as const)("%s has %i spots", (name, count) => {
    expect(LAYOUTS[name].spots).toHaveLength(count);
  });

  it.each(layoutNames)("%s keeps every spot on its grid", (name) => {
    const shape = LAYOUTS[name];
    expect(shape.columnWeights).toHaveLength(shape.columns);
    for (const area of shape.spots) {
      expect(area.col).toBeGreaterThanOrEqual(0);
      expect(area.row).toBeGreaterThanOrEqual(0);
      expect(area.col + area.width).toBeLessThanOrEqual(shape.columns);
      expect(area.row + area.height).toBeLessThanOrEqual(shape.rows);
    }
  });

  it.each(layoutNames)("%s covers every cell exactly once", (name) => {
    const shape = LAYOUTS[name];
    const uses = new Map<string, number>();
    for (const area of shape.spots) {
      for (let col = area.col; col < area.col + area.width; col += 1) {
        for (let row = area.row; row < area.row + area.height; row += 1) {
          const cell = `${col},${row}`;
          uses.set(cell, (uses.get(cell) ?? 0) + 1);
        }
      }
    }
    expect(uses.size).toBe(shape.columns * shape.rows);
    expect([...uses.values()].every((count) => count === 1)).toBe(true);
  });

  it("makes balanced's first spot a tall photo spot beside three stacked tiles", () => {
    const [first, ...rest] = LAYOUTS.balanced.spots;
    expect(first).toEqual({ col: 0, row: 0, width: 1, height: 3 });
    expect(rest.map((area) => area.row)).toEqual([0, 1, 2]);
    expect(rest.every((area) => area.col === 1)).toBe(true);
  });

  it("fills busier in reading order", () => {
    expect(LAYOUTS.busier.spots.map((area) => [area.row, area.col])).toEqual([
      [0, 0],
      [0, 1],
      [0, 2],
      [1, 0],
      [1, 1],
      [1, 2],
    ]);
  });
});

describe("layoutShape", () => {
  it("returns the shape for each known layout", () => {
    expect(layoutShape(simple)).toBe(LAYOUTS.simple);
    expect(layoutShape(busier)).toBe(LAYOUTS.busier);
  });

  it("returns nothing for an unknown name", () => {
    expect(layoutShape({ name: "huge", spots: 8 })).toBeUndefined();
    expect(layoutShape({ name: "constructor", spots: 2 })).toBeUndefined();
  });

  it("returns nothing when the server's spot count differs", () => {
    expect(layoutShape({ name: "simple", spots: 3 })).toBeUndefined();
  });
});

describe("isSpotInLayout", () => {
  it("accepts spots from 0 to one less than the count", () => {
    expect(isSpotInLayout(0, simple)).toBe(true);
    expect(isSpotInLayout(1, simple)).toBe(true);
    expect(isSpotInLayout(2, simple)).toBe(false);
  });

  it("rejects negative, fractional and missing spots", () => {
    expect(isSpotInLayout(-1, busier)).toBe(false);
    expect(isSpotInLayout(1.5, busier)).toBe(false);
    expect(isSpotInLayout(Number.NaN, busier)).toBe(false);
  });
});

describe("firstEmptySpot", () => {
  it("is spot 0 on an empty screen", () => {
    expect(firstEmptySpot([], balanced)).toBe(0);
  });

  it("skips taken spots", () => {
    expect(firstEmptySpot([{ spot: 0 }, { spot: 1 }], balanced)).toBe(2);
  });

  it("finds a gap before later spots", () => {
    expect(firstEmptySpot([{ spot: 0 }, { spot: 2 }, { spot: 3 }], balanced)).toBe(1);
  });

  it("is null when every spot is taken", () => {
    expect(firstEmptySpot([{ spot: 0 }, { spot: 1 }], simple)).toBeNull();
  });

  it("ignores widgets kept in hidden spots", () => {
    expect(firstEmptySpot([{ spot: 4 }, { spot: 5 }], simple)).toBe(0);
  });

  it("does not let a hidden widget fill a visible spot", () => {
    expect(firstEmptySpot([{ spot: 0 }, { spot: 3 }], simple)).toBe(1);
  });
});

describe("layoutLabel", () => {
  it("names each known layout", () => {
    expect(layoutLabel(simple)).toBe("Simple");
    expect(layoutLabel(busier)).toBe("Busier");
  });

  it("shows an unknown layout as the server sent it", () => {
    expect(layoutLabel({ name: "huge", spots: 8 })).toBe("huge");
  });
});
