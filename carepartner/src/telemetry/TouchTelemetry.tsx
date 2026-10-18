/**
 * TouchTelemetry.tsx
 *
 * Records where the resident touches the screen, through `touchLog.ts`.
 *
 * - `TouchTelemetryProvider` (in `_layout.tsx`) keeps the touch log while the
 *   tablet is paired.
 * - `TouchArea` wraps a screen and records each touch on it when it ends.
 *   It only listens: it never claims a touch, so nothing becomes tappable and
 *   anything that does respond to taps still gets them.
 * - Touch events bubble from the touched view up through its parents, so a
 *   view inside a `TouchArea` names itself as the target by calling the
 *   function from `useMarkTouchTarget` in its own `onTouchStart`, which runs
 *   before the area's. The grid does this for each widget.
 */

import React, { createContext, useContext, useEffect, useRef, useState } from "react";
import { View, type GestureResponderEvent, type StyleProp, type ViewStyle } from "react-native";
import { now } from "../clock/clock";
import { usePlayback } from "../playback/PlaybackProvider";
import { createReportQueue } from "../reports/queue";
import { gestureOf, type TouchPoint } from "./gestures";
import { createTouchLog, type TouchLog, type TouchScreen, type TouchTarget } from "./touchLog";

const LogContext = createContext<TouchLog | null>(null);
const MarkContext = createContext<((target: TouchTarget) => void) | null>(null);

export function TouchTelemetryProvider({ children }: { children: React.ReactNode }) {
  const carepartnerId = usePlayback().state.carepartnerId;
  const [log, setLog] = useState<TouchLog | null>(null);

  useEffect(() => {
    if (carepartnerId == null) return;
    const queue = createReportQueue();
    const opened = createTouchLog(queue, carepartnerId);
    setLog(opened);
    return () => {
      opened.stop();
      queue.stop();
      setLog(null);
    };
  }, [carepartnerId]);

  return <LogContext.Provider value={log}>{children}</LogContext.Provider>;
}

/** For a view inside a `TouchArea`: names it as what was touched, from its `onTouchStart`. */
export function useMarkTouchTarget(): ((target: TouchTarget) => void) | null {
  return useContext(MarkContext);
}

type Props = {
  screen: TouchScreen;
  /** What a touch is on when no view inside names itself, such as the media playing. */
  target?: TouchTarget | null;
  style?: StyleProp<ViewStyle>;
  testID?: string;
  children: React.ReactNode;
};

type Started = TouchPoint & { finger: string; target: TouchTarget | null };

export function TouchArea({ screen, target = null, style, testID, children }: Props) {
  const log = useContext(LogContext);
  const marked = useRef<TouchTarget | null>(null);
  const started = useRef<Started | null>(null);
  const mark = useRef((touched: TouchTarget) => {
    marked.current = touched;
  }).current;

  // Only the first finger counts; a second one during it is ignored.
  const onTouchStart = (e: GestureResponderEvent) => {
    const touched = marked.current;
    marked.current = null;
    if (started.current) return;
    const { pageX, pageY, identifier } = e.nativeEvent;
    started.current = { x: pageX, y: pageY, at: now().getTime(), finger: String(identifier), target: touched };
  };

  const onTouchEnd = (e: GestureResponderEvent) => {
    const start = started.current;
    const { pageX, pageY, identifier } = e.nativeEvent;
    if (!start || String(identifier) !== start.finger) return;
    started.current = null;
    const at = now();
    log?.record({
      screen,
      target: start.target ?? target,
      gesture: gestureOf(start, { x: pageX, y: pageY, at: at.getTime() }),
      at,
    });
  };

  return (
    <MarkContext.Provider value={mark}>
      <View testID={testID} style={style} onTouchStart={onTouchStart} onTouchEnd={onTouchEnd} onTouchCancel={onTouchEnd}>
        {children}
      </View>
    </MarkContext.Provider>
  );
}
