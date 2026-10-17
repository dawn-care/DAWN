/**
 * Shared widget types. Every widget on the always-on screen has the same outer
 * shape: where it sits on the grid, which renderer draws it, and that
 * renderer's data. The server decides placement; the tablet only draws it.
 */

import type { GridArea } from "@shared";

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

/** A caregiver's short message, with an optional photo and a role such as "Nurse". */
export type MessageData = {
  text: string;
  label?: string | null;
  photoUri?: string | null;
};

/** One reminder: a time of day, a few words and an icon key (see `reminderIcon`). */
export type ReminderItem = {
  /** Minutes after midnight, for picking the ones still to come. */
  minute: number;
  /** For example "8:30 AM". */
  timeLabel: string;
  text: string;
  icon: string;
};

/** The day's reminders, in time order. */
export type ReminderData = { items: ReminderItem[] };

/** One news headline. */
export type HeadlineData = { title: string };

type Placed<R extends string, D> = GridArea & {
  id: string;
  renderer: R;
  data: D;
};

export type WidgetPlacement =
  | Placed<"orientation", OrientationData>
  | Placed<"photo", PhotoData>
  | Placed<"tile", TileData>
  | Placed<"message", MessageData>
  | Placed<"reminder", ReminderData>
  | Placed<"headline", HeadlineData>;

export type RendererName = WidgetPlacement["renderer"];
