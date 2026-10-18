/**
 * useSavedLayout.ts
 *
 * User: Caregiver
 *
 * Loads the carepartner's saved screen layout for the home page, and loads it
 * again when the carepartner changes or `reload` is called. A response that
 * arrives after a newer request was started is ignored.
 */

import { useCallback, useEffect, useState } from "react";
import { ApiError, friendlyError, loadWidgets } from "./api";
import { fromLayout } from "./layout";
import type { LoadedLayout } from "./layout";

export type SavedLayoutState =
  | { status: "loading" }
  | { status: "error"; message: string }
  | { status: "ready"; layout: LoadedLayout };

// The save wording ("Two widgets overlap"...) doesn't fit a failed load, so only a lost connection keeps its own
function loadErrorMessage(err: unknown): string {
  if (err instanceof ApiError && err.code === "NETWORK_ERROR") return friendlyError(err);
  return "Her screen couldn't be loaded. Please try again.";
}

export function useSavedLayout(carepartnerId: number) {
  const [state, setState] = useState<SavedLayoutState>({ status: "loading" });
  const [attempt, setAttempt] = useState(0);

  useEffect(() => {
    let current = true;
    setState({ status: "loading" });
    loadWidgets(carepartnerId).then(
      (widgets) => current && setState({ status: "ready", layout: fromLayout(widgets) }),
      (err: unknown) => current && setState({ status: "error", message: loadErrorMessage(err) }),
    );
    return () => {
      current = false;
    };
  }, [carepartnerId, attempt]);

  const reload = useCallback(() => setAttempt((n) => n + 1), []);
  return { state, reload };
}
