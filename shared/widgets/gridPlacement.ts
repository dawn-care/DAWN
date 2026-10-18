/**
 * gridPlacement.ts
 *
 * Where widgets can go on the screen's 4 x 3 grid (grid.ts). Each widget type has
 * a fixed size in cells (`WIDGET_SIZES` in widgetSizes.ts) and sits at one cell,
 * its top left corner, so a widget is placed as a type plus a `col` and `row`. Cells count from 0 with (0, 0) at the
 * top left, the same as `GridArea`. The caregiver highlights `openCells` for the
 * widget being placed and sends the whole list of placed widgets when saving; the
 * tablet draws each one at `widgetArea`.
 */

import { areasOverlap, fitsGrid, GRID } from "./grid";
import type { GridArea } from "./grid";
import { LOCKED_WIDGET_TYPES } from "./screenWidgets";
import type { ScreenWidgetType } from "./screenWidgets";
import { WIDGET_SIZES } from "./widgetSizes";

/** A cell on the grid. Columns and rows count from 0. */
export type Cell = { col: number; row: number };

/** A widget on the grid: its type and the cell holding its top left corner. */
export type PlacedWidget = Cell & {
  /** Tells widgets apart, so a widget being moved doesn't block its own new place. */
  id: string | number;
  type: ScreenWidgetType;
};

/** The cells a widget of this type covers when its top left corner is at `cell`. */
export function widgetArea(widget: Cell & { type: ScreenWidgetType }): GridArea {
  return { col: widget.col, row: widget.row, ...WIDGET_SIZES[widget.type] };
}

/**
 * Whether a widget of this type can have its top left corner at `cell`: it stays
 * on the grid and covers no other widget. Pass `movingId` when moving a widget so
 * it isn't blocked by its own old place.
 */
export function canPlace(
  widgets: readonly PlacedWidget[],
  type: ScreenWidgetType,
  cell: Cell,
  movingId?: PlacedWidget["id"],
): boolean {
  const area = widgetArea({ type, ...cell });
  if (!fitsGrid(area)) return false;
  return widgets.every((other) => other.id === movingId || !areasOverlap(area, widgetArea(other)));
}

/**
 * Every cell a widget of this type can be placed at, in reading order (rows top
 * to bottom, then columns left to right). The caregiver highlights these.
 */
export function openCells(
  widgets: readonly PlacedWidget[],
  type: ScreenWidgetType,
  movingId?: PlacedWidget["id"],
): Cell[] {
  const cells: Cell[] = [];
  for (let row = 0; row < GRID.rows; row += 1) {
    for (let col = 0; col < GRID.columns; col += 1) {
      if (canPlace(widgets, type, { col, row }, movingId)) cells.push({ col, row });
    }
  }
  return cells;
}

/** Whether a widget of this type fits anywhere. When false the caregiver sees "full". */
export function hasRoom(widgets: readonly PlacedWidget[], type: ScreenWidgetType): boolean {
  return openCells(widgets, type).length > 0;
}

/**
 * Whether every widget fits on the grid, none overlap, and there's at most one of
 * each locked type (`day_time`). Checked before saving the list.
 */
export function isValidPlacement(widgets: readonly PlacedWidget[]): boolean {
  const lockedTwice = LOCKED_WIDGET_TYPES.some(
    (type) => widgets.filter((widget) => widget.type === type).length > 1,
  );
  if (lockedTwice) return false;
  return widgets.every((widget, index) => {
    const area = widgetArea(widget);
    return (
      fitsGrid(area) &&
      widgets.slice(index + 1).every((other) => !areasOverlap(area, widgetArea(other)))
    );
  });
}
