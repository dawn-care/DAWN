/**
 * staffItem.ts
 *
 * Turns a staff play command into something the tablet can play. A command
 * names a media item; `GET /media/{id}` says what kind it is and where its
 * file is (see docs/api/screen-and-media.md).
 */

import { API_URL } from "../api/config";
import type { MediaItem } from "./media";

/**
 * How long to wait for the lookup. React Native's `fetch` has no time limit of
 * its own, and a stalled lookup would leave staff waiting with nothing on screen.
 */
export const MEDIA_LOOKUP_TIMEOUT_MS = 5_000;

/**
 * The media item to play, or null if it can't be looked up in time. A photo
 * shows for the default time, since a staff play has no display time.
 */
export async function fetchMediaItem(mediaId: number): Promise<MediaItem | null> {
  const controller = new AbortController();
  const timer = setTimeout(() => controller.abort(), MEDIA_LOOKUP_TIMEOUT_MS);
  try {
    const res = await fetch(`${API_URL}/media/${mediaId}`, { signal: controller.signal });
    if (!res.ok) throw new Error(`Media request failed (${res.status})`);
    const media: unknown = await res.json();
    if (typeof media !== "object" || media === null) throw new Error("Media response is not an object");
    const { kind, url, display_name } = media as Record<string, unknown>;
    if (typeof kind !== "string" || typeof url !== "string") {
      throw new Error("Media response is missing its kind or file");
    }
    return { module_type: kind, metadata: { id: mediaId, url, displayed_name: display_name } };
  } catch (error) {
    console.warn(`Couldn't look up media ${mediaId}`, error);
    return null;
  } finally {
    clearTimeout(timer);
  }
}
