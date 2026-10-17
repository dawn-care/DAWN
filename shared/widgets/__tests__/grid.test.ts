import { areasOverlap, areaToCorners, cornersToArea, fitsGrid, GRID } from "../grid";

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

describe("cornersToArea", () => {
  it.each([
    ["1x1, the same cell for both corners", { top_left: { col: 2, row: 1 }, bottom_right: { col: 2, row: 1 } }, { col: 2, row: 1, width: 1, height: 1 }],
    ["2x2", { top_left: { col: 0, row: 0 }, bottom_right: { col: 1, row: 1 } }, { col: 0, row: 0, width: 2, height: 2 }],
    ["the whole grid", { top_left: { col: 0, row: 0 }, bottom_right: { col: 3, row: 2 } }, { col: 0, row: 0, width: GRID.columns, height: GRID.rows }],
  ])("converts %s", (_name, corners, area) => {
    expect(cornersToArea(corners)).toEqual(area);
    expect(areaToCorners(area)).toEqual(corners);
  });

  it("round-trips both ways", () => {
    const corners = { top_left: { col: 1, row: 0 }, bottom_right: { col: 3, row: 1 } };
    const area = { col: 0, row: 2, width: 4, height: 1 };
    expect(areaToCorners(cornersToArea(corners))).toEqual(corners);
    expect(cornersToArea(areaToCorners(area))).toEqual(area);
  });

  it("gives an area fitsGrid rejects when a corner is off the grid", () => {
    expect(fitsGrid(cornersToArea({ top_left: { col: 3, row: 0 }, bottom_right: { col: 4, row: 0 } }))).toBe(false);
  });
});

describe("areasOverlap", () => {
  const photo = { col: 0, row: 0, width: 2, height: 2 };

  it.each([
    ["the same area", photo],
    ["one cell in common", { col: 1, row: 1, width: 2, height: 1 }],
    ["an area inside it", { col: 1, row: 0, width: 1, height: 1 }],
  ])("is true for %s", (_name, other) => {
    expect(areasOverlap(photo, other)).toBe(true);
    expect(areasOverlap(other, photo)).toBe(true);
  });

  it.each([
    ["an area beside it", { col: 2, row: 0, width: 2, height: 1 }],
    ["an area below it", { col: 0, row: 2, width: 2, height: 1 }],
    ["an area touching only at a corner", { col: 2, row: 2, width: 1, height: 1 }],
  ])("is false for %s", (_name, other) => {
    expect(areasOverlap(photo, other)).toBe(false);
    expect(areasOverlap(other, photo)).toBe(false);
  });
});
