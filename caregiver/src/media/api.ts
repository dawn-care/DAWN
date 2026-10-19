/**
 * api.ts
 *
 * User: Caregiver
 *
 * Calls to the media and scheduled-item endpoints (docs/api/screen-and-media.md),
 * plus the types of their responses. Errors from the server are thrown as
 * ApiError with the server's stable `code`; friendlyError turns any thrown
 * value into wording for the caregiver.
 *
 * Which carepartner a request is for is passed in by the caller (from
 * `activeCarePartnerId`), as the server expects it in the query or body.
 */

import { ERROR_MESSAGES, GENERIC_ERROR } from "./constants";
import type { MediaKind, Situation } from "./constants";

const API_URL = import.meta.env.VITE_API_URL;

export type MediaUsage = "on_screen" | "audio_background" | "scheduled";

export interface Media {
  id: number;
  carepartner_id: number;
  kind: MediaKind;
  display_name: string;
  mime_type: string;
  size_bytes: number;
  /** Seconds; null for photos */
  duration_seconds: number | null;
  /** RFC 3339 */
  created_at: string;
  /** Where it is used; empty means saved but unused */
  usage: MediaUsage[];
  /** Path to the file, relative to the API base; use fileUrl() */
  url: string;
  /** Situations staff find it under; never empty */
  situations: Situation[];
  /** Family members it is about; empty when it isn't about anyone in particular */
  family_member_ids: number[];
}

/** Someone in the carepartner's family; a folder for grouping media, not an account */
export interface FamilyMember {
  id: number;
  carepartner_id: number;
  name: string;
  /** e.g. "Daughter"; may be empty */
  relationship: string;
  /** What staff should say or avoid when the carepartner asks for them */
  guidance_note: string | null;
}

/** How staff find a media item */
export interface MediaTags {
  /** Empty means every situation */
  situations: Situation[];
  familyMemberIds: number[];
}

export type Repeat = "once" | "daily" | "weekly" | "monthly";

// Only the fields the media page reads; the server sends more
export interface ScheduledItem {
  id: number;
  media_id: number;
  /** "YYYY-MM-DD", the carepartner's local date */
  start_date: string;
  /** "HH:MM", 24 hour */
  start_time: string;
  repeat: Repeat;
  end_date: string | null;
  /** Single days removed with "delete just this day" */
  skip_dates: string[];
}

/* ============================
   Fetch helper
============================ */

export class ApiError extends Error {
  status: number;
  code: string;

  constructor(status: number, code: string, message: string) {
    super(message);
    this.status = status;
    this.code = code;
  }
}

// Lists change whenever something is saved or removed, so never reuse a stored response
const NO_CACHE: RequestInit = { cache: "no-store" };

// The code to use when the response has no JSON body, e.g. a proxy's own error page
function codeFromStatus(status: number): string {
  return status === 413 ? "FILE_TOO_LARGE" : "INTERNAL_ERROR";
}

async function api<T>(path: string, init: RequestInit = {}): Promise<T> {
  const headers = new Headers(init.headers);
  // A FormData body needs no Content-Type: the browser adds the multipart boundary
  if (init.body && !(init.body instanceof FormData)) headers.set("Content-Type", "application/json");

  let res: Response;
  try {
    res = await fetch(`${API_URL}${path}`, { ...init, headers });
  } catch (err) {
    // A request the caller cancelled is not a failure to report
    if (err instanceof DOMException && err.name === "AbortError") throw err;
    console.error(`${init.method ?? "GET"} ${path} could not reach the server`, err);
    throw new ApiError(0, "NETWORK_ERROR", "could not reach the server");
  }
  if (res.status === 204) return undefined as T;

  const body = await res.json().catch(() => null);
  if (!res.ok) {
    const code = body?.code ?? codeFromStatus(res.status);
    if (!ERROR_MESSAGES[code]) console.error(`${init.method ?? "GET"} ${path} failed`, res.status, body);
    throw new ApiError(res.status, code, body?.error ?? res.statusText);
  }
  // A 200 that isn't JSON means we reached something other than the API (e.g. the web app's catch-all page)
  if (body === null) throw new ApiError(res.status, "INTERNAL_ERROR", "response was not JSON");
  return body as T;
}

/**
 * Wording to show the caregiver for anything thrown by a call here. Unknown
 * codes and anything that isn't an ApiError get the generic message.
 */
export function friendlyError(err: unknown): string {
  if (err instanceof ApiError) return ERROR_MESSAGES[err.code] ?? GENERIC_ERROR;
  return GENERIC_ERROR;
}

/** The caregiver cancelled a request, e.g. closed the panel during an upload */
export function isAbort(err: unknown): boolean {
  return err instanceof DOMException && err.name === "AbortError";
}

/* ============================
   Media
============================ */

/** Newest first */
export function listMedia(carepartnerId: number): Promise<Media[]> {
  return api<Media[]>(`/media?carepartner_id=${carepartnerId}`, NO_CACHE);
}

interface UploadInput {
  carepartnerId: number;
  /** Needs a real file name: the server reads the extension to identify some formats */
  file: File;
  kind: MediaKind;
  displayName: string;
  tags: MediaTags;
  /** Aborting it cancels the upload; the call then rejects (see isAbort) */
  signal?: AbortSignal;
}

export function uploadMedia({ carepartnerId, file, kind, displayName, tags, signal }: UploadInput): Promise<Media> {
  const form = new FormData();
  form.append("file", file);
  form.append("kind", kind);
  form.append("carepartner_id", String(carepartnerId));
  form.append("display_name", displayName.trim());
  // Repeated fields; none means every situation
  tags.situations.forEach((situation) => form.append("situations", situation));
  tags.familyMemberIds.forEach((id) => form.append("family_member_ids", String(id)));
  return api<Media>("/media", { method: "POST", body: form, signal });
}

/** Replaces how staff find an item; no situations means every situation */
export function setMediaTags(id: number, tags: MediaTags): Promise<Media> {
  return api<Media>(`/media/${id}/tags`, {
    method: "PUT",
    body: JSON.stringify({ situations: tags.situations, family_member_ids: tags.familyMemberIds }),
  });
}

export function renameMedia(id: number, displayName: string): Promise<Media> {
  return api<Media>(`/media/${id}`, {
    method: "PATCH",
    body: JSON.stringify({ display_name: displayName.trim() }),
  });
}

/**
 * Deletes the item and its file. The server also removes what depends on it
 * (widgets, scheduled items); the remove dialog tells the caregiver so.
 */
export function deleteMedia(id: number): Promise<void> {
  return api<void>(`/media/${id}`, { method: "DELETE" });
}

/**
 * Deletes an item as the page is being left, when a normal request would be
 * cancelled. `keepalive` lets the browser finish it after the page is gone,
 * so there is no result to wait for.
 */
export function deleteMediaOnExit(id: number): void {
  void fetch(`${API_URL}/media/${id}`, { method: "DELETE", keepalive: true }).catch(() => {
    // Nothing can be shown once the page is closing; the item is simply still there next visit
  });
}

/** Address of the file itself, for <img>, <video> and <audio>; supports seeking */
export function fileUrl(media: Pick<Media, "url">): string {
  return `${API_URL}${media.url}`;
}

/* ============================
   Family members
============================ */

/** In the order they were added */
export function listFamilyMembers(carepartnerId: number): Promise<FamilyMember[]> {
  return api<FamilyMember[]>(`/family-members?carepartner_id=${carepartnerId}`, NO_CACHE);
}

export function createFamilyMember(carepartnerId: number, name: string, relationship: string): Promise<FamilyMember> {
  return api<FamilyMember>("/family-members", {
    method: "POST",
    body: JSON.stringify({ carepartner_id: carepartnerId, name: name.trim(), relationship: relationship.trim() }),
  });
}

/* ============================
   Scheduled items
============================ */

/** Every scheduled item, once each. Used to say when a media item plays */
export function listScheduledItems(carepartnerId: number): Promise<ScheduledItem[]> {
  return api<ScheduledItem[]>(`/scheduled-items?carepartner_id=${carepartnerId}`, NO_CACHE);
}

/* ============================
   Length of a local file
============================ */

// How long to wait for a file's length before giving up on it
const PROBE_TIMEOUT_MS = 8000;

export interface MediaProbe {
  /** Length in seconds, or null if the browser couldn't tell */
  seconds: number | null;
  /** True when a file read as video has no picture, i.e. it is really audio; false if unknown */
  audioOnly: boolean;
}

const UNKNOWN_PROBE: MediaProbe = { seconds: null, audioOnly: false };

/**
 * Reads a video or audio file the caregiver picked from its metadata in the
 * browser. Both kinds are read with a <video> element so a picture-less file
 * (a voice memo saved as .webm, .ogg or .mp4) shows up as `audioOnly`.
 * Returns unknown values when the browser can't tell (an unsupported codec, a
 * file with no stored length, or one that takes too long to read); the caller
 * should then let the server check.
 */
export function probeMedia(file: Blob): Promise<MediaProbe> {
  return new Promise((resolve) => {
    const url = URL.createObjectURL(file);
    const el = document.createElement("video");
    let settled = false;
    const done = (probe: MediaProbe) => {
      if (settled) return;
      settled = true;
      window.clearTimeout(giveUp);
      el.removeAttribute("src");
      el.load();
      URL.revokeObjectURL(url);
      resolve(probe);
    };

    // Some files neither load nor fail; don't wait on them forever
    const giveUp = window.setTimeout(() => {
      console.warn("Timed out reading the length of a picked file");
      done(UNKNOWN_PROBE);
    }, PROBE_TIMEOUT_MS);

    el.preload = "metadata";
    el.onloadedmetadata = () =>
      done({
        seconds: Number.isFinite(el.duration) ? el.duration : null,
        audioOnly: el.videoWidth === 0,
      });
    el.onerror = () => done(UNKNOWN_PROBE);
    el.src = url;
  });
}
