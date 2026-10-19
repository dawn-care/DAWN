/**
 * CarepartnerScreen.tsx
 *
 * User: Caregiver
 *
 * The carepartner's screen exactly as the tablet draws it: the carepartner web
 * app in an iframe, in its preview mode (see previewUrl.ts), so it shows this
 * carepartner without pairing and never acts as a tablet. Callers check
 * previewUrl first and show something else when it's null.
 *
 * The iframe is laid out at the tablet's own size (TABLET_SIZE) and scaled down
 * to fit its container, keeping its shape and centered, so the tablet app lays
 * out and sizes its text just as it does on the tablet instead of squeezing it
 * into a small frame.
 *
 * "Loading carepartner screen…" covers the frame until it has loaded, and again
 * each time it loads anew (another carepartner, or `reloadKey` after a save).
 */

import { useLayoutEffect, useRef, useState } from "react";

/** The tablet's screen in layout units: 1280 x 800, the shape of the visual design */
const TABLET_SIZE = { width: 1280, height: 800 };

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
  const box = useRef<HTMLDivElement>(null);
  const [scale, setScale] = useState(0);

  // Follow the container's size, so the tablet always fits it whole
  useLayoutEffect(() => {
    const element = box.current;
    if (!element) return;
    const fit = () =>
      setScale(Math.min(element.clientWidth / TABLET_SIZE.width, element.clientHeight / TABLET_SIZE.height));
    fit();
    const observer = new ResizeObserver(fit);
    observer.observe(element);
    return () => observer.disconnect();
  }, []);

  return (
    <div ref={box} className="relative w-full h-full overflow-hidden">
      <iframe
        src={src}
        title="The carepartner's screen, as it looks on the tablet"
        onLoad={() => setLoaded(true)}
        className="absolute left-1/2 top-1/2 block border-0"
        style={{
          width: TABLET_SIZE.width,
          height: TABLET_SIZE.height,
          transform: `translate(-50%, -50%) scale(${scale})`,
        }}
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
