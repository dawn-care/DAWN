/**
 * screenWidgets.ts
 *
 * The kinds of widget a caregiver can put on the carepartner's screen, and what
 * to call them. Where a widget sits and how big it is is gridPlacement.ts's job.
 */

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
