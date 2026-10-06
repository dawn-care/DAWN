import { space } from "../theme/theme";
import type { GridArea } from "./types";

/** The always-on screen's grid. */
export const GRID = {
  columns: 4,
  rows: 3,
  gap: space.md,
  padding: space.lg,
} as const;

export type GridSpec = { columns: number; rows: number; gap: number; padding: number };
export type Size = { width: number; height: number };
export type Rect = { left: number; top: number; width: number; height: number };

/** Whether an area fits on the grid; anything else is dropped rather than drawn off-screen. */
export function fitsGrid(area: GridArea, spec: GridSpec = GRID): boolean {
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

/**
 * Where a grid area sits on a screen of the given size, in layout units.
 * Cells share the space left after the outer padding and the gaps between them.
 */
export function cellRect(area: GridArea, screen: Size, spec: GridSpec = GRID): Rect {
  const cellWidth = (screen.width - 2 * spec.padding - (spec.columns - 1) * spec.gap) / spec.columns;
  const cellHeight = (screen.height - 2 * spec.padding - (spec.rows - 1) * spec.gap) / spec.rows;
  return {
    left: spec.padding + area.col * (cellWidth + spec.gap),
    top: spec.padding + area.row * (cellHeight + spec.gap),
    width: area.width * cellWidth + (area.width - 1) * spec.gap,
    height: area.height * cellHeight + (area.height - 1) * spec.gap,
  };
}
