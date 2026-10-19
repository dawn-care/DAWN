/**
 * layout.ts
 *
 * User: Caregiver
 *
 * Converts between the screen layout API's widgets (two inclusive corners, see
 * api.ts) and the editor's widgets (a type at one cell, as the placement rules
 * in @shared take them). A widget's size always comes from its type
 * (`WIDGET_SIZES`), so the corners sent on save always match what the tablet
 * draws.
 */

import {
  areaToCorners,
  areasOverlap,
  cornersToArea,
  FIXED_AREAS,
  fitsGrid,
  hasWidgetSize,
  isScreenWidgetType,
  widgetArea,
} from "@shared";
import type { GridArea, PlacedWidget } from "@shared";
import type { LayoutWidget, LayoutWidgetInput, WidgetConfig } from "./api";

/**
 * A widget in the editor. `id` is the server's id for a saved widget, or a
 * string the editor makes up for one added since the last save.
 */
export type EditorWidget = PlacedWidget & {
  config: WidgetConfig;
  photoMediaId: number | null;
  /** Path to the photo, relative to the API base; use photoSrc() from api.ts */
  photoUrl: string | null;
};

export interface LoadedLayout {
  /** Widgets the editor shows and lets the caregiver move, change or remove */
  widgets: EditorWidget[];
  /**
   * Widgets of a type this app doesn't know (from a newer server). They aren't
   * shown or checked, and are sent back unchanged on save so they aren't deleted.
   */
  unknown: LayoutWidget[];
  /**
   * Widgets the tablet can't draw either: a stored day and time, not their
   * type's size, off the grid, or overlapping the day and time or a widget
   * before them. They aren't shown, and saving removes them from the screen.
   */
  dropped: LayoutWidget[];
}

/**
 * Splits the server's widgets into what the editor shows, what it keeps aside
 * unchanged, and what it drops. Widgets are checked in the order given (the
 * server's reading order), so of two overlapping widgets the later one is
 * dropped, as on the tablet.
 */
export function fromLayout(loaded: readonly LayoutWidget[]): LoadedLayout {
  const result: LoadedLayout = { widgets: [], unknown: [], dropped: [] };
  // The day and time is always in the top left, so nothing saved can be there
  const taken: GridArea[] = [...FIXED_AREAS];

  for (const widget of loaded) {
    const { type } = widget;
    if (!isScreenWidgetType(type)) {
      result.unknown.push(widget);
      continue;
    }
    const area = cornersToArea(widget);
    // A stored day and time is never shown: the tablet draws its own in the top left
    const placeable =
      type !== "day_time" &&
      hasWidgetSize(type, area) &&
      fitsGrid(area) &&
      !taken.some((other) => areasOverlap(other, area));
    if (!placeable) {
      result.dropped.push(widget);
      continue;
    }
    taken.push(area);
    result.widgets.push({
      id: widget.id,
      type,
      col: area.col,
      row: area.row,
      config: widget.config,
      photoMediaId: widget.photo_media_id,
      photoUrl: widget.photo_url,
    });
  }
  return result;
}

/**
 * The list to send with saveWidgets: every editor widget, with corners from its
 * type's size, then the unknown widgets as they came. A widget added since the
 * last save (a string id) is sent without an id, so the server creates it.
 */
export function toLayout(widgets: readonly EditorWidget[], unknown: readonly LayoutWidget[] = []): LayoutWidgetInput[] {
  const edited = widgets.map(
    (widget): LayoutWidgetInput => ({
      ...(typeof widget.id === "number" ? { id: widget.id } : {}),
      type: widget.type,
      ...areaToCorners(widgetArea(widget)),
      config: widget.config,
      photo_media_id: widget.photoMediaId,
    }),
  );
  const kept = unknown.map(
    (widget): LayoutWidgetInput => ({
      id: widget.id,
      type: widget.type,
      top_left: widget.top_left,
      bottom_right: widget.bottom_right,
      config: widget.config,
      photo_media_id: widget.photo_media_id,
    }),
  );
  return [...edited, ...kept];
}
