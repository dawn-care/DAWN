/**
 * widgetSettings.ts
 *
 * User: Caregiver
 *
 * Which widget types have a settings panel in the editor, and the rules their
 * settings must follow. Used by the panel (WidgetPanel) as the caregiver types,
 * and again before saving.
 */

import { isValidPlacement } from "@shared";
import type { ScreenWidgetType } from "@shared";
import type { EditorWidget } from "./layout";
import type { MessageConfig, SportsConfig, WeatherConfig } from "./api";

/**
 * The types whose settings the editor changes in a panel. News has none, and
 * reminders are not changed in the editor.
 */
export const PANEL_TYPES: readonly ScreenWidgetType[] = ["photo", "message", "weather", "sports"];

/** Whether selecting a widget of this type opens a settings panel */
export function hasPanel(type: ScreenWidgetType): boolean {
  return PANEL_TYPES.includes(type);
}

/** The settings of a widget that the panel changes, as they are now */
export function settingsOf(widget: EditorWidget): Pick<EditorWidget, "config" | "photoMediaId" | "photoUrl"> {
  return { config: widget.config, photoMediaId: widget.photoMediaId, photoUrl: widget.photoUrl };
}

/** Whether two versions of a widget differ in the settings the panel changes */
export function settingsChanged(before: EditorWidget, after: EditorWidget): boolean {
  return (
    before.photoMediaId !== after.photoMediaId ||
    before.photoUrl !== after.photoUrl ||
    JSON.stringify(before.config) !== JSON.stringify(after.config)
  );
}

/**
 * What stops the screen from being saved, or null when it can be. A widget
 * missing what it needs is named by `widgetId`, so the editor can select it.
 */
export function saveProblem(widgets: readonly EditorWidget[]): { message: string; widgetId?: EditorWidget["id"] } | null {
  const incomplete = widgets.find((widget) => !isComplete(widget));
  if (incomplete) return { message: "Finish the settings of the selected widget before saving.", widgetId: incomplete.id };
  if (!isValidPlacement(widgets)) return { message: "A widget doesn't fit on the carepartner's screen." };
  return null;
}

/** A weather widget's zip code: exactly five digits */
export function isValidZip(zip: string): boolean {
  return /^\d{5}$/.test(zip);
}

/**
 * Whether a widget has everything it needs to be saved: a message has text, a
 * weather widget a five digit zip, a photo widget a photo, and a sports widget
 * both its league and its team. Types without settings are always complete.
 * The panel's Save is off until this is true, and so is saving the screen.
 */
export function isComplete(widget: EditorWidget): boolean {
  switch (widget.type) {
    case "message":
      return ((widget.config as Partial<MessageConfig>).text ?? "").trim() !== "";
    case "weather":
      return isValidZip((widget.config as Partial<WeatherConfig>).zip ?? "");
    case "photo":
      return widget.photoMediaId !== null;
    case "sports": {
      const { league, team } = widget.config as Partial<SportsConfig>;
      return !!league && !!team;
    }
    default:
      return true;
  }
}
