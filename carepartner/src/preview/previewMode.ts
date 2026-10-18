/**
 * previewMode.ts
 *
 * The caregiver app shows a carepartner's screen by opening this app's web
 * build in an iframe at `/?preview=1&carepartner_id=N`. In preview mode the app
 * shows that carepartner's screen as the tablet would, but doesn't act as a
 * tablet: no pairing, no keeping the screen awake, no hidden system bars, no
 * staff commands or socket, and no scheduled playback. Logic only.
 */

import { Platform } from "react-native";

/** Preview mode, with the carepartner to show (null when the address names none). */
export type Preview = { carepartnerId: number | null };

/** Reads preview mode from an address's query string, or null when it isn't a preview. */
export function parsePreview(search: string): Preview | null {
  const params = new URLSearchParams(search);
  if (params.get("preview") !== "1") return null;
  const raw = params.get("carepartner_id") ?? "";
  const id = /^\d+$/.test(raw) ? Number(raw) : NaN;
  return { carepartnerId: Number.isSafeInteger(id) && id > 0 ? id : null };
}

/**
 * Preview mode for this run of the app, read once from the address it opened
 * at. Only the web build can be a preview; the tablet never is.
 */
export const PREVIEW: Preview | null =
  Platform.OS === "web" && typeof window !== "undefined" ? parsePreview(window.location.search) : null;
