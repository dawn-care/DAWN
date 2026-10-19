/**
 * api.ts
 *
 * User: Caregiver
 *
 * Calls to the calendar, scheduled-item and screen-settings endpoints
 * (docs/api/screen-and-media.md), plus the types of their responses. Errors
 * from the server are thrown as ApiError with the server's stable `code`;
 * friendlyError turns any thrown value into wording for the caregiver.
 *
 * Dates and times are the carepartner's local wall time ("YYYY-MM-DD",
 * "HH:MM") and are never converted to UTC. Media itself is listed with
 * listMedia in ../media/api.
 */

import type { Repeat } from "../media/api";
import { ERROR_MESSAGES, GENERIC_ERROR } from "./constants";

const API_URL = import.meta.env.VITE_API_URL;

export type { Repeat };

/** What plays: a video, audio, audio over one photo, or a timed photo */
export type Shape = "video" | "audio" | "audio_with_background" | "photo";

export interface MediaRef {
  id: number;
  kind: "photo" | "video" | "audio";
  display_name: string;
  /** Seconds; null for photos */
  duration_seconds: number | null;
  /** Path to the file, relative to the API base */
  url: string;
}

export interface PhotoRef {
  id: number;
  display_name: string;
  url: string;
}

export interface ScreenSettings {
  carepartner_id: number;
  /** "HH:MM", 24 hour */
  wake_time: string;
  bed_time: string;
}

export interface ScheduledItem {
  id: number;
  carepartner_id: number;
  media_id: number;
  background_photo_id: number | null;
  display_seconds: number | null;
  start_date: string;
  start_time: string;
  repeat: Repeat;
  end_date: string | null;
  shape: Shape;
  media: MediaRef;
  background_photo: PhotoRef | null;
  /** Single days removed with "delete just this day" */
  skip_dates: string[];
  /** Worked out from the current awake hours on every request */
  outside_awake_hours: boolean;
}

/** One dated showing of a scheduled item, as the calendar draws it */
export interface Occurrence {
  scheduled_item_id: number;
  date: string;
  time: string;
  repeat: Repeat;
  shape: Shape;
  media: MediaRef;
  background_photo: PhotoRef | null;
  display_seconds: number | null;
  /** How long it stays on screen; 0 if unknown */
  duration_seconds: number;
  outside_awake_hours: boolean;
}

/** Body for creating or replacing a scheduled item */
export interface ItemInput {
  carepartner_id: number;
  media_id: number;
  background_photo_id?: number | null;
  display_seconds?: number | null;
  start_date: string;
  start_time: string;
  repeat: Repeat;
  end_date?: string | null;
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

// The calendar changes whenever something is saved or removed, so never reuse a stored response
const NO_CACHE: RequestInit = { cache: "no-store" };

async function api<T>(path: string, init: RequestInit = {}): Promise<T> {
  const headers = new Headers(init.headers);
  if (init.body) headers.set("Content-Type", "application/json");

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
    const code = body?.code ?? "INTERNAL_ERROR";
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

/** The page moved on and cancelled the request, e.g. to another week */
export function isAbort(err: unknown): boolean {
  return err instanceof DOMException && err.name === "AbortError";
}

/* ============================
   Calendar
============================ */

/**
 * Every dated occurrence from `from` to `to`, inclusive ("YYYY-MM-DD"), sorted
 * by date then time, with deleted days already left out. The server allows at
 * most 62 days.
 */
export function getCalendar(carepartnerId: number, from: string, to: string, signal?: AbortSignal): Promise<Occurrence[]> {
  return api<Occurrence[]>(`/calendar?carepartner_id=${carepartnerId}&from=${from}&to=${to}`, { ...NO_CACHE, signal });
}

/* ============================
   Scheduled items
============================ */

/** Every scheduled item once, as a rule. Used to find what falls outside new awake hours */
export function listScheduledItems(carepartnerId: number): Promise<ScheduledItem[]> {
  return api<ScheduledItem[]>(`/scheduled-items?carepartner_id=${carepartnerId}`, NO_CACHE);
}

export function createScheduledItem(input: ItemInput): Promise<ScheduledItem> {
  return api<ScheduledItem>("/scheduled-items", { method: "POST", body: JSON.stringify(input) });
}

/** Replaces the item; days already deleted stay deleted */
export function updateScheduledItem(id: number, input: ItemInput): Promise<ScheduledItem> {
  return api<ScheduledItem>(`/scheduled-items/${id}`, { method: "PUT", body: JSON.stringify(input) });
}

export type DeleteScope = { scope: "series" } | { scope: "occurrence"; date: string };

function deletePath(id: number, target: DeleteScope): string {
  const query = target.scope === "series" ? "scope=series" : `scope=occurrence&date=${target.date}`;
  return `/scheduled-items/${id}?${query}`;
}

/** "Every time" deletes the item; "just this day" removes one date of a repeat */
export function deleteScheduledItem(id: number, target: DeleteScope): Promise<void> {
  return api<void>(deletePath(id, target), { method: "DELETE" });
}

/**
 * Deletes as the page is being left, when a normal request would be cancelled.
 * `keepalive` lets the browser finish it after the page is gone, so there is no
 * result to wait for.
 */
export function deleteScheduledItemOnExit(id: number, target: DeleteScope): void {
  void fetch(`${API_URL}${deletePath(id, target)}`, { method: "DELETE", keepalive: true }).catch(() => {
    // Nothing can be shown once the page is closing; the item is simply still there next visit
  });
}

/* ============================
   Awake hours
============================ */

export function getScreenSettings(carepartnerId: number): Promise<ScreenSettings> {
  return api<ScreenSettings>(`/screen-settings?carepartner_id=${carepartnerId}`, NO_CACHE);
}

export function saveScreenSettings(carepartnerId: number, wakeTime: string, bedTime: string): Promise<ScreenSettings> {
  return api<ScreenSettings>(`/screen-settings?carepartner_id=${carepartnerId}`, {
    method: "PUT",
    body: JSON.stringify({ wake_time: wakeTime, bed_time: bedTime }),
  });
}

/** One item, as a rule. Opening an occurrence to edit needs its start date and end date, which occurrences leave out */
export function getScheduledItem(id: number): Promise<ScheduledItem> {
  return api<ScheduledItem>(`/scheduled-items/${id}`, NO_CACHE);
}
