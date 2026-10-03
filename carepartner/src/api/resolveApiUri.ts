import { API_URL } from "./config";

/**
 * Turns a media path from the server into a full address. The server sends
 * either a full URL or a path relative to the API, such as "/audio/cover?id=3".
 * Returns undefined for a missing or blank path.
 */
export function resolveApiUri(path: string | null | undefined): string | undefined {
  const trimmed = path?.trim();
  if (!trimmed) return undefined;
  if (/^[a-z][a-z0-9+.-]*:\/\//i.test(trimmed)) return trimmed;
  return `${API_URL}${trimmed.startsWith("/") ? "" : "/"}${trimmed}`;
}
