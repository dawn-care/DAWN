/**
 * gridPlacement.ts
 *
 * Where widgets can go on the screen's 4 x 3 grid (grid.ts). The day and time
 * always fills the top left (`DAY_TIME_AREA`), so no widget can go there, and
 * `day_time` itself is never placed. The server doesn't check this; only these
 * rules and the tablet do. Each widget type has a fixed size in cells
 * (`WIDGET_SIZES` in widgetSizes.ts) and sits at one cell, its top left corner,
 * so a widget is placed as a type plus a `col` and `row`. Cells count from 0
 * with (0, 0) at the top left, the same as `GridArea`. The caregiver highlights
 * `openCells` for the widget being placed and sends the whole list of placed
 * widgets when saving, with corners from `widgetArea`, so the tablet draws each
 * one there.
 */

import { areasOverlap, fitsGrid, GRID } from "./grid";
import type { GridArea } from "./grid";
import type { ScreenWidgetType } from "./screenWidgets";
import { hasWidgetSize, WIDGET_SIZES } from "./widgetSizes";

/** A cell on the grid. Columns and rows count from 0. */
export type Cell = { col: number; row: number };

/** A widget on the grid: its type and the cell holding its top left corner. */
export type PlacedWidget = Cell & {
  /** Tells widgets apart, so a widget being moved doesn't block its own new place. */
  id: string | number;
  type: ScreenWidgetType;
};

/**
 * The day and time of day, in the top left of every carepartner screen. It
 * isn't stored: both apps draw it here, and no widget may cover it.
 */
export const DAY_TIME_AREA: Readonly<GridArea> = { col: 0, row: 0, ...WIDGET_SIZES.day_time };

/** Parts of the grid every screen has, which no widget may cover. */
export const FIXED_AREAS: readonly Readonly<GridArea>[] = [DAY_TIME_AREA];

/** Whether an area is on the grid and clear of the fixed areas, ignoring other widgets. */
function fitsScreen(area: GridArea): boolean {
  return fitsGrid(area) && FIXED_AREAS.every((fixed) => !areasOverlap(area, fixed));
}

/** Whether a widget of this type and area can be on the screen; the day and time is never placed, anywhere. */
function placeable(type: ScreenWidgetType, area: GridArea): boolean {
  return type !== "day_time" && fitsScreen(area);
}

/** The cells a widget of this type covers when its top left corner is at `cell`. */
export function widgetArea(widget: Cell & { type: ScreenWidgetType }): GridArea {
  return { col: widget.col, row: widget.row, ...WIDGET_SIZES[widget.type] };
}

/**
 * Whether a widget of this type can have its top left corner at `cell`: it stays
 * on the grid, off the fixed areas, and covers no other widget. Never true for
 * `day_time`. Pass `movingId` when moving a widget so it isn't blocked by its
 * own old place.
 */
export function canPlace(
  widgets: readonly PlacedWidget[],
  type: ScreenWidgetType,
  cell: Cell,
  movingId?: PlacedWidget["id"],
): boolean {
  const area = widgetArea({ type, ...cell });
  if (!placeable(type, area)) return false;
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
 * Whether every widget fits on the grid, off the fixed areas, and none overlap,
 * with no `day_time` among them. Checked before saving the list.
 */
export function isValidPlacement(widgets: readonly PlacedWidget[]): boolean {
  return widgets.every((widget, index) => {
    const area = widgetArea(widget);
    return (
      placeable(widget.type, area) &&
      widgets.slice(index + 1).every((other) => !areasOverlap(area, widgetArea(other)))
    );
  });
}

/**
 * Checks a saved layout's widgets one at a time, in the server's order, for
 * whether the screen draws them, so the tablet and the caregiver's editor drop
 * the same ones. A widget is drawn if its type is known and isn't `day_time`,
 * it is its type's size, and it is on the grid, off the fixed areas and clear
 * of every widget drawn before it, so of two that overlap the first wins. Make
 * a new check for each layout.
 */
export function createDrawCheck(): (type: string, area: GridArea) => boolean {
  const drawn: GridArea[] = [];
  return (type, area) => {
    const ok =
      type !== "day_time" &&
      hasWidgetSize(type, area) &&
      fitsScreen(area) &&
      drawn.every((other) => !areasOverlap(area, other));
    if (ok) drawn.push(area);
    return ok;
  };
}
