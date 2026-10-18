import React, { useEffect } from "react";
import { Stack } from "expo-router";
import * as SplashScreen from "expo-splash-screen";
import DevToolbar from "../src/dev/DevToolbar";
import TabletDevice from "../src/display/TabletDevice";
import TabletDisplay from "../src/display/TabletDisplay";
import PlayReports from "../src/playback/PlayReports";
import { PlaybackProvider } from "../src/playback/PlaybackProvider";
import StaffCommands from "../src/playback/StaffCommands";
import { PREVIEW } from "../src/preview/previewMode";
import SchedulePlayer from "../src/schedule/SchedulePlayer";
import { ScreenProvider } from "../src/screen/ScreenProvider";
import { TouchTelemetryProvider } from "../src/telemetry/TouchTelemetry";
import { useAppFonts } from "../src/theme/fonts";

SplashScreen.preventAutoHideAsync();

export default function RootLayout() {
  const [fontsLoaded, fontError] = useAppFonts();
  const ready = fontsLoaded || !!fontError;

  useEffect(() => {
    if (fontError) console.warn("Fonts failed to load; using the system font", fontError);
    if (ready) SplashScreen.hideAsync();
  }, [ready, fontError]);

  // A preview (in the caregiver's iframe) only shows the screen: it doesn't keep
  // the screen on, take staff commands, play the schedule, report plays or
  // touches (the caregiver's clicks aren't the resident's), or show the toolbar.
  const isTablet = !PREVIEW;

  return (
    <>
      {/* Keeps the screen on from the start, even while the fonts load */}
      {isTablet && <TabletDevice />}
      {/* Keep the splash screen up until the fonts are ready. If they fail, carry on
          with the system font rather than leave the tablet stuck on the splash. */}
      {ready && (
        <PlaybackProvider>
          <ScreenProvider>
            <RecordTouches enabled={isTablet}>
              <TabletDisplay />
              {isTablet && <StaffCommands />}
              {isTablet && <PlayReports />}
              {isTablet && <SchedulePlayer />}
              <Stack screenOptions={{ headerShown: false, animation: "none" }} />
              {__DEV__ && isTablet && <DevToolbar />}
            </RecordTouches>
          </ScreenProvider>
        </PlaybackProvider>
      )}
    </>
  );
}

/** Records touches only on a real tablet; without the provider, touch areas record nothing. */
function RecordTouches({ enabled, children }: { enabled: boolean; children: React.ReactNode }) {
  return enabled ? <TouchTelemetryProvider>{children}</TouchTelemetryProvider> : <>{children}</>;
}
