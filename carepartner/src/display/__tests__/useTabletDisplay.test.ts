import { act, renderHook } from "@testing-library/react-native";
import { Platform } from "react-native";
import * as Brightness from "expo-brightness";
import { setClock } from "../../clock/clock";
import { BRIGHTNESS_CHECK_MS, NIGHT_BRIGHTNESS, useTabletDisplay } from "../useTabletDisplay";

jest.mock("expo-brightness", () => ({
  setBrightnessAsync: jest.fn(() => Promise.resolve()),
  restoreSystemBrightnessAsync: jest.fn(() => Promise.resolve()),
}));

async function startAt(time: string) {
  setClock(new Date(time));
  return renderHook(() => useTabletDisplay());
}

async function waitOneCheck() {
  await act(async () => {
    jest.advanceTimersByTime(BRIGHTNESS_CHECK_MS);
  });
}

describe("useTabletDisplay", () => {
  beforeEach(() => {
    jest.useFakeTimers();
    jest.clearAllMocks();
    jest.replaceProperty(Platform, "OS", "android");
  });

  afterEach(() => {
    setClock(null);
    jest.useRealTimers();
    jest.restoreAllMocks();
  });

  it("uses the system brightness by day", async () => {
    await startAt("2026-10-03T12:00:00");

    expect(Brightness.restoreSystemBrightnessAsync).toHaveBeenCalledTimes(1);
    expect(Brightness.setBrightnessAsync).not.toHaveBeenCalled();
  });

  it("dims the screen at night", async () => {
    await startAt("2026-10-03T22:00:00");

    expect(Brightness.setBrightnessAsync).toHaveBeenCalledWith(NIGHT_BRIGHTNESS);
  });

  it("dims when night starts and restores when it ends", async () => {
    await startAt("2026-10-03T19:59:30");
    expect(Brightness.restoreSystemBrightnessAsync).toHaveBeenCalledTimes(1);

    await waitOneCheck();
    expect(Brightness.setBrightnessAsync).toHaveBeenCalledWith(NIGHT_BRIGHTNESS);

    setClock(new Date("2026-10-04T06:59:30"));
    await waitOneCheck();
    expect(Brightness.restoreSystemBrightnessAsync).toHaveBeenCalledTimes(2);
  });

  it("follows the night hours it is given", async () => {
    setClock(new Date("2026-10-03T21:00:00"));
    const nightHours = { startMinute: 22 * 60, endMinute: 6 * 60 };

    await renderHook(() => useTabletDisplay(nightHours));

    expect(Brightness.restoreSystemBrightnessAsync).toHaveBeenCalledTimes(1);
    expect(Brightness.setBrightnessAsync).not.toHaveBeenCalled();
  });

  it("applies new night hours straight away", async () => {
    setClock(new Date("2026-10-03T21:00:00"));
    const { rerender } = await renderHook(
      ({ startMinute }: { startMinute: number }) =>
        useTabletDisplay({ startMinute, endMinute: 6 * 60 }),
      { initialProps: { startMinute: 22 * 60 } },
    );
    expect(Brightness.setBrightnessAsync).not.toHaveBeenCalled();

    await rerender({ startMinute: 20 * 60 + 30 });

    expect(Brightness.setBrightnessAsync).toHaveBeenCalledWith(NIGHT_BRIGHTNESS);
  });

  it("changes the brightness only when day or night changes", async () => {
    await startAt("2026-10-03T12:00:00");

    await waitOneCheck();
    await waitOneCheck();

    expect(Brightness.restoreSystemBrightnessAsync).toHaveBeenCalledTimes(1);
  });

  it("tries again on the next check if changing the brightness fails", async () => {
    jest.spyOn(console, "warn").mockImplementation(() => {});
    jest.mocked(Brightness.setBrightnessAsync).mockRejectedValueOnce(new Error("no window"));
    await startAt("2026-10-03T22:00:00");

    await waitOneCheck();

    expect(Brightness.setBrightnessAsync).toHaveBeenCalledTimes(2);
  });

  it("leaves the brightness alone on web", async () => {
    jest.replaceProperty(Platform, "OS", "web");

    await startAt("2026-10-03T22:00:00");
    await waitOneCheck();

    expect(Brightness.setBrightnessAsync).not.toHaveBeenCalled();
    expect(Brightness.restoreSystemBrightnessAsync).not.toHaveBeenCalled();
  });

  it("stops checking once unmounted", async () => {
    const { unmount } = await startAt("2026-10-03T19:59:30");
    await unmount();

    await waitOneCheck();

    expect(Brightness.setBrightnessAsync).not.toHaveBeenCalled();
  });
});
