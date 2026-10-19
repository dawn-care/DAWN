/**
 * Fetch helper for the server's API, from docs/api/screen-and-media.md.
 * Error responses become an ApiError carrying the HTTP status, so screens can
 * tell "tablet is not connected" (409) apart from a real failure.
 */

import { API_URL } from "./config";

export class ApiError extends Error {
  constructor(
    public status: number,
    public code: string,
    message: string,
  ) {
    super(message);
  }
}

/** Like `api`, but also returns the status, for endpoints where 200 and 202 differ. */
export async function apiWithStatus<T>(path: string, init: RequestInit = {}): Promise<{ status: number; body: T }> {
  const headers =
    init.body && !(init.body instanceof FormData)
      ? { "Content-Type": "application/json", ...init.headers }
      : init.headers;
  const res = await fetch(`${API_URL}${path}`, { ...init, headers });
  if (res.status === 204) return { status: 204, body: undefined as T };
  const body = await res.json().catch(() => null);
  if (!res.ok) {
    throw new ApiError(res.status, body?.code ?? "INTERNAL_ERROR", body?.error ?? res.statusText);
  }
  return { status: res.status, body: body as T };
}

export async function api<T>(path: string, init: RequestInit = {}): Promise<T> {
  return (await apiWithStatus<T>(path, init)).body;
}

/** A media `url` such as "/media/4/file", made absolute. */
export const mediaUrl = (path: string) => `${API_URL}${path}`;
