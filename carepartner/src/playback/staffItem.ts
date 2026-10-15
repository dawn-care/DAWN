/**
 * staffItem.ts
 *
 * Turns a staff play command into something the tablet can play. A command
 * names a module, which points at the media it plays (`metadata_id` is the id
 * the media file endpoints take).
 *
 * This is the only code that knows about modules. Once the server sends media
 * ids instead (#348), this fetches `/media/{id}` and nothing else changes.
 */

import { API_URL } from "../api/config";
import type { MediaItem } from "./media";

/**
 * The item a module plays, or null if it can't be looked up. It carries only
 * the media's id, so an audio item has no title or cover and a photo shows for
 * the default time.
 */
export async function fetchModuleItem(moduleId: number): Promise<MediaItem | null> {
  try {
    const res = await fetch(`${API_URL}/modules/${moduleId}`);
    if (!res.ok) throw new Error(`Module request failed (${res.status})`);
    const module: unknown = await res.json();
    if (typeof module !== "object" || module === null) throw new Error("Module response is not an object");
    const { module_type, metadata_id } = module as Record<string, unknown>;
    if (typeof module_type !== "string" || typeof metadata_id !== "number") {
      throw new Error("Module response is missing its type or media");
    }
    return { module_type, metadata: { id: metadata_id } };
  } catch (error) {
    console.warn(`Couldn't look up module ${moduleId}`, error);
    return null;
  }
}
