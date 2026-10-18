/**
 * touchLog.ts
 *
 * Collects the resident's touches and sends them to the server in batches,
 * as feedback on where they try to interact with a screen that needs no
 * taps. Logic only.
 *
 * - A touch on the same thing, with the same gesture, within `REPEAT_WINDOW_MS`
 *   of the last one adds to its count instead of making a new event, so
 *   repeated taps read as one try and can't flood the queue.
 * - Events are sent every `FLUSH_MS`, or once `BATCH_LIMIT` are waiting,
 *   through a report queue, so ones made while offline go out later.
 *
 * Events hold structured fields only: which screen, what was under the
 * finger (a widget's type and id, an empty cell, or the kind of media
 * playing), the gesture, a count and times. Never text or a position.
 */

import { postJson } from "../api/request";
import type { ReportQueue } from "../reports/queue";
import type { Gesture } from "./gestures";

/** The screens touches are recorded on; the pair screen is for staff, so it isn't one. */
export type TouchScreen = "always_on" | "default" | "night" | "playing";

/** What was under the finger, if anything the screen can name. */
export type TouchTarget =
  | { type: "widget"; widgetType: string; widgetId: string }
  | { type: "empty" }
  | { type: "media"; mediaKind: string };

export type Touch = { screen: TouchScreen; target: TouchTarget | null; gesture: Gesture; at: Date };

export const REPEAT_WINDOW_MS = 2_000;
export const FLUSH_MS = 60_000;
export const BATCH_LIMIT = 50;

type Event = {
  key: string;
  screen: TouchScreen;
  target: TouchTarget | null;
  gesture: Gesture;
  count: number;
  firstAt: Date;
  lastAt: Date;
};

const keyOf = ({ screen, target, gesture }: Touch) => JSON.stringify([screen, target, gesture]);

/** The request body's form of an event, in the server's snake_case. */
function toWire(event: Event) {
  const { target } = event;
  return {
    screen: event.screen,
    target_type: target?.type ?? null,
    widget_type: target?.type === "widget" ? target.widgetType : null,
    widget_id: target?.type === "widget" ? target.widgetId : null,
    media_kind: target?.type === "media" ? target.mediaKind : null,
    gesture: event.gesture,
    count: event.count,
    first_at: event.firstAt.toISOString(),
    last_at: event.lastAt.toISOString(),
  };
}

export type TouchLog = {
  record: (touch: Touch) => void;
  /** Sends what's waiting now. */
  flush: () => void;
  stop: () => void;
};

export function createTouchLog(queue: ReportQueue, carepartnerId: number): TouchLog {
  let events: Event[] = [];
  let batch = 0;

  const flush = () => {
    if (!events.length) return;
    const body = { carepartner_id: carepartnerId, events: events.map(toWire) };
    events = [];
    queue.add({ group: `touches-${++batch}`, send: async () => void (await postJson("/tablet-events/touches", body)) });
  };
  const timer = setInterval(flush, FLUSH_MS);

  return {
    record(touch) {
      const key = keyOf(touch);
      const last = events[events.length - 1];
      if (last?.key === key && touch.at.getTime() - last.lastAt.getTime() <= REPEAT_WINDOW_MS) {
        last.count++;
        last.lastAt = touch.at;
        return;
      }
      events.push({ key, screen: touch.screen, target: touch.target, gesture: touch.gesture, count: 1, firstAt: touch.at, lastAt: touch.at });
      if (events.length >= BATCH_LIMIT) flush();
    },
    flush,
    stop() {
      clearInterval(timer);
      flush();
    },
  };
}
