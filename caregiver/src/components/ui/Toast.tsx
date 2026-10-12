/**
 * Toast.tsx
 *
 * User: Caregiver
 *
 * Status toast with an optional action (e.g. Undo). One toast shows at a
 * time and closes itself after a few seconds.
 *
 *  - ToastProvider holds the toast state and provides useToast().
 *  - ToastViewport renders it, top center of the area it is placed in.
 *    Put it as the first child of <main> so it centers over the page
 *    content, clear of the nav bar, and stays visible while scrolling.
 *
 * The viewport is a role="status" region that stays mounted, so screen
 * readers announce each new message. A Dismiss (X) button closes it early.
 */

import { useCallback, useContext, useEffect, useMemo, useRef, useState } from "react";
import type { ReactNode } from "react";
import { X } from "lucide-react";
import IconButton from "./IconButton";
import { ToastContext, ToastStateContext, useToast } from "./toastContext";
import type { ToastApi, ToastOptions, ToastState } from "./toastContext";
import { FOCUS_RING } from "./styles";

const TOAST_DURATION_MS = 5000;

/* ============================
   Provider
============================ */

export function ToastProvider({ children }: { children: ReactNode }) {
  const [toast, setToast] = useState<ToastState | null>(null);
  const timer = useRef<ReturnType<typeof setTimeout> | undefined>(undefined);
  const nextId = useRef(0);

  const dismiss = useCallback(() => {
    clearTimeout(timer.current);
    setToast(null);
  }, []);

  const show = useCallback((text: string, options: ToastOptions = {}) => {
    clearTimeout(timer.current);
    nextId.current += 1;
    setToast({ id: nextId.current, text, ...options });
    timer.current = setTimeout(() => setToast(null), TOAST_DURATION_MS);
  }, []);

  // Don't fire the timer after the provider unmounts
  useEffect(() => () => clearTimeout(timer.current), []);

  const api = useMemo<ToastApi>(() => ({ show, dismiss }), [show, dismiss]);

  return (
    <ToastContext.Provider value={api}>
      <ToastStateContext.Provider value={toast}>{children}</ToastStateContext.Provider>
    </ToastContext.Provider>
  );
}

/* ============================
   Viewport
============================ */

export function ToastViewport() {
  const toast = useContext(ToastStateContext);
  const { dismiss } = useToast();

  const handleAction = () => {
    toast?.onAction?.();
    dismiss();
  };

  return (
    // Zero-height sticky strip: the toast hangs below it, centered, and
    // stays at the top of the viewport while the page scrolls
    <div role="status" className="sticky top-4 z-50 h-0 flex justify-center pointer-events-none">
      {toast && (
        <div
          key={toast.id}
          className="pointer-events-auto max-w-[520px] h-fit flex items-center gap-3.5 py-2.5 pl-4 pr-3 bg-dawn-avatar border-[1.5px] border-dawn-control rounded-[10px] shadow-toast text-[15px] leading-snug text-dawn-ink"
        >
          <span>{toast.text}</span>
          {toast.actionLabel && (
            <button
              type="button"
              onClick={handleAction}
              className={`shrink-0 h-9 px-2.5 bg-transparent text-[15px] font-bold text-dawn-accent hover:text-dawn-accent-hover cursor-pointer ${FOCUS_RING}`}
            >
              {toast.actionLabel}
            </button>
          )}
          <IconButton
            label="Dismiss"
            icon={<X size={18} strokeWidth={2.4} aria-hidden="true" />}
            onClick={dismiss}
            className="-ml-1.5 w-9 h-9"
          />
        </div>
      )}
    </div>
  );
}
