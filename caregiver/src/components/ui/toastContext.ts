/**
 * toastContext.ts
 *
 * User: Caregiver
 *
 * Contexts and hook for showing toasts. Kept apart from Toast.tsx so that
 * file only exports components.
 */

import { createContext, useContext } from "react";

export interface ToastOptions {
  /** Label for an optional action button, e.g. "Undo" */
  actionLabel?: string;
  /** Called when the action button is clicked; the toast then closes */
  onAction?: () => void;
}

export interface ToastState extends ToastOptions {
  id: number;
  text: string;
  /** What had focus when the toast was shown; focus returns here when the user closes it */
  returnFocus: HTMLElement | null;
}

export interface ToastApi {
  /**
   * Adds a toast and returns its id. If one is already showing, it waits
   * behind it in a stack.
   */
  show: (text: string, options?: ToastOptions) => number;
  /**
   * Closes the toast with this id, wherever it is in the stack. With no id,
   * closes the toast on top. Does nothing if that toast is already gone.
   */
  dismiss: (id?: number) => void;
}

// What ToastViewport needs beyond the API: the stack to draw, and a way to
// pause the top toast's timer while someone is hovering or focused on it
export interface ToastViewportState {
  toasts: ToastState[];
  setPaused: (paused: boolean) => void;
}

// Stable functions for showing/dismissing; components that only call show()
// don't re-render when the toasts change
export const ToastContext = createContext<ToastApi | null>(null);

// The stack of toasts, first = on top. Read only by ToastViewport
export const ToastStateContext = createContext<ToastViewportState>({
  toasts: [],
  setPaused: () => {},
});

/**
 * Returns show(text, options) and dismiss(id). Must be used under ToastProvider.
 */
export function useToast(): ToastApi {
  const api = useContext(ToastContext);
  if (!api) throw new Error("useToast must be used inside ToastProvider");
  return api;
}
