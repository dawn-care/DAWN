/**
 * Shared widget types. Every widget on the always-on screen has the same outer
 * shape: where it sits on the grid, which renderer draws it, and that
 * renderer's data. The server decides placement; the tablet only draws it.
 */

/** Day and time of day. The tablet works these out from its own clock. */
export type OrientationData = {
  /** Also show the date, for example "October 5". */
  showDate: boolean;
};

/** A large family photo with the person's name. */
export type PhotoData = {
  uri?: string | null;
  name?: string | null;
};

/** A small tile: one icon, one word, and an optional short label. */
export type TileData = {
  /** One of the icon keys in `renderers/icons.ts`; unknown keys get a plain icon. */
  icon: string;
  word: string;
  label?: string | null;
};

/** A spot on the grid, in cells. Columns and rows count from 0. */
export type GridArea = {
  col: number;
  row: number;
  width: number;
  height: number;
};

type Placed<R extends string, D> = GridArea & {
  id: string;
  renderer: R;
  data: D;
};

export type WidgetPlacement =
  | Placed<"orientation", OrientationData>
  | Placed<"photo", PhotoData>
  | Placed<"tile", TileData>;

export type RendererName = WidgetPlacement["renderer"];
