/**
 * ScreenGrid.tsx
 *
 * User: Caregiver
 *
 * The carepartner's 4 x 3 screen grid in the editor: dashed outlines for every
 * square, and a placeholder tile over the squares each widget covers. Read-only
 * for now; the editing controls (lit squares, pencil and bin) are added on top
 * of the same grid.
 */

import { GRID, widgetArea } from "@shared";
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

interface ScreenGridProps {
  widgets: readonly EditorWidget[];
}

export default function ScreenGrid({ widgets }: ScreenGridProps) {
  return (
    <div
      role="group"
      aria-label={`Her screen, ${GRID.columns} squares across and ${GRID.rows} down`}
      className="grid h-full gap-2"
      style={{
        gridTemplateColumns: `repeat(${GRID.columns}, minmax(0, 1fr))`,
        gridTemplateRows: `repeat(${GRID.rows}, minmax(0, 1fr))`,
      }}
    >
      {CELLS.map((cell) => (
        <div
          key={`${cell.col}-${cell.row}`}
          aria-hidden="true"
          className="border-[1.5px] border-dashed border-dawn-dash rounded-xl"
          style={gridPosition({ ...cell, width: 1, height: 1 })}
        />
      ))}
      {widgets.map((widget) => (
        <PlaceholderTile key={widget.id} widget={widget} />
      ))}
    </div>
  );
}

function PlaceholderTile({ widget }: { widget: EditorWidget }) {
  const { name, detail } = placeholderText(widget);
  const area = widgetArea(widget);
  // A tile one square wide stacks its icon over its words; wider tiles put them side by side
  const narrow = area.width === 1;

  return (
    <div
      className={`relative z-10 flex min-w-0 min-h-0 gap-2 p-3 rounded-xl border-2 border-dawn-tile-line bg-white ${
        narrow ? "flex-col justify-end" : "items-center"
      }`}
      style={gridPosition(area)}
    >
      <WidgetIcon type={widget.type} className="shrink-0 text-dawn-ink" />
      <div className="flex flex-col min-w-0">
        <span className="text-lg font-bold leading-tight">{name}</span>
        {detail && <span className="text-[15px] leading-snug text-dawn-muted truncate">{detail}</span>}
      </div>
    </div>
  );
}
