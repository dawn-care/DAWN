/**
 * layoutPresets.ts
 *
 * The screen layouts a caregiver can pick, and the rules for putting widgets in
 * them. The server stores each widget's `spot` (a number counted from 0) and
 * knows only how many spots a layout has. Where each spot sits on the screen is
 * decided here, so the caregiver's preview and the tablet match.
 *
 * Any widget can go in any spot. Choosing a smaller layout never deletes a
 * widget; spots past the end are hidden and come back when the layout grows.
 */

/** The layouts the caregiver can choose from. */
export type LayoutName = "simple" | "balanced" | "grid" | "busier";

/** A layout as the server describes it: its name and how many spots it has. */
export type LayoutPreset = {
  name: string;
  spots: number;
};

/** A spot's place on its layout's grid, in cells. Columns and rows count from 0. */
export type SpotArea = {
  col: number;
  row: number;
  width: number;
  height: number;
};

/** The grid a layout draws on, and where each spot sits. Spots count from 0 in reading order. */
export type LayoutShape = {
  /** How the layout is named in the picker, for example "Simple". */
  label: string;
  /** A short description for the picker, for example "Photo + 1". */
  summary: string;
  columns: number;
  rows: number;
  /** Relative width of each column, so the first column can be wider. */
  columnWeights: readonly number[];
  /** The area for each spot, by spot number. */
  spots: readonly SpotArea[];
};

export const LAYOUTS: Readonly<Record<LayoutName, LayoutShape>> = {
  simple: {
    label: "Simple",
    summary: "Photo + 1",
    columns: 2,
    rows: 1,
    columnWeights: [1.2, 1],
    spots: [
      { col: 0, row: 0, width: 1, height: 1 },
      { col: 1, row: 0, width: 1, height: 1 },
    ],
  },
  balanced: {
    label: "Balanced",
    summary: "Photo + 3",
    columns: 2,
    rows: 3,
    columnWeights: [1.7, 1],
    spots: [
      { col: 0, row: 0, width: 1, height: 3 },
      { col: 1, row: 0, width: 1, height: 1 },
      { col: 1, row: 1, width: 1, height: 1 },
      { col: 1, row: 2, width: 1, height: 1 },
    ],
  },
  grid: {
    label: "Grid",
    summary: "4 equal",
    columns: 2,
    rows: 2,
    columnWeights: [1, 1],
    spots: [
      { col: 0, row: 0, width: 1, height: 1 },
      { col: 1, row: 0, width: 1, height: 1 },
      { col: 0, row: 1, width: 1, height: 1 },
      { col: 1, row: 1, width: 1, height: 1 },
    ],
  },
  busier: {
    label: "Busier",
    summary: "6 equal",
    columns: 3,
    rows: 2,
    columnWeights: [1, 1, 1],
    spots: [
      { col: 0, row: 0, width: 1, height: 1 },
      { col: 1, row: 0, width: 1, height: 1 },
      { col: 2, row: 0, width: 1, height: 1 },
      { col: 0, row: 1, width: 1, height: 1 },
      { col: 1, row: 1, width: 1, height: 1 },
      { col: 2, row: 1, width: 1, height: 1 },
    ],
  },
};

/**
 * The shape to draw for a server layout. Undefined when the name is unknown or
 * the server's spot count differs from the shape's, so a mismatch is never drawn
 * as a guess.
 */
export function layoutShape(preset: LayoutPreset): LayoutShape | undefined {
  if (!isLayoutName(preset.name)) return undefined;
  const shape = LAYOUTS[preset.name];
  return shape.spots.length === preset.spots ? shape : undefined;
}

function isLayoutName(name: string): name is LayoutName {
  return Object.prototype.hasOwnProperty.call(LAYOUTS, name);
}

/** Whether a spot number is one of the layout's spots. */
export function isSpotInLayout(spot: number, preset: LayoutPreset): boolean {
  return Number.isInteger(spot) && spot >= 0 && spot < preset.spots;
}

/** The lowest spot with no widget in it, or null when every spot is taken. */
export function firstEmptySpot(
  widgets: readonly { spot: number }[],
  preset: LayoutPreset,
): number | null {
  const taken = new Set(widgets.map((widget) => widget.spot));
  for (let spot = 0; spot < preset.spots; spot += 1) {
    if (!taken.has(spot)) return spot;
  }
  return null;
}

/** How a layout is named for the caregiver; an unknown name is shown as the server sent it. */
export function layoutLabel(preset: LayoutPreset): string {
  return isLayoutName(preset.name) ? LAYOUTS[preset.name].label : preset.name;
}
