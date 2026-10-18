/**
 * useScreenEditor.ts
 *
 * User: Caregiver
 *
 * Edit mode for the home page. Edit screen copies the saved widgets into a
 * draft that every change works on; Cancel throws the draft away, so nothing
 * reaches the tablet until the draft is saved. A new saved layout (another
 * carepartner, a reload, or a save) ends editing.
 *
 * While editing, the caregiver can pick a widget type to place (`pendingType`),
 * place it on an open square (`add`), change a widget's settings (`update`),
 * or remove it (`remove`). `selectedId` is the widget being edited. A widget
 * added since editing began stays in `newIds` until `keep` says its settings
 * were confirmed, so closing its popup with X can take it back off the screen.
 *
 * FUTURE IMPLEMENTATION (#333): saving the draft is added in a later step.
 */

import { canPlace } from "@shared";
import type { Cell, ScreenWidgetType } from "@shared";
import { useCallback, useRef, useState } from "react";
import type { WidgetConfig } from "./api";
import type { EditorWidget, LoadedLayout } from "./layout";

/** The types the add bar offers: reminders are out of scope, and the day and time is never added */
export type AddableWidgetType = Exclude<ScreenWidgetType, "day_time" | "reminder">;

export const ADDABLE_WIDGET_TYPES: readonly AddableWidgetType[] = ["photo", "message", "weather", "sports", "news"];

/** The zip a new weather widget starts with (St. Ann's) */
export const DEFAULT_WEATHER_ZIP = "14617";

/** What a widget has before the caregiver has filled in anything */
function defaultConfig(type: AddableWidgetType): WidgetConfig {
  return type === "weather" ? { zip: DEFAULT_WEATHER_ZIP } : {};
}

/** The settings of a widget that can change in the editor */
export type WidgetChanges = Partial<Pick<EditorWidget, "config" | "photoMediaId" | "photoUrl">>;

export function useScreenEditor(saved: LoadedLayout | null) {
  // null while not editing
  const [draft, setDraft] = useState<EditorWidget[] | null>(null);
  const [selectedId, setSelectedId] = useState<EditorWidget["id"] | null>(null);
  const [pendingType, setPendingType] = useState<AddableWidgetType | null>(null);
  const [newIds, setNewIds] = useState<readonly EditorWidget["id"][]>([]);
  const nextId = useRef(0);

  const reset = (widgets: EditorWidget[] | null) => {
    setDraft(widgets);
    setSelectedId(null);
    setPendingType(null);
    setNewIds([]);
  };

  // Drop the draft when the layout it was copied from is replaced. Adjusting
  // state during render like this avoids drawing one frame with the old draft.
  const [draftOf, setDraftOf] = useState(saved);
  if (saved !== draftOf) {
    setDraftOf(saved);
    reset(null);
  }

  const start = useCallback(() => {
    if (saved) {
      setDraft(saved.widgets.map((widget) => ({ ...widget })));
      setSelectedId(null);
      setPendingType(null);
      setNewIds([]);
    }
  }, [saved]);

  const cancel = useCallback(() => {
    setDraft(null);
    setSelectedId(null);
    setPendingType(null);
    setNewIds([]);
  }, []);

  /** Picks the type to place next, or null to stop placing. Picking the same type again stops. */
  const pick = useCallback((type: AddableWidgetType | null) => {
    setPendingType((current) => (current === type ? null : type));
  }, []);

  /**
   * Places a new widget of the picked type, or of `type` when given, with its
   * top left corner at `cell`, and selects it. Does nothing if it doesn't fit
   * there. Returns the new widget's id, or null when nothing was placed.
   */
  const add = useCallback(
    (cell: Cell, type: AddableWidgetType | null = pendingType): EditorWidget["id"] | null => {
      if (!draft || !type || !canPlace(draft, type, cell)) return null;
      const id = `new-${nextId.current++}`;
      const widget: EditorWidget = {
        id,
        type,
        ...cell,
        config: defaultConfig(type),
        photoMediaId: null,
        photoUrl: null,
      };
      setDraft([...draft, widget]);
      setNewIds((ids) => [...ids, id]);
      setSelectedId(id);
      setPendingType(null);
      return id;
    },
    [draft, pendingType],
  );

  /** Changes a widget's settings; its type and place stay as they are */
  const update = useCallback((id: EditorWidget["id"], changes: WidgetChanges) => {
    setDraft((widgets) => widgets && widgets.map((widget) => (widget.id === id ? { ...widget, ...changes } : widget)));
  }, []);

  /** Takes a widget off the screen */
  const remove = useCallback((id: EditorWidget["id"]) => {
    setDraft((widgets) => widgets && widgets.filter((widget) => widget.id !== id));
    setNewIds((ids) => ids.filter((other) => other !== id));
    setSelectedId((current) => (current === id ? null : current));
  }, []);

  /** Selects a widget (the one being edited), or null to select none */
  const select = useCallback((id: EditorWidget["id"] | null) => setSelectedId(id), []);

  /** Marks a new widget's settings as confirmed, so closing its popup no longer takes it back off */
  const keep = useCallback((id: EditorWidget["id"]) => {
    setNewIds((ids) => ids.filter((other) => other !== id));
  }, []);

  return {
    editing: draft !== null,
    draft,
    selectedId,
    pendingType,
    /** Whether the widget was added since editing began and its settings aren't confirmed yet */
    isNew: (id: EditorWidget["id"]) => newIds.includes(id),
    start,
    cancel,
    pick,
    add,
    update,
    remove,
    select,
    keep,
  };
}
