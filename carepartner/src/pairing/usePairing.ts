/**
 * usePairing.ts
 *
 * Runs pairing for the pair screen with no tap needed. Logic only.
 *
 * 1. Finds this tablet's carepartner: the stored one if the server still has
 *    it, or a new one if there is none or the server says it's gone (404).
 *    While the server can't be reached it waits and tries again, and never
 *    creates one, so a tablet that is briefly offline keeps its resident.
 * 2. Every `PAIRING_POLL_MS` it asks whether a caregiver has entered the code.
 *    Once one has, pairing is recorded and the playback machine told.
 * 3. When the code expires, or `newCode` is called, it gets a new one. The
 *    status is always checked first, so a code claimed just before it
 *    expired still completes, and no unused code is made after a claim (the
 *    server's status reads only the newest code, #371).
 */

import { useCallback, useEffect, useRef, useState } from "react";
import { now } from "../clock/clock";
import { usePlayback } from "../playback/PlaybackProvider";
import { getCarepartnerId, markPairingComplete, storeCarepartnerId } from "./carepartnerId";
import { createCarepartner, findCarepartner, requestPairingCode, type PairingCode } from "./pairingApi";
import { fetchPairingStatus } from "./pairingState";

/** How often to check for pairing, and to try again while offline. */
export const PAIRING_POLL_MS = 5_000;

export type PairingView =
  | { kind: "starting" }
  | ({ kind: "showing" } & PairingCode)
  | { kind: "offline" }
  | { kind: "paired" };

/** The stored carepartner if the server still has it, otherwise a new one. Throws while offline. */
async function carepartnerForThisTablet(): Promise<number> {
  const stored = await getCarepartnerId();
  if (stored != null && (await findCarepartner(stored))) return stored;
  const id = await createCarepartner();
  await storeCarepartnerId(id);
  return id;
}

export function usePairing(): { view: PairingView; newCode: () => void } {
  const { checkPairing } = usePlayback();
  const [view, setView] = useState<PairingView>({ kind: "starting" });
  const newCodeAsked = useRef(false);
  /** Ends the current wait early, for `newCode` and unmounting. */
  const wake = useRef<() => void>(() => {});

  useEffect(() => {
    let stopped = false;

    const pause = (ms: number) =>
      new Promise<void>((resolve) => {
        const done = () => {
          clearTimeout(timer);
          wake.current = () => {};
          resolve();
        };
        const timer = setTimeout(done, ms);
        wake.current = done;
      });

    async function run() {
      let id: number | null = null;
      while (!stopped && id == null) {
        try {
          id = await carepartnerForThisTablet();
        } catch (error) {
          console.warn("Couldn't set up this tablet's carepartner, trying again", error);
          if (!stopped) setView({ kind: "offline" });
          await pause(PAIRING_POLL_MS);
        }
      }

      let code: PairingCode | null = null;
      while (!stopped && id != null) {
        try {
          if (await fetchPairingStatus(id)) {
            await markPairingComplete();
            if (stopped) return;
            await checkPairing();
            if (!stopped) setView({ kind: "paired" });
            return;
          }
          if (!code || now() >= code.expiresAt || newCodeAsked.current) {
            newCodeAsked.current = false;
            code = await requestPairingCode(id);
          }
          if (!stopped) setView({ kind: "showing", ...code });
        } catch (error) {
          // A code still on screen keeps working if the server comes back, so
          // only a missing or expired one gives way to the offline message.
          console.warn("Couldn't check pairing, trying again", error);
          if (!stopped && (!code || now() >= code.expiresAt)) setView({ kind: "offline" });
        }
        // Wake at the code's expiry if that comes first; an already expired
        // code (the server is down) waits the usual time, not zero.
        const untilExpiry = code ? code.expiresAt.getTime() - now().getTime() : 0;
        await pause(untilExpiry > 0 ? Math.min(PAIRING_POLL_MS, untilExpiry) : PAIRING_POLL_MS);
      }
    }

    run();
    return () => {
      stopped = true;
      wake.current();
    };
  }, [checkPairing]);

  const newCode = useCallback(() => {
    newCodeAsked.current = true;
    wake.current();
  }, []);

  return { view, newCode };
}
