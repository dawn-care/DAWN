/**
 * Toast.tsx
 *
 * User: Caregiver
 *
 * Status toast with an optional action (e.g. Undo) and an optional
 * onDismiss, which runs when the toast closes without its action being used
 * (timeout, its X, or dismiss()).
 *
 *  - ToastProvider holds the toast stack and provides useToast(). Mount it
 *    once at the app root so any page can show a toast.
 *  - ToastViewport renders it, top center of the area it is placed in.
 *    Each layout renders one: AppShell puts it as the first child of <main>
 *    so it centers over the page content, clear of the nav bar; pages
 *    without the nav bar get one from ToastLayout. It stays visible while
 *    scrolling.
 *
 * Toasts stack: a new toast waits behind the one on top, and its timer
 * does not start until it reaches the top. So every toast gets its full
 * time on screen, and an Undo is never skipped over. Only the top toast
 * is clickable; the ones behind it show as edges of cards.
 *
 * Timing: the top toast's timer pauses while the pointer is over it or
 * focus is inside it, and resumes with the time it had left (WCAG 2.2.1).
 *
 * Screen readers: a visually hidden role="status" region that stays mounted
 * holds only the top toast's message, so the buttons are not read out with
 * it. The visible card itself has no live-region role.
 *
 * Focus: when the user closes a toast with its X or action button, focus
 * goes back to whatever had it when the toast was shown (or <main> if that
 * is gone). A toast that times out never moves focus.
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
import type { DismissOptions, ToastApi, ToastOptions, ToastState, ToastViewportState } from "./toastContext";
import { FOCUS_RING } from "./styles";

// Long enough to read the message and reach an Undo
const TOAST_DURATION_MS = 12000;

// How many waiting toasts are drawn behind the top one
const MAX_PEEKING = 2;

const CARD_CLASSES =
  "col-start-1 row-start-1 bg-dawn-avatar border-[1.5px] border-dawn-control rounded-[10px] shadow-toast";

/* ============================
   Helpers
============================ */

/**
 * Moves focus back to where it was before the toast, falling back to <main>
 * if that element has since left the page.
 */
function restoreFocus(target: HTMLElement | null) {
  const next =
    target && document.body.contains(target) ? target : document.querySelector<HTMLElement>("main");
  next?.focus();
}

/* ============================
   Provider
============================ */

export function ToastProvider({ children }: { children: ReactNode }) {
  const [toasts, setToasts] = useState<ToastState[]>([]);
  const [paused, setPaused] = useState(false);
  const nextId = useRef(0);
  // Which toast the timer is counting down for, and how much time it has left
  const timedId = useRef<number | undefined>(undefined);
  const remaining = useRef(TOAST_DURATION_MS);
  // The same stack as `toasts`, kept current as soon as it changes, so dismiss
  // can tell which toast it is closing (and close each only once) without
  // waiting for a render
  const stack = useRef<ToastState[]>([]);
  const topId = toasts[0]?.id;

  const dismiss = useCallback((id?: number, options: DismissOptions = {}) => {
    const target = id === undefined ? stack.current[0] : stack.current.find((t) => t.id === id);
    if (!target) return;
    stack.current = stack.current.filter((t) => t.id !== target.id);
    setToasts(stack.current);
    if (!options.actionTaken) {
      try {
        target.onDismiss?.();
      } catch (err) {
        // The toast is already closed; a failing callback must not become an uncaught error in a timer or click
        console.error("A toast's onDismiss threw", err);
      }
    }
  }, []);

  const show = useCallback((text: string, options: ToastOptions = {}) => {
    nextId.current += 1;
    const id = nextId.current;
    const active = document.activeElement;
    const returnFocus =
      active instanceof HTMLElement && active !== document.body ? active : null;
    stack.current = [...stack.current, { id, text, returnFocus, ...options }];
    setToasts(stack.current);
    return id;
  }, []);

  // The timer belongs to whichever toast is on top. A new top toast starts
  // with the full time; pausing keeps what was left and resuming uses it.
  // Cleanup also runs on unmount, so no timer outlives the provider.
  useEffect(() => {
    if (topId === undefined) {
      timedId.current = undefined;
      return;
    }
    if (timedId.current !== topId) {
      timedId.current = topId;
      remaining.current = TOAST_DURATION_MS;
    }
    if (paused) return;

    const startedAt = Date.now();
    const timer = setTimeout(() => dismiss(topId), remaining.current);
    return () => {
      clearTimeout(timer);
      remaining.current -= Date.now() - startedAt;
    };
  }, [topId, paused, dismiss]);

  const api = useMemo<ToastApi>(() => ({ show, dismiss }), [show, dismiss]);
  const viewportState = useMemo<ToastViewportState>(() => ({ toasts, setPaused }), [toasts]);

  return (
    <ToastContext.Provider value={api}>
      <ToastStateContext.Provider value={viewportState}>{children}</ToastStateContext.Provider>
    </ToastContext.Provider>
  );
}

/* ============================
   Viewport
============================ */

export function ToastViewport() {
  const { toasts, setPaused } = useContext(ToastStateContext);
  const { dismiss } = useToast();
  const topCard = useRef<HTMLDivElement>(null);
  // The toast the pointer is over / that has focus inside it. Stored by id so
  // a stale value from a toast that is gone can never pause the next one.
  const [hoverId, setHoverId] = useState<number | null>(null);
  const [focusId, setFocusId] = useState<number | null>(null);

  const top = toasts[0];
  const shown = toasts.slice(0, MAX_PEEKING + 1);
  const paused = top !== undefined && (hoverId === top.id || focusId === top.id);

  useEffect(() => {
    setPaused(paused);
  }, [paused, setPaused]);

  // Close from the toast's own buttons: give focus back if it was inside the
  // toast, so keyboard users don't restart from the top of the page
  // Closing with the action button is not a dismissal, so onDismiss is skipped
  const close = (toast: ToastState, actionTaken: boolean, afterClose?: () => void) => {
    const focusWasInside = topCard.current?.contains(document.activeElement) ?? false;
    dismiss(toast.id, { actionTaken });
    if (focusWasInside) restoreFocus(toast.returnFocus);
    afterClose?.();
  };

  return (
    // Zero-height sticky strip: the toasts hang below it, centered, and
    // stay at the top of the viewport while the page scrolls
    <div className="sticky top-4 z-50 h-0 flex justify-center pointer-events-none">
      {/* Announces only the message; keyed so a repeated message is read again */}
      <div role="status" className="sr-only">
        {top && <span key={top.id}>{top.text}</span>}
      </div>

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
              ref={topCard}
              style={{ zIndex: shown.length }}
              onMouseOver={() => setHoverId(toast.id)}
              onMouseLeave={() => setHoverId(null)}
              onFocus={() => setFocusId(toast.id)}
              onBlur={(e) => {
                // Moving between the toast's own buttons doesn't count as leaving
                if (!e.currentTarget.contains(e.relatedTarget)) setFocusId(null);
              }}
              className={`${CARD_CLASSES} flex items-center gap-3.5 py-2.5 pl-4 pr-3 text-[15px] leading-snug text-dawn-ink`}
            >
              <span>{toast.text}</span>
              {toast.actionLabel && (
                // Close first, so a toast the action shows (e.g. "Restored") isn't closed with it
                <button
                  type="button"
                  onClick={() => close(toast, true, toast.onAction)}
                  className={`shrink-0 h-9 px-2.5 bg-transparent text-[15px] font-bold text-dawn-accent hover:text-dawn-accent-hover cursor-pointer ${FOCUS_RING}`}
                >
                  {toast.actionLabel}
                </button>
              )}
              <IconButton
                label="Dismiss"
                icon={<X size={18} strokeWidth={2.4} aria-hidden="true" />}
                onClick={() => close(toast, false)}
                className="-ml-1.5 w-9 h-9"
              />
            </div>
          );
        })}
      </div>
    </div>
  );
}
