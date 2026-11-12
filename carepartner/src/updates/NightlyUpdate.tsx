/**
 * NightlyUpdate.tsx
 *
 * Applies over-the-air updates (EAS Update, `expo-updates`) at night.
 * Rendered in `_layout.tsx` inside `PlaybackProvider`; draws nothing.
 *
 * `expo-updates` by itself downloads an update at launch and applies it at the
 * next launch, but the tablet runs for weeks without restarting. So while the
 * night screen shows (night, nothing playing), this checks for an update,
 * downloads it and restarts the app into it. The restart lands back on the
 * night screen. It checks again every `UPDATE_CHECK_MS` through the night, in
 * case the network was down or an update was published later.
 *
 * If something starts playing while an update downloads, the restart waits
 * for the next time the night screen shows, so a clip is never cut off.
 * Development builds, the web build and builds without an update address have
 * updates switched off (`Updates.isEnabled`), and this does nothing.
 */

import { useEffect, useRef } from "react";
import * as Updates from "expo-updates";
import { usePlayback } from "../playback/PlaybackProvider";

/** How often to check again while the night screen shows. */
export const UPDATE_CHECK_MS = 60 * 60_000;

export default function NightlyUpdate() {
  const quiet = usePlayback().screen === "night";
  const quietNow = useRef(quiet);
  quietNow.current = quiet;
  // An update has been downloaded and is waiting for a restart.
  const downloaded = useRef(false);

  useEffect(() => {
    if (!quiet || !Updates.isEnabled) return;
    let cancelled = false;
    let timer: ReturnType<typeof setTimeout> | undefined;

    const check = async () => {
      try {
        if (!downloaded.current) {
          const found = await Updates.checkForUpdateAsync();
          if (found.isAvailable || found.isRollBackToEmbedded) {
            const fetched = await Updates.fetchUpdateAsync();
            downloaded.current = fetched.isNew || fetched.isRollBackToEmbedded;
          }
        }
        if (downloaded.current) {
          if (!cancelled && quietNow.current) await Updates.reloadAsync();
          return;
        }
      } catch (error) {
        console.warn("Couldn't check for an app update, trying again later", error);
      }
      if (!cancelled) timer = setTimeout(check, UPDATE_CHECK_MS);
    };
    check();

    return () => {
      cancelled = true;
      clearTimeout(timer);
    };
  }, [quiet]);

  return null;
}
