/**
 * previewUrl.ts
 *
 * User: Caregiver
 *
 * Where the carepartner web app shows one carepartner's screen, for the home
 * page's iframe: `VITE_CAREPARTNER_URL` with `?preview=1&carepartner_id=N`.
 */

const CAREPARTNER_URL: string | undefined = import.meta.env.VITE_CAREPARTNER_URL;

/** The preview address for a carepartner, or null when the carepartner app's address isn't set */
export function previewUrl(carepartnerId: number): string | null {
  if (!CAREPARTNER_URL) return null;
  const url = new URL(CAREPARTNER_URL);
  url.searchParams.set("preview", "1");
  url.searchParams.set("carepartner_id", String(carepartnerId));
  return url.toString();
}
