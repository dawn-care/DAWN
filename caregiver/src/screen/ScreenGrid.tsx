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
 * hiding or being mistaken for its text. Read-only for now.
 *
 * FUTURE IMPLEMENTATION (#333): the editing controls (add bar, lit squares,
 * pencil and bin) are added on top of the same grid.
 */

import { DAY_TIME_AREA, GRID, widgetArea } from "@shared";
import type { GridArea } from "@shared";
import type { CSSProperties } from "react";
import type { EditorWidget } from "./layout";
import { placeholderText } from "./placeholders";
import WidgetIcon from "./WidgetIcon";

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

interface ScreenGridProps {
  widgets: readonly EditorWidget[];
}

export default function ScreenGrid({ widgets }: ScreenGridProps) {
  return (
    <div
      role="group"
      aria-label={`The carepartner's screen, ${GRID.columns} squares across and ${GRID.rows} down`}
      className="grid h-full gap-2"
      style={{
        gridTemplateColumns: `repeat(${GRID.columns}, minmax(0, 1fr))`,
        gridTemplateRows: `repeat(${GRID.rows}, minmax(0, 1fr))`,
      }}
    >
      <PlaceholderTile widget={DAY_TIME_TILE} inFirstSquare />
      {widgets.map((widget) => (
        <PlaceholderTile key={widget.id} widget={widget} />
      ))}
      {/* Over the tiles, and never in the way of a click */}
      {CELLS.map((cell) => (
        <div
          key={`${cell.col}-${cell.row}`}
          aria-hidden="true"
          className="z-20 pointer-events-none border-2 border-dashed border-black/30 rounded-xl"
          style={gridPosition({ ...cell, width: 1, height: 1 })}
        />
      ))}
      {/* Each tile's edge, over the dashed outlines so they never show through it */}
      {[DAY_TIME_TILE, ...widgets].map((widget) => (
        <div
          key={`edge-${widget.id}`}
          aria-hidden="true"
          className="z-30 pointer-events-none rounded-xl ring-2 ring-inset ring-dawn-control"
          style={gridPosition(widgetArea(widget))}
        />
      ))}
    </div>
  );
}

interface PlaceholderTileProps {
  widget: EditorWidget;
  /** Centers the words in the tile's top left square rather than across the whole tile */
  inFirstSquare?: boolean;
}

function PlaceholderTile({ widget, inFirstSquare = false }: PlaceholderTileProps) {
  const { name, detail } = placeholderText(widget);
  const area = widgetArea(widget);
  // A tile one square wide stacks its icon over its words; wider tiles put them side by side
  const narrow = area.width === 1;

  const words = (
    <>
      <WidgetIcon type={widget.type} className="shrink-0 text-dawn-ink" />
      <div className="flex flex-col min-w-0">
        <span className="text-lg font-bold leading-tight">{name}</span>
        {detail && <span className="text-[15px] leading-snug text-dawn-muted truncate">{detail}</span>}
      </div>
    </>
  );

  if (inFirstSquare) {
    // The tile's own squares, with the outer grid's gap, so the first one lines up with the dashed square under it
    return (
      <div
        className="relative z-10 grid gap-2 min-w-0 min-h-0 rounded-xl bg-dawn-avatar shadow-sm"
        style={{
          ...gridPosition(area),
          gridTemplateColumns: `repeat(${area.width}, minmax(0, 1fr))`,
          gridTemplateRows: `repeat(${area.height}, minmax(0, 1fr))`,
        }}
      >
        <div className="flex items-center justify-center gap-2 p-3 min-w-0">{words}</div>
      </div>
    );
  }

  return (
    <div
      className={`relative z-10 flex min-w-0 min-h-0 gap-2 p-3 rounded-xl bg-dawn-avatar shadow-sm ${
        narrow ? "flex-col justify-end" : "items-center"
      }`}
      style={gridPosition(area)}
    >
      {words}
    </div>
  );
}
