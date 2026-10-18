/**
 * playedStore.ts
 *
 * Remembers which scheduled plays have started, on the device, so a play isn't
 * started again after the app restarts within its grace window. This moves to
 * the server's Playback records once scheduled plays are reported (#299).
 */

import { getStoredItem, setStoredItem } from "../storage/storage";

const STORAGE_KEY = "playedScheduledItems";

/** The played keys (`playKey`), or none if nothing is stored or it can't be read. */
export async function loadPlayed(): Promise<Set<string>> {
  try {
    const stored = JSON.parse((await getStoredItem(STORAGE_KEY)) ?? "[]");
    return new Set(Array.isArray(stored) ? stored.filter((key) => typeof key === "string") : []);
  } catch (error) {
    console.warn("Couldn't read which scheduled plays have played", error);
    return new Set();
  }
}

/** Stores the played keys, dropping days before `today` ("YYYY-MM-DD") so the list stays small. */
export async function savePlayed(played: Set<string>, today: string): Promise<void> {
  const kept = [...played].filter((key) => key.slice(key.indexOf("@") + 1) >= today);
  try {
    await setStoredItem(STORAGE_KEY, JSON.stringify(kept));
  } catch (error) {
    console.warn("Couldn't store which scheduled plays have played", error);
  }
}
