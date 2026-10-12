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
}

export interface ToastApi {
  /** Adds a toast. If one is already showing, it waits behind it in a stack */
  show: (text: string, options?: ToastOptions) => void;
  /** Closes the toast on top; the next one in the stack moves up */
  dismiss: () => void;
}

// Stable functions for showing/dismissing; components that only call show()
// don't re-render when the toasts change
export const ToastContext = createContext<ToastApi | null>(null);

// The stack of toasts, first = on top. Read only by ToastViewport
export const ToastStateContext = createContext<ToastState[]>([]);

/**
 * Returns show(text, options) and dismiss(). Must be used under ToastProvider.
 */
export function useToast(): ToastApi {
  const api = useContext(ToastContext);
  if (!api) throw new Error("useToast must be used inside ToastProvider");
  return api;
}
