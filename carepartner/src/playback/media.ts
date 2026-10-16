/**
 * media.ts
 *
 * Media items as the server sends them, and the checks that decide whether the
 * tablet can play one. This is where server data enters playback. Logic only,
 * so the schedule can ask what is playable without loading the players.
 */

/**
 * A module and its media, as the server sends it: the parts that say what to
 * play. Today these come from scheduled configurations; staff and tap plays
 * will use the same shape.
 */
export type MediaItem = {
  module_type: string;
  /**
   * Seconds to show a photo. It comes from the schedule entry, so a play that
   * isn't scheduled may have none and gets the default.
   */
  duration?: number | null;
  /**
   * The module's media details, unchecked server JSON (possibly null), so it
   * is checked once in `readMetadata` before use. It can be typed precisely
   * once responses are checked where they are fetched (#287).
   */
  metadata: unknown;
};

/**
 * The server doesn't yet guarantee a duration above 0 (#344), so a missing or
 * non-positive one gets the caregiver form's default. Remove once #344 lands.
 */
export const DEFAULT_DURATION_SECONDS = 30;

const PLAYABLE_TYPES = ["video", "audio", "photo"];

/** Whether `MediaPlayer` can play this item. */
export function isPlayable(item: MediaItem): boolean {
  return PLAYABLE_TYPES.includes(item.module_type) && readMetadata(item.metadata) !== null;
}

/** The media metadata fields playback uses, or null if there's no usable id. */
export function readMetadata(
  value: unknown,
): { id: number; url?: string; displayedName?: string; coverUri?: string } | null {
  if (typeof value !== "object" || value === null) return null;
  const { id, url, displayed_name, cover_uri } = value as Record<string, unknown>;
  if (typeof id !== "number") return null;
  return {
    id,
    url: typeof url === "string" ? url : undefined,
    displayedName: typeof displayed_name === "string" ? displayed_name : undefined,
    coverUri: typeof cover_uri === "string" ? cover_uri : undefined,
  };
}
