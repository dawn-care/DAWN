import { useEffect, useMemo, useRef } from "react";
import type { PlayerProps } from "./types";

/**
 * A player's `onStarted`, `onFinished` and `onFailed`, wrapped as the contract
 * in types.ts requires: started counts once and never after the end, and only
 * the first end counts (an error can follow the end, or repeat). The returned
 * functions never change and always call the latest callbacks, so a parent
 * passing new functions on every render doesn't restart a player's timers.
 */
export function useReportOnce({ onStarted, onFinished, onFailed }: Pick<PlayerProps, "onStarted" | "onFinished" | "onFailed">) {
  const callbacks = useRef({ onStarted, onFinished, onFailed });
  useEffect(() => {
    callbacks.current = { onStarted, onFinished, onFailed };
  });

  const started = useRef(false);
  const ended = useRef(false);
  return useMemo(() => {
    const once = (report: () => void) => {
      if (ended.current) return;
      ended.current = true;
      report();
    };
    return {
      started: () => {
        if (started.current || ended.current) return;
        started.current = true;
        callbacks.current.onStarted?.();
      },
      finished: () => once(() => callbacks.current.onFinished()),
      failed: (reason: string) => once(() => callbacks.current.onFailed(reason)),
    };
  }, []);
}
