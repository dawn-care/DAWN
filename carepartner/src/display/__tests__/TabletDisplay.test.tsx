import React from "react";
import { act, render } from "@testing-library/react-native";
import { DEFAULT_NIGHT_HOURS } from "@shared";
import { setClock } from "../../clock/clock";
import { usePlayback } from "../../playback/PlaybackProvider";
import type { Screen } from "../../screen/screen";
import { useScreen } from "../../screen/ScreenProvider";
import TabletDisplay, { NIGHT_CHECK_MS } from "../TabletDisplay";
import { useTabletDisplay } from "../useTabletDisplay";

jest.mock("../useTabletDisplay", () => ({ useTabletDisplay: jest.fn() }));
jest.mock("../../playback/PlaybackProvider", () => ({ usePlayback: jest.fn() }));
jest.mock("../../screen/ScreenProvider", () => ({ useScreen: jest.fn() }));

const setNight = jest.fn();

function withAwakeHours(awakeHours: Screen["awakeHours"] | null) {
  const screen = awakeHours ? ({ awakeHours, widgets: [], schedule: [] } as Screen) : null;
  jest.mocked(useScreen).mockReturnValue({ screen, refresh: jest.fn(), fetchedAt: null });
}

/** Whether the machine was last told it is night. */
const toldNight = () => setNight.mock.calls.at(-1)?.[0];

describe("TabletDisplay", () => {
  beforeEach(() => {
    jest.useFakeTimers();
    jest.mocked(usePlayback).mockReturnValue({ setNight } as unknown as ReturnType<typeof usePlayback>);
  });

  afterEach(() => {
    setClock(null);
    jest.useRealTimers();
    jest.clearAllMocks();
  });

  it("uses the default night hours before the first screen", async () => {
    withAwakeHours(null);
    setClock(new Date("2026-10-12T20:30:00"));
    await render(<TabletDisplay />);

    expect(useTabletDisplay).toHaveBeenLastCalledWith(DEFAULT_NIGHT_HOURS);
    expect(toldNight()).toBe(true);
  });

  it("runs night from the carepartner's bed time to their wake time", async () => {
    withAwakeHours({ wake: "10:00", bed: "21:00" });
    setClock(new Date("2026-10-12T09:44:00"));
    await render(<TabletDisplay />);

    expect(useTabletDisplay).toHaveBeenLastCalledWith({ startMinute: 21 * 60, endMinute: 10 * 60 });
    expect(toldNight()).toBe(true);
  });

  it("switches to day at wake time", async () => {
    withAwakeHours({ wake: "10:00", bed: "21:00" });
    setClock(new Date("2026-10-12T09:59:30"));
    await render(<TabletDisplay />);
    expect(toldNight()).toBe(true);

    await act(async () => jest.advanceTimersByTime(NIGHT_CHECK_MS));

    expect(toldNight()).toBe(false);
  });

  it("applies new awake hours straight away", async () => {
    withAwakeHours({ wake: "07:00", bed: "21:00" });
    setClock(new Date("2026-10-12T20:30:00"));
    const view = await render(<TabletDisplay />);
    expect(toldNight()).toBe(false);

    withAwakeHours({ wake: "07:00", bed: "20:00" });
    await view.rerender(<TabletDisplay />);

    expect(toldNight()).toBe(true);
  });
});
