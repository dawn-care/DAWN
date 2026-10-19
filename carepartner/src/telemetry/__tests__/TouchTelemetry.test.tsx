import React from "react";
import { Text } from "react-native";
import { act, fireEvent, render, screen } from "@testing-library/react-native";
import { setClock } from "../../clock/clock";
import { usePlayback } from "../../playback/PlaybackProvider";
import Grid from "../../widgets/Grid";
import type { WidgetPlacement } from "../../widgets/types";
import { TouchArea, TouchTelemetryProvider } from "../TouchTelemetry";
import { createTouchLog, type Touch } from "../touchLog";

jest.mock("../../playback/PlaybackProvider", () => ({ usePlayback: jest.fn() }));
jest.mock("../touchLog", () => ({ createTouchLog: jest.fn() }));

const record = jest.fn<void, [Touch]>();
const weather: WidgetPlacement = {
  id: "12",
  widgetType: "weather",
  renderer: "wideTile",
  col: 0,
  row: 0,
  width: 2,
  height: 1,
  data: { icon: "sunny", label: "Sunny", word: "48°" },
};

/** A finger going down at (x, y) and, `heldMs` later, coming up at (toX, toY). */
async function touch(on: ReturnType<typeof screen.getByTestId>, area: ReturnType<typeof screen.getByTestId>, path: { x?: number; toX?: number; heldMs?: number } = {}) {
  const { x = 100, toX = x, heldMs = 100 } = path;
  const at = (pageX: number) => ({ nativeEvent: { pageX, pageY: 100, identifier: 1 } });
  // Touch events bubble: the touched view's handler runs, then its parents'.
  await act(async () => {
    if (on !== area) fireEvent(on, "touchStart", at(x));
    fireEvent(area, "touchStart", at(x));
  });
  await act(async () => jest.advanceTimersByTime(heldMs));
  await act(async () => fireEvent(area, "touchEnd", at(toX)));
}

async function renderGrid() {
  await render(
    <TouchTelemetryProvider>
      <TouchArea testID="area" screen="always_on" target={{ type: "empty" }} style={{ flex: 1 }}>
        <Grid widgets={[weather]} initialSize={{ width: 1280, height: 800 }} />
      </TouchArea>
    </TouchTelemetryProvider>,
  );
}

describe("touch telemetry", () => {
  beforeEach(() => {
    jest.useFakeTimers();
    setClock(new Date("2026-10-18T10:00:00"));
    jest.mocked(usePlayback).mockReturnValue({ state: { carepartnerId: 7 } } as ReturnType<typeof usePlayback>);
    jest.mocked(createTouchLog).mockReturnValue({ record, flush: jest.fn(), stop: jest.fn() });
  });

  afterEach(() => {
    setClock(null);
    jest.useRealTimers();
    jest.clearAllMocks();
  });

  it("records a tap on a widget as that widget's type and id", async () => {
    await renderGrid();

    await touch(screen.getByTestId("widget-12"), screen.getByTestId("area"));

    expect(record).toHaveBeenCalledWith(
      expect.objectContaining({ screen: "always_on", target: { type: "widget", widgetType: "weather", widgetId: "12" }, gesture: "tap" }),
    );
  });

  it("records a touch outside every widget as an empty cell", async () => {
    await renderGrid();

    await touch(screen.getByTestId("area"), screen.getByTestId("area"));

    expect(record).toHaveBeenCalledWith(expect.objectContaining({ target: { type: "empty" } }));
  });

  it("records long presses and swipes", async () => {
    await renderGrid();
    const area = screen.getByTestId("area");

    await touch(area, area, { heldMs: 800 });
    await touch(area, area, { x: 100, toX: 400 });

    expect(record.mock.calls.map(([t]) => t.gesture)).toEqual(["long_press", "swipe"]);
  });

  it("doesn't carry a widget over to the next touch", async () => {
    await renderGrid();
    const area = screen.getByTestId("area");

    await touch(screen.getByTestId("widget-12"), area);
    await touch(area, area);

    expect(record.mock.calls[1][0].target).toEqual({ type: "empty" });
  });

  it("ignores a second finger during a touch", async () => {
    await renderGrid();
    const area = screen.getByTestId("area");
    const finger = (identifier: number) => ({ nativeEvent: { pageX: 100, pageY: 100, identifier } });

    await act(async () => fireEvent(area, "touchStart", finger(1)));
    await act(async () => fireEvent(area, "touchStart", finger(2)));
    await act(async () => fireEvent(area, "touchEnd", finger(2)));
    await act(async () => fireEvent(area, "touchEnd", finger(1)));

    expect(record).toHaveBeenCalledTimes(1);
  });

  it("records nothing while the tablet isn't paired", async () => {
    jest.mocked(usePlayback).mockReturnValue({ state: { carepartnerId: null } } as ReturnType<typeof usePlayback>);
    await render(
      <TouchTelemetryProvider>
        <TouchArea testID="area" screen="night">
          <Text>night</Text>
        </TouchArea>
      </TouchTelemetryProvider>,
    );

    await touch(screen.getByTestId("area"), screen.getByTestId("area"));

    expect(createTouchLog).not.toHaveBeenCalled();
    expect(record).not.toHaveBeenCalled();
  });
});
