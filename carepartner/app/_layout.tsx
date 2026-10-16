import React, { useEffect } from "react";
import { Stack } from "expo-router";
import * as SplashScreen from "expo-splash-screen";
import { useTabletDisplay } from "../src/display/useTabletDisplay";
import { PlaybackProvider } from "../src/playback/PlaybackProvider";
import StaffCommands from "../src/playback/StaffCommands";
import { ScheduleProvider } from "../src/schedule/ScheduleProvider";
import { useAppFonts } from "../src/theme/fonts";

SplashScreen.preventAutoHideAsync();

export default function RootLayout() {
  useTabletDisplay();
  const [fontsLoaded, fontError] = useAppFonts();
  const ready = fontsLoaded || !!fontError;

  useEffect(() => {
    if (fontError) console.warn("Fonts failed to load; using the system font", fontError);
    if (ready) SplashScreen.hideAsync();
  }, [ready, fontError]);

  // Keep the splash screen up until the fonts are ready. If they fail, carry on
  // with the system font rather than leave the tablet stuck on the splash.
  if (!ready) return null;

  return (
    <PlaybackProvider>
      <ScheduleProvider>
        <StaffCommands />
        <Stack screenOptions={{ headerShown: false, animation: "none" }} />
      </ScheduleProvider>
    </PlaybackProvider>
  );
}
