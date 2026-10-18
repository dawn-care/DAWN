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
 * FUTURE IMPLEMENTATION (#333): the photo and sports panels, and Save changes,
 * are added in later steps.
 */

import { widgetName } from "@shared";
import type { Cell } from "@shared";
import { Check, Pencil } from "lucide-react";
import { useEffect, useRef } from "react";
import InfoPopover from "../components/ui/InfoPopover";
import { FOCUS_RING } from "../components/ui/styles";
import { useToast } from "../components/ui/toastContext";
import type { ToastOptions } from "../components/ui/toastContext";
import { useAuth } from "../context/AuthContext";
import AddBar from "../screen/AddBar";
import type { EditorWidget } from "../screen/layout";
import ScreenGrid from "../screen/ScreenGrid";
import WidgetPanel from "../screen/WidgetPanel";
import { hasPanel, settingsChanged, settingsOf } from "../screen/widgetSettings";
import ScreenPreview from "../screen/ScreenPreview";
import { useSavedLayout } from "../screen/useSavedLayout";
import { useScreenEditor } from "../screen/useScreenEditor";

const PRIMARY_BUTTON = `flex items-center gap-2 h-11 px-[18px] rounded-[10px] border-2 border-dawn-accent bg-dawn-accent hover:bg-dawn-accent-hover hover:border-dawn-accent-hover text-white text-base font-bold cursor-pointer disabled:opacity-50 disabled:cursor-default ${FOCUS_RING}`;
const SECONDARY_BUTTON = `h-11 px-[18px] rounded-[10px] border-2 border-dawn-control bg-white hover:bg-dawn-page text-dawn-ink text-base font-bold cursor-pointer ${FOCUS_RING}`;

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
  useEffect(() => {
    if (editor.editing) return;
    toastIds.current.forEach((id) => toast.dismiss(id));
    toastIds.current = [];
  }, [editor.editing, toast]);

  const nameOf = (widget: EditorWidget) => widgetName({ type: widget.type }).toLowerCase();

  // A widget with settings is announced when its panel's Save is pressed; one without, as it is placed
  const placeWidget = (cell: Cell) => {
    const type = editor.pendingType;
    const id = editor.add(cell);
    if (id === null || !type || hasPanel(type)) return;
    editor.keep(id);
    announce(`Added ${widgetName({ type }).toLowerCase()}.`);
  };

  const removeWidget = (id: EditorWidget["id"]) => {
    const widget = editor.draft?.find((other) => other.id === id);
    if (!widget) return;
    editor.remove(id);
    announce(`Removed ${nameOf(widget)}.`, {
      actionLabel: "Undo",
      onAction: () => {
        if (!editor.restore(widget)) announce("That spot is taken now, so it couldn't be put back.");
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
              <span className="text-base text-dawn-accent">Not saved yet</span>
              <button type="button" onClick={editor.cancel} className={SECONDARY_BUTTON}>
                Cancel
              </button>
            </>
          ) : (
            <>
              {state.status === "ready" && (
                <span className="flex items-center gap-1.5 text-base text-dawn-success">
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

      {state.status === "error" && (
        <div role="alert" className="flex flex-wrap items-center gap-3 text-base text-dawn-muted">
          <p className="m-0">{state.message}</p>
          <button type="button" onClick={reload} className={PRIMARY_BUTTON}>
            Try again
          </button>
        </div>
      )}

      <AddBar
        editing={editor.editing}
        pendingType={editor.pendingType}
        fullType={editor.fullType}
        onPick={editor.pick}
      />

      {/* The screen on the left, and the panel for the selected widget's settings, always there, on the right */}
      <div className="flex flex-1 min-h-0 gap-6">
        <ScreenPreview
          carepartnerId={carepartnerId}
          editView={
            editor.draft ? (
              <ScreenGrid
                widgets={editor.draft}
                selectedId={editor.selectedId}
                pendingType={editor.pendingType}
                onPlace={placeWidget}
                onSelect={selectWidget}
                onEdit={selectWidget}
                onRemove={removeWidget}
              />
            ) : undefined
          }
        />
        <div className="w-[330px] shrink-0 pt-2">
          <WidgetPanel
            widget={selected && hasPanel(selected.type) ? selected : null}
            onChange={(changes) => selected && editor.update(selected.id, changes)}
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

      {/* FUTURE IMPLEMENTATION (#333): add that saving removes them, once Save changes exists */}
      {state.status === "ready" && state.layout.dropped.length > 0 && (
        <p className="m-0 text-center text-[15px] text-dawn-muted">
          {state.layout.dropped.length === 1
            ? "1 thing saved on the carepartner's screen can't be shown on the tablet."
            : `${state.layout.dropped.length} things saved on the carepartner's screen can't be shown on the tablet.`}
        </p>
      )}
    </div>
  );
}
