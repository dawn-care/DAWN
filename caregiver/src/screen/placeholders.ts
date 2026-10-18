/**
 * placeholders.ts
 *
 * User: Caregiver
 *
 * What a widget's placeholder tile says in the editor: its name and one short
 * line about its settings ("14617", "NYY"). Placeholders only show that
 * something takes up those squares; how the widget really looks is the
 * tablet's job.
 */

import { widgetName } from "@shared";
import type { ScreenWidgetType } from "@shared";
import type { EditorWidget } from "./layout";

const isObject = (value: unknown): value is Record<string, unknown> => typeof value === "object" && value !== null;
const text = (value: unknown): string | null => (typeof value === "string" && value.trim() ? value.trim() : null);

/** One short line about each type's settings, or null when there's nothing to say */
const DETAILS: Record<ScreenWidgetType, (config: Record<string, unknown>) => string | null> = {
  day_time: (config) => (config.show_date === false ? "Without the date" : "With the date"),
  photo: (config) => text(config.caption),
  message: (config) => text(config.text),
  reminder: (config) => {
    const count = Array.isArray(config.items) ? config.items.length : 0;
    return count === 1 ? "1 reminder" : `${count} reminders`;
  },
  weather: (config) => text(config.zip),
  sports: (config) => {
    const team = text(config.team);
    const league = text(config.league);
    return team && league ? `${team.toUpperCase()} · ${league.toUpperCase()}` : null;
  },
  news: () => null,
};

/** The name and settings line a widget's placeholder shows */
export function placeholderText(widget: Pick<EditorWidget, "type" | "config">): { name: string; detail: string | null } {
  const config = isObject(widget.config) ? widget.config : {};
  return { name: widgetName({ type: widget.type }), detail: DETAILS[widget.type](config) };
}
