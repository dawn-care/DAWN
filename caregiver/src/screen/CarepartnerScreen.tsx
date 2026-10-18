/**
 * CarepartnerScreen.tsx
 *
 * User: Caregiver
 *
 * The carepartner's screen exactly as the tablet draws it: the carepartner web
 * app in an iframe, in its preview mode (see previewUrl.ts), so it shows this
 * carepartner without pairing and never acts as a tablet. Fills its container;
 * the home page and the full screen /preview page size it. Callers check
 * previewUrl first and show something else when it's null.
 *
 * "Loading carepartner screen…" covers the frame until it has loaded, and again
 * each time it loads anew (another carepartner, or `reloadKey` after a save).
 */

import { useState } from "react";

interface CarepartnerScreenProps {
  /** From previewUrl */
  src: string;
  /** Change it (e.g. after a save) to load the screen again */
  reloadKey?: number;
}

export default function CarepartnerScreen({ src, reloadKey = 0 }: CarepartnerScreenProps) {
  // A new key starts a new frame, with its own loading message
  return <Frame key={`${src}#${reloadKey}`} src={src} />;
}

function Frame({ src }: { src: string }) {
  const [loaded, setLoaded] = useState(false);

  return (
    <div className="relative w-full h-full">
      <iframe
        src={src}
        title="The carepartner's screen, as it looks on the tablet"
        onLoad={() => setLoaded(true)}
        className="block w-full h-full border-0"
      />
      {!loaded && (
        <div
          role="status"
          className="absolute inset-0 flex items-center justify-center p-6 bg-dawn-preview text-2xl font-bold text-dawn-muted text-center"
        >
          Loading carepartner screen…
        </div>
      )}
    </div>
  );
}
