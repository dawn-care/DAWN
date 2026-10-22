/**
 * Home.tsx
 *
 * User: Caregiver
 *
 * Carepartner Screen page at `/`, and the redirect target after login,
 * account setup, and carepartner selection. Shows the carepartner's saved
 * screen as the tablet draws it (ScreenPreview's iframe). Edit screen swaps it
 * for the placeholder grid of a draft copy of the saved layout (useScreenEditor),
 * with the add bar (greyed out until then) above it, and Cancel throws the draft away.
 *
 * Selecting a widget with settings shows them in the panel (WidgetPanel) beside the screen.
 *
 * Save changes checks the draft, sends the whole list in one request and loads
 * the screen again. If the check or the save fails it says why and stays in edit
 * mode with the draft as it was.
 */

import { widgetName } from "@shared";
import { Check, Pencil } from "lucide-react";
import { useCallback, useEffect, useRef, useState } from "react";
import InfoPopover from "../components/ui/InfoPopover";
import { FOCUS_RING } from "../components/ui/styles";
import { useToast } from "../components/ui/toastContext";
import type { ToastOptions } from "../components/ui/toastContext";
import { useAuth } from "../context/AuthContext";
import { failedWidgetIndex, friendlyError, isUnreadableSuccess, saveWidgets } from "../screen/api";
import AddBar from "../screen/AddBar";
import { toLayout } from "../screen/layout";
import type { EditorWidget } from "../screen/layout";
import type { AddableWidgetType } from "../screen/useScreenEditor";
import ScreenGrid from "../screen/ScreenGrid";
import WidgetPanel from "../screen/WidgetPanel";
import { hasPanel, saveProblem, settingsChanged, settingsOf } from "../screen/widgetSettings";
import ScreenPreview from "../screen/ScreenPreview";
import { useSavedLayout } from "../screen/useSavedLayout";
import { useScreenEditor } from "../screen/useScreenEditor";

const PRIMARY_BUTTON = `flex items-center gap-2 h-11 px-[18px] rounded-[10px] border-2 border-accent bg-accent hover:brightness-95 text-on-accent text-base font-bold cursor-pointer disabled:opacity-50 disabled:cursor-default ${FOCUS_RING}`;
const SECONDARY_BUTTON = `h-11 px-[18px] rounded-[10px] border-2 border-ink-muted bg-surface hover:bg-page text-ink text-base font-bold cursor-pointer ${FOCUS_RING}`;

const PAGE_INFO =
  "This is where you configure what the carepartner's idle screen will look like. Click Edit screen to change what it looks like.";

export default function Home() {
  // RequireCarePartner only renders this page once a carepartner is chosen
  const { activeCarePartnerId } = useAuth();
  const carepartnerId = activeCarePartnerId as number;
  // The saved layout is what editing starts from; the screen itself doesn't need it
  const { state, reload } = useSavedLayout(carepartnerId);
  const editor = useScreenEditor(state.status === "ready" ? state.layout : null);

  const toast = useToast();
  // Toasts about edits can't be undone once editing ends, so they close with it
  const toastIds = useRef<number[]>([]);
  const announce = (text: string, options?: ToastOptions) => {
    toastIds.current.push(toast.show(text, options));
  };
  const dismissEditToasts = useCallback(() => {
    toastIds.current.forEach((id) => toast.dismiss(id));
    toastIds.current = [];
  }, [toast]);
  useEffect(() => {
    if (!editor.editing) dismissEditToasts();
  }, [editor.editing, dismissEditToasts]);

  // Shown after trying to add something while a new widget is unsaved, until that widget is saved or removed
  const unfinished = editor.draft?.find((widget) => widget.id === editor.unfinishedId) ?? null;
  const finishType = editor.finishFirst && unfinished ? (unfinished.type as AddableWidgetType) : null;

  const [saving, setSaving] = useState(false);
  const [saveError, setSaveError] = useState<string | null>(null);
  // Changes each save, so the screen's iframe loads what was just saved
  const [reloadKey, setReloadKey] = useState(0);

  // An error belongs to the edit it came from
  useEffect(() => {
    if (!editor.editing) setSaveError(null);
  }, [editor.editing]);

  const saveChanges = async () => {
    const draft = editor.draft;
    if (state.status !== "ready" || !draft || saving) return;
    setSaveError(null);

    const problem = saveProblem(draft);
    if (problem) {
      setSaveError(problem.message);
      if (problem.widgetId !== undefined) selectWidget(problem.widgetId);
      return;
    }

    // The draft is sent as it is now, so nothing can change it until the save is done: placing stops,
    // the screen and the panel are locked, and Undo on an edit toast is no longer offered
    editor.pick(null);
    dismissEditToasts();
    setSaving(true);
    try {
      await saveWidgets(carepartnerId, toLayout(draft, state.layout.unknown));
      // Loading the saved screen again ends editing, so the draft is replaced by what the server kept
      reload();
      setReloadKey((key) => key + 1);
      toast.show("Changes saved.");
    } catch (err) {
      console.error(`Saving carepartner ${carepartnerId}'s screen failed`, err);
      if (isUnreadableSuccess(err)) {
        // The server said yes, so the save may have gone through. Keeping the draft would let another Save
        // create its new widgets a second time, so load what the server kept instead.
        reload();
        setReloadKey((key) => key + 1);
        toast.show("Your changes may have been saved. The screen was loaded again to show what was kept.");
        return;
      }
      setSaveError(friendlyError(err));
      // The server names the widget it rejected by its place in the list sent, which starts with the draft
      const failed = failedWidgetIndex(err);
      if (failed !== undefined && draft[failed]) selectWidget(draft[failed].id);
    } finally {
      setSaving(false);
    }
  };

  const nameOf = (widget: EditorWidget) => widgetName({ type: widget.type }).toLowerCase();

  // Removing a widget offers Undo, which puts it back as it was, still unsaved if it was new
  const removeWidget = (id: EditorWidget["id"]) => {
    const widget = editor.draft?.find((other) => other.id === id);
    if (!widget) return;
    const wasNew = editor.isNew(id);
    editor.remove(id);
    announce(`Removed ${nameOf(widget)}.`, {
      actionLabel: "Undo",
      onAction: () => {
        if (!editor.restore(widget, wasNew)) announce("That spot is taken now, so it couldn't be put back.");
      },
    });
  };

  const selected = editor.draft?.find((widget) => widget.id === editor.selectedId) ?? null;

  // Selecting a widget stops any placing in progress
  const selectWidget = (id: EditorWidget["id"]) => {
    editor.pick(null);
    editor.select(id);
  };

  return (
    <div className="flex flex-col gap-3 px-7 pt-5 pb-7 h-screen">
      <div className="flex flex-wrap items-center justify-between gap-4 min-h-11">
        <div className="flex items-center gap-1">
          <h1 className="m-0 text-3xl font-extrabold">Carepartner Screen</h1>
          <InfoPopover label="About this page">{PAGE_INFO}</InfoPopover>
        </div>
        <div className="flex flex-wrap items-center gap-3">
          {editor.editing ? (
            <>
              <span className="text-base text-link">Not saved yet</span>
              <button type="button" onClick={editor.cancel} disabled={saving} className={SECONDARY_BUTTON}>
                Cancel
              </button>
              <button type="button" onClick={saveChanges} disabled={saving} className={PRIMARY_BUTTON}>
                {saving ? "Saving…" : "Save changes"}
              </button>
            </>
          ) : (
            <>
              {state.status === "ready" && (
                <span className="flex items-center gap-1.5 text-base text-status-green-text">
                  <Check size={18} strokeWidth={2.4} aria-hidden="true" />
                  Saved
                </span>
              )}
              {/* Editing starts from the saved layout, so wait for it */}
              <button
                type="button"
                onClick={editor.start}
                disabled={state.status !== "ready"}
                className={PRIMARY_BUTTON}
              >
                <Pencil size={18} strokeWidth={2.2} aria-hidden="true" />
                Edit screen
              </button>
            </>
          )}
        </div>
      </div>

      {saveError && (
        <div role="alert" className="text-base text-link">
          {saveError}
        </div>
      )}

      {state.status === "error" && (
        <div role="alert" className="flex flex-wrap items-center gap-3 text-base text-ink-muted">
          <p className="m-0">{state.message}</p>
          <button type="button" onClick={reload} className={PRIMARY_BUTTON}>
            Try again
          </button>
        </div>
      )}

      <AddBar
        editing={editor.editing}
        saving={saving}
        pendingType={editor.pendingType}
        fullType={editor.fullType}
        finishType={finishType}
        onPick={editor.pick}
      />

      {/* The screen on the left, and the panel for the selected widget's settings, always there, on the right */}
      <div className="flex flex-1 min-h-0 gap-6">
        <ScreenPreview
          carepartnerId={carepartnerId}
          reloadKey={reloadKey}
          editView={
            editor.draft ? (
              <ScreenGrid
                widgets={editor.draft}
                selectedId={editor.selectedId}
                pendingType={editor.pendingType}
                disabled={saving}
                onPlace={editor.add}
                onSelect={selectWidget}
                onEdit={selectWidget}
                onRemove={removeWidget}
              />
            ) : undefined
          }
        />
        <div className="w-[330px] shrink-0 pt-2">
          <WidgetPanel
            carepartnerId={carepartnerId}
            widget={selected && hasPanel(selected.type) ? selected : null}
            disabled={saving}
            onChange={(changes) => selected && editor.update(selected.id, changes)}
            // A new widget is announced when its panel's Save is pressed, not when it is placed
            onDone={(original) => {
              if (!selected) return;
              const wasNew = editor.isNew(selected.id);
              editor.keep(selected.id);
              editor.select(null);
              if (wasNew) {
                announce(`Added ${nameOf(selected)}.`);
              } else if (settingsChanged(original, selected)) {
                announce(`Updated ${nameOf(selected)}.`, {
                  actionLabel: "Undo",
                  onAction: () => editor.update(selected.id, settingsOf(original)),
                });
              }
            }}
          />
        </div>
      </div>

      {state.status === "ready" && state.layout.dropped.length > 0 && (
        <p className="m-0 text-center text-[15px] text-ink-muted">
          {state.layout.dropped.length === 1
            ? "1 thing saved on the carepartner's screen can't be shown on the tablet."
            : `${state.layout.dropped.length} things saved on the carepartner's screen can't be shown on the tablet.`}
          {editor.editing && " Saving removes it from the screen."}
        </p>
      )}
    </div>
  );
}
