/**
 * widgetSizes.ts
 *
 * Each widget type has one fixed size on the grid; a widget in another size is
 * a different type. The caregiver's editor places widgets at these sizes, and
 * the tablet skips a widget whose size doesn't match, rather than draw it
 * stretched or cramped. The server doesn't check them yet.
 */

import type { GridArea } from "./grid";

export type WidgetSize = { width: number; height: number };

export const WIDGET_SIZES = {
  /** Day and time of day. The server doesn't have this type yet (#366). */
  day_time: { width: 2, height: 2 },
  photo: { width: 2, height: 2 },
  message: { width: 2, height: 1 },
  reminder: { width: 2, height: 1 },
  news: { width: 2, height: 1 },
  /** Wide enough for the team, the result and score, and the opponent. */
  sports: { width: 2, height: 1 },
  weather: { width: 1, height: 1 },
} as const satisfies Record<string, WidgetSize>;

export type WidgetType = keyof typeof WIDGET_SIZES;

/** The fixed size of a widget type, or undefined for a type this version doesn't know. */
export function widgetSize(type: string): WidgetSize | undefined {
  return Object.prototype.hasOwnProperty.call(WIDGET_SIZES, type) ? WIDGET_SIZES[type as WidgetType] : undefined;
}

/** Whether a widget of this type takes up exactly its fixed size. */
export function hasWidgetSize(type: string, area: GridArea): boolean {
  const size = widgetSize(type);
  return !!size && area.width === size.width && area.height === size.height;
}
