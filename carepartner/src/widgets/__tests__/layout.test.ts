import { cellRect, fitsGrid, GRID } from "../layout";

describe("cellRect", () => {
  // At the visual design's size: (1280 - 2*48 - 3*32) / 4 = 272 wide, (800 - 2*48 - 2*32) / 3 tall.
  const screen = { width: 1280, height: 800 };
  const cellHeight = (800 - 96 - 64) / 3;

  it("places a single cell after the outer padding", () => {
    expect(cellRect({ col: 0, row: 0, width: 1, height: 1 }, screen)).toEqual({
      left: 48,
      top: 48,
      width: 272,
      height: cellHeight,
    });
  });

  it("includes the gaps a multi-cell area spans", () => {
    expect(cellRect({ col: 2, row: 0, width: 2, height: 3 }, screen)).toEqual({
      left: 48 + 2 * (272 + 32),
      top: 48,
      width: 2 * 272 + 32,
      height: 3 * cellHeight + 2 * 32,
    });
  });

  it("scales with the screen size", () => {
    const rect = cellRect({ col: 0, row: 0, width: 4, height: 3 }, { width: 1920, height: 1200 });
    expect(rect).toEqual({ left: 48, top: 48, width: 1920 - 96, height: 1200 - 96 });
  });
});

describe("fitsGrid", () => {
  it("accepts areas inside the grid", () => {
    expect(fitsGrid({ col: 0, row: 0, width: GRID.columns, height: GRID.rows })).toBe(true);
    expect(fitsGrid({ col: 3, row: 2, width: 1, height: 1 })).toBe(true);
  });

  it.each([
    ["running off the right", { col: 3, row: 0, width: 2, height: 1 }],
    ["running off the bottom", { col: 0, row: 2, width: 1, height: 2 }],
    ["starting before the grid", { col: -1, row: 0, width: 1, height: 1 }],
    ["with no size", { col: 0, row: 0, width: 0, height: 1 }],
    ["with fractional cells", { col: 0.5, row: 0, width: 1, height: 1 }],
  ])("rejects an area %s", (_name, area) => {
    expect(fitsGrid(area)).toBe(false);
  });
});
