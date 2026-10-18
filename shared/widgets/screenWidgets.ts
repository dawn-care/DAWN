/**
 * screenWidgets.ts
 *
 * The kinds of widget on the carepartner's screen, and what to call them. Where a widget sits and how big it is is
 * gridPlacement.ts's job.
 */

export const SCREEN_WIDGET_TYPES = ["day_time", "photo", "message", "reminder", "weather", "sports", "news"] as const;

export type ScreenWidgetType = (typeof SCREEN_WIDGET_TYPES)[number];

/** Whether a string from the server is a widget type this app knows how to draw and place. */
export function isScreenWidgetType(type: string): type is ScreenWidgetType {
  return (SCREEN_WIDGET_TYPES as readonly string[]).includes(type);
}

/**
 * Splits widgets read from the server into those with a known type and the rest.
 * Place and check only the known ones; send the unknown ones back unchanged when
 * saving, so a newer server's widget isn't deleted by an older app.
 */
export function splitKnownWidgets<T extends { type: string }>(
  widgets: readonly T[],
): { known: (T & { type: ScreenWidgetType })[]; unknown: T[] } {
  const known: (T & { type: ScreenWidgetType })[] = [];
  const unknown: T[] = [];
  for (const widget of widgets) {
    if (isScreenWidgetType(widget.type)) known.push(widget as T & { type: ScreenWidgetType });
    else unknown.push(widget);
  }
  return { known, unknown };
}

const WIDGET_LABELS: Readonly<Record<ScreenWidgetType, string>> = {
  day_time: "Day and time",
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
