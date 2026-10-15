import { GRID as SHARED_GRID, type GridArea } from "@shared";
import { space } from "../theme/theme";

/** The always-on screen's grid: the shared cell counts plus the tablet's padding and gap. */
export const GRID = {
  ...SHARED_GRID,
  gap: space.md,
  padding: space.lg,
} as const;

export type GridSpec = { columns: number; rows: number; gap: number; padding: number };
export type Size = { width: number; height: number };
export type Rect = { left: number; top: number; width: number; height: number };

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
