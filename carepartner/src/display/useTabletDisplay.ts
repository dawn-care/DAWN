/**
 * useTabletDisplay.ts
 *
 * Dims the tablet's screen at night. (`app/_layout.tsx` keeps it on.)
 *
 * Brightness is set for the app's own window, not the system setting, so it
 * needs no permission and anything outside the app is unaffected. By day the
 * window follows the system brightness that staff set on the device. Only
 * Android can set it; on web the brightness is left alone.
 */

import { useEffect } from "react";
import { Platform } from "react-native";
import * as Brightness from "expo-brightness";
import { now } from "../clock/clock";
import { DEFAULT_NIGHT_HOURS, isNight, type NightHours } from "@shared";

/** Window brightness at night, from 0 to 1. Low, but still readable. */
export const NIGHT_BRIGHTNESS = 0.15;

/** How often to check whether night has started or ended. */
export const BRIGHTNESS_CHECK_MS = 60_000;

/** Dims the screen during `nightHours`, the carepartner's once the screen has them. */
export function useTabletDisplay(nightHours: NightHours = DEFAULT_NIGHT_HOURS): void {
  const { startMinute, endMinute } = nightHours;

  useEffect(() => {
    if (Platform.OS !== "android") return;

    // Whether the window is currently dimmed; null until first set, or after a
    // failure so the next check tries again. New night hours start over, so
    // they apply straight away.
    let dimmed: boolean | null = null;

    async function update() {
      const night = isNight(now(), { startMinute, endMinute });
      if (night === dimmed) return;
      dimmed = night;
      try {
        if (night) await Brightness.setBrightnessAsync(NIGHT_BRIGHTNESS);
        else await Brightness.restoreSystemBrightnessAsync();
      } catch (error) {
        dimmed = null;
        console.warn("Couldn't change the screen brightness", error);
      }
    }

    update();
    const timer = setInterval(update, BRIGHTNESS_CHECK_MS);
    return () => clearInterval(timer);
  }, [startMinute, endMinute]);
}
