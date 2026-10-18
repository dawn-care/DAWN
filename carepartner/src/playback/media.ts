/**
 * media.ts
 *
 * What the tablet plays, and the check that turns the server's media into it.
 * This is where server media enters playback, from `/media/{id}` (staff plays)
 * and from the schedule in `/screen`. Logic only, so the schedule can use it
 * without loading the players.
 */

export type MediaKind = "video" | "audio" | "photo";

/** One thing to play, with everything `MediaPlayer` needs. */
export type MediaItem = {
  /** The server's media id, for reporting plays; missing only in hand-made items. */
  id?: number;
  kind: MediaKind;
  /** The file: a path relative to `API_URL` such as "/media/4/file", or a full URL. */
  url: string;
  /** Shown while audio plays. */
  title?: string;
  /** Shown while audio plays: a scheduled item's background photo. */
  coverUrl?: string;
  /** Seconds to show a photo. A staff play has none and gets the default. */
  durationSeconds?: number | null;
};

/** How long to show a photo that has no display time, such as one staff start. */
export const DEFAULT_DURATION_SECONDS = 30;

const KINDS: readonly string[] = ["video", "audio", "photo"] satisfies MediaKind[];

/**
 * The playable part of a server media object (`{kind, url, display_name}`, as
 * in `/media/{id}` and a schedule entry's `media`), or null if it isn't one
 * the tablet can play.
 */
export function readMedia(value: unknown): MediaItem | null {
  if (typeof value !== "object" || value === null) return null;
  const { id, kind, url, display_name } = value as Record<string, unknown>;
  if (typeof kind !== "string" || !KINDS.includes(kind) || typeof url !== "string" || !url) return null;
  return {
    ...(typeof id === "number" ? { id } : {}),
    kind: kind as MediaKind,
    url,
    title: typeof display_name === "string" ? display_name : undefined,
  };
}
