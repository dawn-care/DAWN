/**
 * screenWidgets.ts
 *
 * The kinds of widget a caregiver can put on the carepartner's screen, and what
 * to call them. Where a widget sits is the layout's job (layoutPresets.ts); this
 * file works out which widgets a layout shows and how to name the ones it hides.
 */

import { isSpotInLayout, layoutLabel } from "./layoutPresets";
import type { LayoutPreset } from "./layoutPresets";

export type ScreenWidgetType = "photo" | "message" | "reminder" | "weather" | "sports" | "news";

const WIDGET_LABELS: Readonly<Record<ScreenWidgetType, string>> = {
  photo: "Photo",
  message: "Message",
  reminder: "Reminder",
  weather: "Weather",
  sports: "Sports",
  news: "News",
};

/**
 * What to call a widget in the caregiver's words: its type, or for a sports
 * widget with a chosen team, the team's name (the caller looks the name up).
 */
export function widgetName(widget: { type: ScreenWidgetType; teamName?: string }): string {
  if (widget.type === "sports" && widget.teamName) return widget.teamName;
  return WIDGET_LABELS[widget.type];
}

/** The widgets that have a spot in this layout. Order is kept. */
export function shownWidgets<T extends { spot: number }>(
  widgets: readonly T[],
  preset: LayoutPreset,
): T[] {
  return widgets.filter((widget) => isSpotInLayout(widget.spot, preset));
}

/** The widgets with no spot in this layout. They are kept and come back with a bigger layout. */
export function hiddenWidgets<T extends { spot: number }>(
  widgets: readonly T[],
  preset: LayoutPreset,
): T[] {
  return widgets.filter((widget) => !isSpotInLayout(widget.spot, preset));
}

/**
 * The note naming the widgets that don't fit in a layout, for example
 * "Reminder and Buffalo Bills don't fit in Simple. They're kept, and come back
 * if you switch." Empty when everything fits.
 */
export function doesntFitNote(
  widgets: readonly { spot: number; name: string }[],
  preset: LayoutPreset,
): string {
  const names = hiddenWidgets(widgets, preset).map((widget) => widget.name);
  if (names.length === 0) return "";
  const many = names.length > 1;
  const list = many ? `${names.slice(0, -1).join(", ")} and ${names[names.length - 1]}` : names[0];
  return (
    `${list} ${many ? "don't" : "doesn't"} fit in ${layoutLabel(preset)}. ` +
    `${many ? "They're" : "It's"} kept, and ${many ? "come" : "comes"} back if you switch.`
  );
}
