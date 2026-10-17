import React, { useEffect } from "react";
import { Stack } from "expo-router";
import { useKeepAwake } from "expo-keep-awake";
import * as SplashScreen from "expo-splash-screen";
import TabletDisplay from "../src/display/TabletDisplay";
import { PlaybackProvider } from "../src/playback/PlaybackProvider";
import StaffCommands from "../src/playback/StaffCommands";
import SchedulePlayer from "../src/schedule/SchedulePlayer";
import { ScreenProvider } from "../src/screen/ScreenProvider";
import { useAppFonts } from "../src/theme/fonts";

SplashScreen.preventAutoHideAsync();

export default function RootLayout() {
  // Keep the screen on from the start, even while the fonts load.
  useKeepAwake();
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
      <ScreenProvider>
        <TabletDisplay />
        <StaffCommands />
        <SchedulePlayer />
        <Stack screenOptions={{ headerShown: false, animation: "none" }} />
      </ScreenProvider>
    </PlaybackProvider>
  );
}
