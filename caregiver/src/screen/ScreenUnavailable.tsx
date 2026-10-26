/**
 * ScreenUnavailable.tsx
 *
 * User: Caregiver
 *
 * Shown in place of the carepartner's screen when the carepartner app's
 * address (`VITE_CAREPARTNER_URL`) isn't set, e.g. in a Docker build. Used by
 * the home page and the full screen /preview page so both say the same thing.
 */

export default function ScreenUnavailable() {
  return (
    <p className="m-0 h-full flex items-center justify-center p-6 text-lg text-ink-muted text-center">
      Carepartner screen can't be shown here.
    </p>
  );
}
