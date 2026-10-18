/**
 * api.ts
 *
 * User: Caregiver
 *
 * Calls to the screen layout endpoints (docs/api/screen-and-media.md, "Screen
 * layout"), plus the types of their requests and responses. The home page loads
 * the carepartner's widgets with loadWidgets and saves the whole list at once
 * with saveWidgets when the caregiver presses Save changes. Errors from the
 * server are thrown as ApiError with the server's stable `code`; friendlyError
 * turns any thrown value into wording for the caregiver.
 *
 * Which carepartner a request is for is passed in by the caller (from
 * `activeCarePartnerId`), as the server expects it in the query.
 */

import type { GridCell } from "@shared";

const API_URL = import.meta.env.VITE_API_URL;

/* ============================
   Types
============================ */

/*
 * Settings for each widget type the editor can add or change (the "config"
 * object); the server checks them. Reminders are out of scope for the editor: a
 * reminder already on the screen keeps its config and is sent back unchanged.
 */
export interface PhotoConfig {
  caption?: string;
}
export interface MessageConfig {
  text: string;
  location?: string;
  about?: "caregiver" | "carepartner";
  /** A role under the name, e.g. "Nurse" */
  label?: string;
}
export interface WeatherConfig {
  /** 5 digit US zip code */
  zip: string;
}
export interface SportsConfig {
  league: "nfl" | "nba" | "mlb" | "nhl";
  /** The team's abbreviation, e.g. "NYY" */
  team: string;
}
export type NewsConfig = Record<string, never>;
/** The server doesn't have this type yet (#366) */
export interface DayTimeConfig {
  show_date?: boolean;
}

export type WidgetConfig =
  | PhotoConfig
  | MessageConfig
  | WeatherConfig
  | SportsConfig
  | NewsConfig
  | DayTimeConfig;

/**
 * A widget as the server stores it. `type` is a plain string because a newer
 * server may send a type this app doesn't know; split those off with
 * splitKnownWidgets from @shared and send them back unchanged.
 */
export interface LayoutWidget {
  id: number;
  carepartner_id: number;
  type: string;
  /** Both corners are inclusive; convert with cornersToArea from @shared */
  top_left: GridCell;
  bottom_right: GridCell;
  config: WidgetConfig;
  photo_media_id: number | null;
  /** Path to the photo, relative to the API base; use photoSrc() */
  photo_url: string | null;
}

/** One widget of a save. Leave out `id` to create a widget; keep it to update that one. */
export interface LayoutWidgetInput {
  id?: number;
  type: string;
  top_left: GridCell;
  bottom_right: GridCell;
  config: WidgetConfig;
  photo_media_id?: number | null;
}

interface Layout {
  widgets: LayoutWidget[];
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
  const headers = new Headers(init.headers);
  if (init.body) headers.set("Content-Type", "application/json");

  let res: Response;
  try {
    res = await fetch(`${API_URL}${path}`, { ...init, headers });
  } catch (err) {
    console.error(`${init.method ?? "GET"} ${path} could not reach the server`, err);
    throw new ApiError(0, "NETWORK_ERROR", "could not reach the server");
  }

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

/* ============================
   Errors
============================ */

const GENERIC_ERROR = "Something went wrong. Please try again.";

// Wording for the codes a layout save can return (docs/api/screen-and-media.md, "Error codes")
const ERROR_MESSAGES: Readonly<Record<string, string>> = {
  NETWORK_ERROR: "Couldn't reach the server. Check your connection and try again.",
  INVALID_REQUEST: GENERIC_ERROR,
  INTERNAL_ERROR: GENERIC_ERROR,
  NOT_FOUND: "Something on her screen was changed somewhere else. Refresh to see the latest.",
  INVALID_TYPE: "One of the widgets can't be saved yet.",
  INVALID_CONFIG: "Check the settings of the highlighted widget.",
  INVALID_PHOTO: "Choose a photo from her library.",
  INVALID_POSITION: "A widget doesn't fit on her screen.",
  WIDGETS_OVERLAP: "Two widgets overlap. Move one of them.",
};

/**
 * Wording to show the caregiver for anything thrown by a call here. Unknown
 * codes and anything that isn't an ApiError get the generic message.
 */
export function friendlyError(err: unknown): string {
  if (err instanceof ApiError) return ERROR_MESSAGES[err.code] ?? GENERIC_ERROR;
  return GENERIC_ERROR;
}

/**
 * Which widget a save error is about, from the server's "widgets[2]: ..." message,
 * so the editor can highlight it. Undefined when the error names no single widget.
 */
export function failedWidgetIndex(err: unknown): number | undefined {
  if (!(err instanceof ApiError)) return undefined;
  const match = /^widgets\[(\d+)\]:/.exec(err.message);
  return match ? Number(match[1]) : undefined;
}

/* ============================
   Screen layout
============================ */

// The layout changes whenever it's saved, so never reuse a stored response
const NO_CACHE: RequestInit = { cache: "no-store" };

/** Every widget on the carepartner's screen, in reading order (top row first, then left to right) */
export async function loadWidgets(carepartnerId: number): Promise<LayoutWidget[]> {
  const layout = await api<Layout>(`/screen-layout?carepartner_id=${carepartnerId}`, NO_CACHE);
  return layout.widgets;
}

/**
 * Replaces the carepartner's whole screen with `widgets`, in one request: widgets
 * with an `id` are updated, ones without are created, and any of the
 * carepartner's widgets left out are deleted. `[]` clears the screen. Nothing is
 * saved if any widget is invalid. Returns the saved widgets in the order sent, so
 * a new widget's `id` can be read from its position.
 */
export async function saveWidgets(carepartnerId: number, widgets: LayoutWidgetInput[]): Promise<LayoutWidget[]> {
  const layout = await api<Layout>(`/screen-layout?carepartner_id=${carepartnerId}`, {
    method: "PUT",
    body: JSON.stringify({ widgets }),
  });
  return layout.widgets;
}

/** Address of a widget's photo, for <img>; null when it has none */
export function photoSrc(widget: Pick<LayoutWidget, "photo_url">): string | null {
  return widget.photo_url ? `${API_URL}${widget.photo_url}` : null;
}
