/**
 * gridPlacement.ts
 *
 * Where widgets can go on the screen's 4 x 3 grid (grid.ts). The top row always
 * holds the date and time (`DATE_TIME_AREA`), so the caregiver's widgets go in the
 * rows below it. Each widget type has a fixed size in cells and sits at one cell,
 * its top left corner, so a widget is stored as a type plus a `col` and `row`. Cells count from 0 with (0, 0) at the
 * top left, the same as `GridArea`. The caregiver highlights `openCells` for the
 * widget being placed and sends the whole list of placed widgets when saving; the
 * tablet draws each one at `widgetArea`.
 */

import { fitsGrid, GRID } from "./grid";
import type { GridArea } from "./grid";
import type { ScreenWidgetType } from "./screenWidgets";

/** A cell on the grid. Columns and rows count from 0. */
export type Cell = { col: number; row: number };

/** A widget on the grid: its type and the cell holding its top left corner. */
export type PlacedWidget = Cell & {
  /** Tells widgets apart, so a widget being moved doesn't block its own new place. */
  id: string | number;
  type: ScreenWidgetType;
};

/**
 * The date and time widget: the day of the week and the time of day, across the
 * whole top row of every screen. It isn't stored or placed by the caregiver, and
 * no other widget can cover it.
 */
export const DATE_TIME_AREA: Readonly<GridArea> = { col: 0, row: 0, width: GRID.columns, height: 1 };

/** Parts of the grid every screen has, which no widget can cover. */
export const FIXED_AREAS: readonly Readonly<GridArea>[] = [DATE_TIME_AREA];

/** How many cells each widget type covers. */
export const WIDGET_SIZES: Readonly<Record<ScreenWidgetType, { width: number; height: number }>> = {
  photo: { width: 2, height: 2 },
  message: { width: 2, height: 1 },
  reminder: { width: 2, height: 1 },
  news: { width: 2, height: 1 },
  weather: { width: 1, height: 1 },
  sports: { width: 1, height: 1 },
};

/** The cells a widget of this type covers when its top left corner is at `cell`. */
export function widgetArea(widget: Cell & { type: ScreenWidgetType }): GridArea {
  return { col: widget.col, row: widget.row, ...WIDGET_SIZES[widget.type] };
}

/** Whether two areas share at least one cell. */
export function areasOverlap(a: GridArea, b: GridArea): boolean {
  return (
    a.col < b.col + b.width &&
    b.col < a.col + a.width &&
    a.row < b.row + b.height &&
    b.row < a.row + a.height
  );
}

/** Whether an area is on the grid and clear of the fixed areas, ignoring other widgets. */
function fitsScreen(area: GridArea): boolean {
  return fitsGrid(area) && FIXED_AREAS.every((fixed) => !areasOverlap(area, fixed));
}

/**
 * Whether a widget of this type can have its top left corner at `cell`: it stays
 * on the grid, off the fixed areas, and covers no other widget. Pass `movingId` when moving a widget so
 * it isn't blocked by its own old place.
 */
export function canPlace(
  widgets: readonly PlacedWidget[],
  type: ScreenWidgetType,
  cell: Cell,
  movingId?: PlacedWidget["id"],
): boolean {
  const area = widgetArea({ type, ...cell });
  if (!fitsScreen(area)) return false;
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

/** Whether every widget fits on the grid, off the fixed areas, and none overlap. Checked before saving the list. */
export function isValidPlacement(widgets: readonly PlacedWidget[]): boolean {
  return widgets.every((widget, index) => {
    const area = widgetArea(widget);
    return (
      fitsScreen(area) &&
      widgets.slice(index + 1).every((other) => !areasOverlap(area, widgetArea(other)))
    );
  });
}
