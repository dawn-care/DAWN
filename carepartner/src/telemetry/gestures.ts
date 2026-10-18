/**
 * gestures.ts
 *
 * What kind of touch a resident made, from where and when it started and
 * ended. Logic only.
 */

export type Gesture = "tap" | "long_press" | "swipe";

/** A touch held at least this long without moving counts as a long press. */
export const LONG_PRESS_MS = 500;
/** A touch that moves at least this far, in layout units, counts as a swipe. */
export const SWIPE_DISTANCE = 30;

export type TouchPoint = { x: number; y: number; at: number };

export function gestureOf(start: TouchPoint, end: TouchPoint): Gesture {
  if (Math.hypot(end.x - start.x, end.y - start.y) >= SWIPE_DISTANCE) return "swipe";
  return end.at - start.at >= LONG_PRESS_MS ? "long_press" : "tap";
}
