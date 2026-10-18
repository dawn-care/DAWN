import { API_URL } from "./config";

/**
 * How long a request may take by default. React Native's `fetch` has no time
 * limit of its own, so a stalled request would otherwise wait forever.
 */
export const REQUEST_TIMEOUT_MS = 10_000;

type Options = { timeoutMs?: number };

/**
 * GETs `path` (relative to `API_URL`, such as "/media/9") and returns the
 * parsed JSON, unchecked; callers check its shape where it enters the app.
 * Throws on a network error, an error status or a request that takes longer
 * than `timeoutMs`.
 */
export function getJson(path: string, options?: Options): Promise<unknown> {
  return requestJson(path, {}, options);
}

/** POSTs `body` as JSON to `path`; otherwise the same as `getJson`. */
export function postJson(path: string, body: unknown, options?: Options): Promise<unknown> {
  return requestJson(
    path,
    { method: "POST", headers: { "Content-Type": "application/json" }, body: JSON.stringify(body) },
    options,
  );
}

async function requestJson(path: string, init: RequestInit, { timeoutMs = REQUEST_TIMEOUT_MS }: Options = {}): Promise<unknown> {
  const controller = new AbortController();
  const timer = setTimeout(() => controller.abort(), timeoutMs);
  try {
    const res = await fetch(`${API_URL}${path}`, { ...init, signal: controller.signal });
    if (!res.ok) throw new Error(`Request for ${path} failed (${res.status})`);
    return await res.json();
  } finally {
    clearTimeout(timer);
  }
}
