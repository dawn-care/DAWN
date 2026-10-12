/**
 * Toast.tsx
 *
 * User: Caregiver
 *
 * Status toast with an optional action (e.g. Undo).
 *
 *  - ToastProvider holds the toast stack and provides useToast().
 *  - ToastViewport renders it, top center of the area it is placed in.
 *    Put it as the first child of <main> so it centers over the page
 *    content, clear of the nav bar, and stays visible while scrolling.
 *
 * Toasts stack: a new toast waits behind the one on top, and its timer
 * does not start until it reaches the top. So every toast gets its full
 * time on screen, and an Undo is never skipped over. Only the top toast
 * is read out and clickable; the ones behind it show as edges of cards.
 *
 * The viewport is a role="status" region that stays mounted, so screen
 * readers announce each toast as it reaches the top. A Dismiss (X) button
 * closes the top toast early.
 *
 * Note: while a Dialog is open, Radix hides everything outside it from
 * screen readers, so a toast shown at that moment is not announced. Show
 * toasts after the dialog closes.
 */

import { useCallback, useContext, useEffect, useMemo, useRef, useState } from "react";
import type { CSSProperties, ReactNode } from "react";
import { X } from "lucide-react";
import IconButton from "./IconButton";
import { ToastContext, ToastStateContext, useToast } from "./toastContext";
import type { ToastApi, ToastOptions, ToastState } from "./toastContext";
import { FOCUS_RING } from "./styles";

// Long enough to read the message and reach an Undo
const TOAST_DURATION_MS = 12000;

// How many waiting toasts are drawn behind the top one
const MAX_PEEKING = 2;

const CARD_CLASSES =
  "col-start-1 row-start-1 bg-dawn-avatar border-[1.5px] border-dawn-control rounded-[10px] shadow-toast";

/* ============================
   Provider
============================ */

export function ToastProvider({ children }: { children: ReactNode }) {
  const [toasts, setToasts] = useState<ToastState[]>([]);
  const nextId = useRef(0);
  const topId = toasts[0]?.id;

  const dismiss = useCallback(() => setToasts((stack) => stack.slice(1)), []);

  const show = useCallback((text: string, options: ToastOptions = {}) => {
    nextId.current += 1;
    const toast: ToastState = { id: nextId.current, text, ...options };
    setToasts((stack) => [...stack, toast]);
  }, []);

  // The timer belongs to whichever toast is on top, so it restarts each time
  // a new one reaches the top and is cleared on unmount
  useEffect(() => {
    if (topId === undefined) return;
    const timer = setTimeout(() => setToasts((stack) => stack.slice(1)), TOAST_DURATION_MS);
    return () => clearTimeout(timer);
  }, [topId]);

  const api = useMemo<ToastApi>(() => ({ show, dismiss }), [show, dismiss]);

  return (
    <ToastContext.Provider value={api}>
      <ToastStateContext.Provider value={toasts}>{children}</ToastStateContext.Provider>
    </ToastContext.Provider>
  );
}

/* ============================
   Viewport
============================ */

export function ToastViewport() {
  const toasts = useContext(ToastStateContext);
  const { dismiss } = useToast();

  const shown = toasts.slice(0, MAX_PEEKING + 1);

  const handleAction = (toast: ToastState) => {
    // Close first, so a toast the action shows (e.g. "Restored") isn't closed with it
    dismiss();
    toast.onAction?.();
  };

  return (
    // Zero-height sticky strip: the toasts hang below it, centered, and
    // stay at the top of the viewport while the page scrolls
    <div role="status" className="sticky top-4 z-50 h-0 flex justify-center pointer-events-none">
      {/* One grid cell holds the whole stack; waiting toasts sit lower and smaller */}
      <div className="pointer-events-auto max-w-[520px] h-fit grid">
        {shown.map((toast, depth) => {
          if (depth > 0) {
            const peekStyle: CSSProperties = {
              transform: `translateY(${depth * 8}px) scale(${1 - depth * 0.04})`,
              zIndex: shown.length - depth,
            };
            return <div key={toast.id} aria-hidden="true" className={CARD_CLASSES} style={peekStyle} />;
          }

          return (
            <div
              key={toast.id}
              style={{ zIndex: shown.length }}
              className={`${CARD_CLASSES} flex items-center gap-3.5 py-2.5 pl-4 pr-3 text-[15px] leading-snug text-dawn-ink`}
            >
              <span>{toast.text}</span>
              {toast.actionLabel && (
                <button
                  type="button"
                  onClick={() => handleAction(toast)}
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
          );
        })}
      </div>
    </div>
  );
}
