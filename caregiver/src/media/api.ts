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
import type { MediaKind } from "./constants";

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

async function api<T>(path: string, init: RequestInit = {}): Promise<T> {
  // A FormData body needs no Content-Type: the browser adds the multipart boundary
  const headers =
    init.body && !(init.body instanceof FormData)
      ? { "Content-Type": "application/json", ...init.headers }
      : init.headers;

  const res = await fetch(`${API_URL}${path}`, { ...init, headers });
  if (res.status === 204) return undefined as T;

  const body = await res.json().catch(() => null);
  if (!res.ok) {
    throw new ApiError(res.status, body?.code ?? "INTERNAL_ERROR", body?.error ?? res.statusText);
  }
  return body as T;
}

/**
 * Wording to show the caregiver for anything thrown by a call here. Network
 * failures and unknown codes get the generic message.
 */
export function friendlyError(err: unknown): string {
  if (err instanceof ApiError) return ERROR_MESSAGES[err.code] ?? GENERIC_ERROR;
  return GENERIC_ERROR;
}

/* ============================
   Media
============================ */

/** Newest first */
export function listMedia(carepartnerId: number): Promise<Media[]> {
  return api<Media[]>(`/media?carepartner_id=${carepartnerId}`);
}

interface UploadInput {
  carepartnerId: number;
  /** Needs a real file name: the server reads the extension to identify some formats */
  file: File;
  kind: MediaKind;
  displayName: string;
}

export function uploadMedia({ carepartnerId, file, kind, displayName }: UploadInput): Promise<Media> {
  const form = new FormData();
  form.append("file", file);
  form.append("kind", kind);
  form.append("carepartner_id", String(carepartnerId));
  form.append("display_name", displayName.trim());
  return api<Media>("/media", { method: "POST", body: form });
}

export function renameMedia(id: number, displayName: string): Promise<Media> {
  return api<Media>(`/media/${id}`, {
    method: "PATCH",
    body: JSON.stringify({ display_name: displayName.trim() }),
  });
}

/**
 * Deletes the item and its file. The server also removes what depends on it
 * (widgets, scheduled items), so warn first when `usage` is not empty.
 */
export function deleteMedia(id: number): Promise<void> {
  return api<void>(`/media/${id}`, { method: "DELETE" });
}

/** Address of the file itself, for <img>, <video> and <audio>; supports seeking */
export function fileUrl(media: Pick<Media, "url">): string {
  return `${API_URL}${media.url}`;
}

/* ============================
   Scheduled items
============================ */

/** Every scheduled item, once each. Used to say when a video or audio plays */
export function listScheduledItems(carepartnerId: number): Promise<ScheduledItem[]> {
  return api<ScheduledItem[]>(`/scheduled-items?carepartner_id=${carepartnerId}`);
}

/* ============================
   Length of a local file
============================ */

/**
 * Length in seconds of a video or audio file the caregiver picked, read from
 * its metadata in the browser. Returns null when the browser can't tell
 * (an unsupported codec, or a fresh browser recording that has no length
 * yet); the caller should then let the server check the length.
 */
export function probeDuration(file: Blob, kind: "video" | "audio"): Promise<number | null> {
  return new Promise((resolve) => {
    const url = URL.createObjectURL(file);
    const el = document.createElement(kind);
    const done = (seconds: number | null) => {
      el.removeAttribute("src");
      el.load();
      URL.revokeObjectURL(url);
      resolve(seconds);
    };

    el.preload = "metadata";
    el.onloadedmetadata = () => done(Number.isFinite(el.duration) ? el.duration : null);
    el.onerror = () => done(null);
    el.src = url;
  });
}
