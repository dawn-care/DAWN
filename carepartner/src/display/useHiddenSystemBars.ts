/**
 * useHiddenSystemBars.ts
 *
 * Hides Android's status bar (clock, battery, notification icons) and
 * navigation bar (back, home, recents), so the resident sees only the app.
 *
 * Staff can still swipe in from the top or bottom edge to bring a bar back for
 * a moment; it hides again by itself. Locking the tablet down fully is kiosk
 * mode (#320). Android can show the bars again while the app is in the
 * background, so they are hidden again whenever it returns. Web and iOS are
 * left alone.
 *
 * `expo-navigation-bar` is native code: an app built before it was added
 * fails with "Cannot find native module 'ExpoNavigationBar'" until it is
 * rebuilt (`npm run android`).
 */

import { useEffect } from "react";
import { AppState, Platform } from "react-native";
import * as NavigationBar from "expo-navigation-bar";
import { setStatusBarHidden } from "expo-status-bar";

function hideSystemBars() {
  setStatusBarHidden(true, "none");
  NavigationBar.setVisibilityAsync("hidden").catch((error) =>
    console.warn("Couldn't hide the navigation bar", error),
  );
}

export function useHiddenSystemBars(): void {
  useEffect(() => {
    if (Platform.OS !== "android") return;
    hideSystemBars();
    const subscription = AppState.addEventListener("change", (state) => {
      if (state === "active") hideSystemBars();
    });
    return () => subscription.remove();
  }, []);
}
