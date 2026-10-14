import { fitsGrid, GRID } from "../grid";

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
