/** The always-on screen's grid size, in cells. */
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
