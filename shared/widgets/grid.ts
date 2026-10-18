/**
 * The always-on screen's grid size, in cells. The server checks saved layouts
 * against the same size (GridColumns and GridRows in
 * server/internal/services/screen_widgets/grid.go), so change both together.
 */
export const GRID = {
  columns: 4,
  rows: 3,
} as const;

/** A spot on the grid, in cells. Columns and rows count from 0. */
export type GridArea = {
  col: number;
  row: number;
  width: number;
  height: number;
};

/** Whether an area fits on the grid; anything else is dropped rather than drawn off-screen. */
export function fitsGrid(
  area: GridArea,
  spec: { columns: number; rows: number } = GRID,
): boolean {
  const whole = [area.col, area.row, area.width, area.height].every(Number.isInteger);
  return (
    whole &&
    area.col >= 0 &&
    area.row >= 0 &&
    area.width >= 1 &&
    area.height >= 1 &&
    area.col + area.width <= spec.columns &&
    area.row + area.height <= spec.rows
  );
}

/** Whether two areas share at least one cell. */
export function areasOverlap(a: GridArea, b: GridArea): boolean {
  return a.col < b.col + b.width && b.col < a.col + a.width && a.row < b.row + b.height && b.row < a.row + a.height;
}

/** One cell of the grid. Columns and rows count from 0. */
export type GridCell = {
  col: number;
  row: number;
};

/**
 * A widget's place as the screen layout API stores it: its top-left and
 * bottom-right cells, both inclusive. The keys match the API's JSON, so a
 * widget from `/screen-layout` can be passed in directly.
 */
export type GridCorners = {
  top_left: GridCell;
  bottom_right: GridCell;
};

/**
 * Converts corners to an area. Both corners are inclusive, so a widget whose
 * corners are the same cell is 1x1. Use this rather than converting by hand.
 */
export function cornersToArea(corners: GridCorners): GridArea {
  const { top_left, bottom_right } = corners;
  return {
    col: top_left.col,
    row: top_left.row,
    width: bottom_right.col - top_left.col + 1,
    height: bottom_right.row - top_left.row + 1,
  };
}

/** Converts an area to the corners the screen layout API takes. */
export function areaToCorners(area: GridArea): GridCorners {
  return {
    top_left: { col: area.col, row: area.row },
    bottom_right: { col: area.col + area.width - 1, row: area.row + area.height - 1 },
  };
}
