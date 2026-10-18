/**
 * previewMode.ts
 *
 * The caregiver app shows a carepartner's screen by opening this app's web
 * build in an iframe at `/?preview=1&carepartner_id=N`. In preview mode the app
 * shows that carepartner's screen as the tablet would, but doesn't act as a
 * tablet: no pairing, no keeping the screen awake, no hidden system bars, no
 * staff commands or socket, and no scheduled playback (see "Preview mode" in
 * carepartner/CLAUDE.md for the full list). The address's names are
 * `PREVIEW_ADDRESS` in @shared, which the caregiver app builds it from. Logic only.
 */

import { PREVIEW_ADDRESS } from "@shared";
import { Platform } from "react-native";

/** Preview mode, with the carepartner to show (null when the address names none). */
export type Preview = { carepartnerId: number | null };

/** Reads preview mode from an address's query string, or null when it isn't a preview. */
export function parsePreview(search: string): Preview | null {
  const params = new URLSearchParams(search);
  if (params.get(PREVIEW_ADDRESS.flag) !== PREVIEW_ADDRESS.on) return null;
  const raw = params.get(PREVIEW_ADDRESS.carepartnerId) ?? "";
  const id = /^\d+$/.test(raw) ? Number(raw) : NaN;
  return { carepartnerId: Number.isSafeInteger(id) && id > 0 ? id : null };
}

/**
 * Preview mode for this run of the app, read once from the address it opened
 * at. Only the web build can be a preview; the tablet never is.
 */
export const PREVIEW: Preview | null =
  Platform.OS === "web" && typeof window !== "undefined" ? parsePreview(window.location.search) : null;
