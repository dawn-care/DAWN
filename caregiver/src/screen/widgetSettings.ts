/**
 * widgetSettings.ts
 *
 * User: Caregiver
 *
 * Which widget types have a settings panel in the editor, and the rules their
 * settings must follow. Used by the panel (WidgetPanel) as the caregiver types,
 * and again before saving.
 */

import type { ScreenWidgetType } from "@shared";
import type { EditorWidget } from "./layout";

/**
 * The types whose settings the editor changes in a panel. News has none, and
 * reminders are not changed in the editor.
 *
 * FUTURE IMPLEMENTATION (#333): photo and sports are added to this list when
 * their panels exist.
 */
export const PANEL_TYPES: readonly ScreenWidgetType[] = ["message", "weather"];

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

/** A weather widget's zip code: exactly five digits */
export function isValidZip(zip: string): boolean {
  return /^\d{5}$/.test(zip);
}
