/**
 * ScreenGrid.tsx
 *
 * User: Caregiver
 *
 * The carepartner's 4 x 3 screen grid in the editor: the day and time, which is
 * always in the top left (`DAY_TIME_AREA`), a placeholder tile over the squares
 * each widget covers, and dashed black outlines of every square
 * drawn over the tiles, with each tile's solid edge drawn over them in turn. The
 * outlines are see-through, so they show which squares a tile covers without
 * hiding or being mistaken for its text. While a widget type is being placed
 * (`pendingType`), the squares it fits at (`openCells`) are lit and clickable.
 * Each widget has a bin (to remove it) at its top right, and a pencil (to
 * change its settings) beside it when its type has a settings panel (`hasPanel`).
 * News has no settings and reminders aren't changed in the editor, so they have
 * only the bin. The day and time has neither. Clicking a widget, or its pencil,
 * selects it, and the selected widget gets a ring. While `disabled` (the draft is
 * being saved) nothing on it can be clicked.
 */

import { DAY_TIME_AREA, GRID, openCells, widgetArea, widgetName } from "@shared";
import type { Cell, GridArea } from "@shared";
import { Pencil, Plus, Trash2 } from "lucide-react";
import type { CSSProperties } from "react";
import { FOCUS_RING } from "../components/ui/styles";
import type { EditorWidget } from "./layout";
import { placeholderText } from "./placeholders";
import type { AddableWidgetType } from "./useScreenEditor";
import WidgetIcon from "./WidgetIcon";
import { hasPanel } from "./widgetSettings";

// Places an area on the CSS grid, which counts lines from 1
function gridPosition(area: GridArea): CSSProperties {
  return {
    gridColumn: `${area.col + 1} / span ${area.width}`,
    gridRow: `${area.row + 1} / span ${area.height}`,
  };
}

const CELLS = Array.from({ length: GRID.rows * GRID.columns }, (_, i) => ({
  col: i % GRID.columns,
  row: Math.floor(i / GRID.columns),
}));

// Not one of the saved widgets: every screen has it, so it's drawn here and can't be changed
const DAY_TIME_TILE: EditorWidget = {
  id: "day_time",
  type: "day_time",
  col: DAY_TIME_AREA.col,
  row: DAY_TIME_AREA.row,
  config: {},
  photoMediaId: null,
  photoUrl: null,
};

// The pencil and bin shrink with the screen: 40px when it is large, never under 26px. The grid is a size
// container, so cqw is a share of its width; this is set on each child, since a container's own cqw
// refers to the container above it.
const TOOL_SIZE = "[--tool:clamp(26px,4.4cqw,40px)]";

const TOOL_BUTTON = `w-[var(--tool)] h-[var(--tool)] flex items-center justify-center p-0 rounded-full bg-surface border-[1.5px] border-ink-muted shadow-sm text-ink cursor-pointer disabled:opacity-50 disabled:cursor-default ${FOCUS_RING}`;

interface ScreenGridProps {
  widgets: readonly EditorWidget[];
  /** The widget being edited, drawn with a ring */
  selectedId?: EditorWidget["id"] | null;
  /** The type being placed: the squares it fits at are lit and can be clicked */
  pendingType?: AddableWidgetType | null;
  /** Nothing can be selected, changed, removed or placed, e.g. while the draft is being saved */
  disabled?: boolean;
  onPlace?: (cell: Cell) => void;
  /** A click on a widget, or its pencil */
  onSelect?: (id: EditorWidget["id"]) => void;
  /** The pencil on a widget */
  onEdit?: (id: EditorWidget["id"]) => void;
  /** The bin on a widget */
  onRemove?: (id: EditorWidget["id"]) => void;
}

export default function ScreenGrid({
  widgets,
  selectedId = null,
  pendingType = null,
  disabled = false,
  onPlace,
  onSelect,
  onEdit,
  onRemove,
}: ScreenGridProps) {
  const lit = pendingType && !disabled ? openCells(widgets, pendingType) : [];
  return (
    <div
      role="group"
      aria-label={`The carepartner's screen, ${GRID.columns} squares across and ${GRID.rows} down`}
      className="grid h-full gap-2 [container-type:inline-size]"
      style={{
        gridTemplateColumns: `repeat(${GRID.columns}, minmax(0, 1fr))`,
        gridTemplateRows: `repeat(${GRID.rows}, minmax(0, 1fr))`,
      }}
    >
      <PlaceholderTile widget={DAY_TIME_TILE} inFirstSquare />
      {widgets.map((widget) => (
        <PlaceholderTile
          key={widget.id}
          widget={widget}
          tools
          selected={widget.id === selectedId}
          disabled={disabled}
          onSelect={() => onSelect?.(widget.id)}
        />
      ))}
      {/* Over the tiles, and never in the way of a click */}
      {CELLS.map((cell) => (
        <div
          key={`${cell.col}-${cell.row}`}
          aria-hidden="true"
          className="z-20 pointer-events-none border-2 border-dashed border-ink/30 rounded-xl"
          style={gridPosition({ ...cell, width: 1, height: 1 })}
        />
      ))}
      {/* Each tile's edge, over the dashed outlines so they never show through it */}
      {[DAY_TIME_TILE, ...widgets].map((widget) => (
        <div
          key={`edge-${widget.id}`}
          aria-hidden="true"
          className={`z-30 pointer-events-none rounded-xl ring-2 ring-inset ring-ink-muted ${
            widget.id === selectedId ? "outline outline-[3px] outline-offset-[3px] outline-focus" : ""
          }`}
          style={gridPosition(widgetArea(widget))}
        />
      ))}
      {/* Above the outlines and edges, so the buttons are never drawn over; clicks pass through everywhere else */}
      {widgets.map((widget) => {
        const name = widgetName({ type: widget.type }).toLowerCase();
        return (
          <div
            key={`tools-${widget.id}`}
            className={`relative z-[35] pointer-events-none ${TOOL_SIZE}`}
            style={gridPosition(widgetArea(widget))}
          >
            <span className="absolute top-[calc(var(--tool)*0.2)] right-[calc(var(--tool)*0.2)] flex gap-[calc(var(--tool)*0.15)] pointer-events-auto">
              {hasPanel(widget.type) && (
                <button type="button" onClick={() => onEdit?.(widget.id)} disabled={disabled} aria-label={`Edit ${name}`} className={TOOL_BUTTON}>
                  <Pencil size={18} strokeWidth={2} aria-hidden="true" className="w-[45%] h-[45%]" />
                </button>
              )}
              <button type="button" onClick={() => onRemove?.(widget.id)} disabled={disabled} aria-label={`Remove ${name}`} className={TOOL_BUTTON}>
                <Trash2 size={18} strokeWidth={2} aria-hidden="true" className="w-[45%] h-[45%]" />
              </button>
            </span>
          </div>
        );
      })}
      {/* Above everything, so a lit square is always clickable */}
      {pendingType &&
        lit.map((cell) => (
          <button
            key={`lit-${cell.col}-${cell.row}`}
            type="button"
            aria-label={`Put the ${widgetName({ type: pendingType }).toLowerCase()} here, column ${cell.col + 1}, row ${cell.row + 1}`}
            onClick={() => onPlace?.(cell)}
            className={`z-40 flex items-center justify-center gap-1.5 rounded-xl border-[2.5px] border-accent bg-accent/[0.12] hover:bg-accent/25 text-[15px] font-bold text-link cursor-pointer ${FOCUS_RING}`}
            style={gridPosition({ ...cell, width: 1, height: 1 })}
          >
            <Plus size={18} strokeWidth={2.6} aria-hidden="true" />
            Place here
          </button>
        ))}
    </div>
  );
}

interface PlaceholderTileProps {
  widget: EditorWidget;
  /** Centers the words in the tile's top left square rather than across the whole tile */
  inFirstSquare?: boolean;
  /** Leaves room for the pencil and bin at the tile's top right */
  tools?: boolean;
  /** Makes the whole tile a button that selects its widget */
  selected?: boolean;
  disabled?: boolean;
  onSelect?: () => void;
}

function PlaceholderTile({
  widget,
  inFirstSquare = false,
  tools = false,
  selected = false,
  disabled = false,
  onSelect,
}: PlaceholderTileProps) {
  const { name, detail } = placeholderText(widget);
  const area = widgetArea(widget);
  // A tile one square wide stacks its icon over its words; wider tiles put them side by side
  const narrow = area.width === 1;

  const words = (
    <>
      <WidgetIcon type={widget.type} className="shrink-0 text-ink" />
      <span className="flex flex-col min-w-0">
        <span className="text-lg font-bold leading-tight">{name}</span>
        {detail && <span className="text-[15px] leading-snug text-ink-muted truncate">{detail}</span>}
      </span>
    </>
  );

  if (inFirstSquare) {
    // The tile's own squares, with the outer grid's gap, so the first one lines up with the dashed square under it
    return (
      <div
        className="relative z-10 grid gap-2 min-w-0 min-h-0 rounded-xl bg-border shadow-sm"
        style={{
          ...gridPosition(area),
          gridTemplateColumns: `repeat(${area.width}, minmax(0, 1fr))`,
          gridTemplateRows: `repeat(${area.height}, minmax(0, 1fr))`,
        }}
      >
        {/* Left aligned and centered top to bottom in the first square, and runs on to the right into the rest of the tile if it needs the room */}
        <div className="flex items-center justify-start gap-2 p-3 min-w-full w-max whitespace-nowrap">{words}</div>
      </div>
    );
  }

  // Clicking anywhere on the tile selects it, so every widget can be selected, with or without a pencil
  return (
    <button
      type="button"
      aria-pressed={selected}
      onClick={onSelect}
      disabled={disabled}
      className={`relative z-10 flex min-w-0 min-h-0 gap-2 p-3 rounded-xl bg-border shadow-sm text-left ${TOOL_SIZE} text-ink cursor-pointer disabled:cursor-default ${FOCUS_RING} ${
        narrow ? "flex-col justify-end" : `items-center ${tools ? "pr-[calc(var(--tool)*2+var(--tool)*0.15+var(--tool)*0.4)]" : ""}`
      }`}
      style={gridPosition(area)}
    >
      {words}
    </button>
  );
}
