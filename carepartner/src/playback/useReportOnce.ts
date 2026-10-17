import { useEffect, useMemo, useRef } from "react";
import type { PlayerProps } from "./types";

/**
 * A player's `onFinished` and `onFailed`, wrapped so only the first report
 * counts, as the contract in types.ts requires: an error can follow the end,
 * or repeat. The returned functions never change and always call the latest
 * callbacks, so a parent passing new functions on every render doesn't
 * restart a player's timers.
 */
export function useReportOnce({ onFinished, onFailed }: Pick<PlayerProps, "onFinished" | "onFailed">) {
  const callbacks = useRef({ onFinished, onFailed });
  useEffect(() => {
    callbacks.current = { onFinished, onFailed };
  });

  const ended = useRef(false);
  return useMemo(() => {
    const once = (report: () => void) => {
      if (ended.current) return;
      ended.current = true;
      report();
    };
    return {
      finished: () => once(() => callbacks.current.onFinished()),
      failed: (reason: string) => once(() => callbacks.current.onFailed(reason)),
    };
  }, []);
}
