/**
 * staffItem.ts
 *
 * Turns a staff play command into something the tablet can play. A command
 * names a media item; `GET /media/{id}` says what kind it is and where its
 * file is (see docs/api/screen-and-media.md).
 */

import { getJson } from "../api/request";
import { readMedia, type MediaItem } from "./media";

/** How long to wait for the lookup, so staff aren't left waiting with nothing on screen. */
export const MEDIA_LOOKUP_TIMEOUT_MS = 5_000;

/**
 * The media item to play, or null if it can't be looked up in time or isn't
 * playable. A photo shows for the default time, since a staff play has no
 * display time.
 */
export async function fetchMediaItem(mediaId: number): Promise<MediaItem | null> {
  try {
    const item = readMedia(await getJson(`/media/${mediaId}`, { timeoutMs: MEDIA_LOOKUP_TIMEOUT_MS }));
    if (!item) throw new Error("Media response isn't something the tablet can play");
    return item;
  } catch (error) {
    console.warn(`Couldn't look up media ${mediaId}`, error);
    return null;
  }
}
