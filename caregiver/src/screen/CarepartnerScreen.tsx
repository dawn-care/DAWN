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
 *
 * If the carepartner app can't be reached (it isn't running), the browser would
 * fill the frame with its own "refused to connect" page, which an iframe gives
 * the page no way to detect. So the app is asked for first, and the frame is
 * drawn only once it answers; when it doesn't, "Could not connect to
 * carepartner tablet." is shown instead, with Try again.
 */

import { useEffect, useLayoutEffect, useRef, useState } from "react";
import { FOCUS_RING } from "../components/ui/styles";

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
  // null while checking, then whether the carepartner app answered
  const [reachable, setReachable] = useState<boolean | null>(null);
  // Counts tries, so Try again loads the frame and asks again
  const [attempt, setAttempt] = useState(0);
  const box = useRef<HTMLDivElement>(null);
  const [scale, setScale] = useState(0);

  // Ask for the page first. no-cors lets this work across origins: the answer can't be read, but it
  // only fails when nothing is listening there, which is the case this is for (the app isn't running).
  // Any answer at all counts as reachable, so an error page (404, 500), another program on that port,
  // or a page that refuses to be framed still loads in the frame as it is.
  useEffect(() => {
    let current = true;
    fetch(src, { mode: "no-cors", cache: "no-store" }).then(
      () => current && setReachable(true),
      (err: unknown) => {
        if (!current) return;
        console.warn("The carepartner app could not be reached", err);
        setReachable(false);
      },
    );
    return () => {
      current = false;
    };
  }, [src, attempt]);

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
      {/* Only once the app answered: a frame drawn sooner could show the browser's own error page, which
          also fires onLoad and would take the loading message away */}
      {reachable === true && (
        <iframe
          key={attempt}
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
      )}
      {reachable === false && (
        <div
          role="alert"
          className="absolute inset-0 flex flex-col items-center justify-center gap-4 p-6 bg-dawn-preview text-2xl font-bold text-dawn-muted text-center"
        >
          Could not connect to carepartner tablet.
          <button
            type="button"
            onClick={() => {
              setReachable(null);
              setLoaded(false);
              setAttempt((n) => n + 1);
            }}
            className={`h-11 px-[18px] rounded-[10px] border-2 border-dawn-control bg-white hover:bg-dawn-page text-dawn-ink text-base font-bold cursor-pointer ${FOCUS_RING}`}
          >
            Try again
          </button>
        </div>
      )}
      {reachable !== false && !loaded && (
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
