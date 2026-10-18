import React from "react";
import { render, screen } from "@testing-library/react-native";
import * as SplashScreen from "expo-splash-screen";
import RootLayout from "../../../app/_layout";
import { useKeepAwake } from "expo-keep-awake";
import { useAppFonts } from "../fonts";
import { useHiddenSystemBars } from "../../display/useHiddenSystemBars";

jest.mock("../fonts", () => ({ useAppFonts: jest.fn() }));
jest.mock("expo-keep-awake", () => ({ useKeepAwake: jest.fn() }));
jest.mock("../../display/useHiddenSystemBars", () => ({ useHiddenSystemBars: jest.fn() }));
jest.mock("../../display/TabletDisplay", () => ({ __esModule: true, default: () => null }));
jest.mock("expo-splash-screen", () => ({
  preventAutoHideAsync: jest.fn(),
  hideAsync: jest.fn(),
}));
jest.mock("expo-router", () => {
  const { Text } = jest.requireActual("react-native");
  return { Stack: () => <Text>app content</Text> };
});
jest.mock("../../playback/PlaybackProvider", () => ({
  PlaybackProvider: ({ children }: { children: React.ReactNode }) => <>{children}</>,
}));
jest.mock("../../screen/ScreenProvider", () => ({
  ScreenProvider: ({ children }: { children: React.ReactNode }) => <>{children}</>,
}));
jest.mock("../../telemetry/TouchTelemetry", () => ({
  TouchTelemetryProvider: ({ children }: { children: React.ReactNode }) => <>{children}</>,
}));
jest.mock("../../schedule/SchedulePlayer", () => ({ __esModule: true, default: () => null }));
jest.mock("../../playback/StaffCommands", () => ({ __esModule: true, default: () => null }));
jest.mock("../../playback/PlayReports", () => ({ __esModule: true, default: () => null }));
jest.mock("../../dev/DevToolbar", () => ({ __esModule: true, default: () => null }));

describe("RootLayout", () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  it("keeps the splash screen up while the fonts load", async () => {
    jest.mocked(useAppFonts).mockReturnValue([false, null]);

    await render(<RootLayout />);

    expect(screen.queryByText("app content")).toBeNull();
    expect(SplashScreen.hideAsync).not.toHaveBeenCalled();
  });

  it("shows the app and hides the splash screen once the fonts load", async () => {
    jest.mocked(useAppFonts).mockReturnValue([true, null]);

    await render(<RootLayout />);

    expect(screen.getByText("app content")).toBeTruthy();
    expect(SplashScreen.hideAsync).toHaveBeenCalled();
  });

  it("keeps the screen on and hides the system bars from the start, before the fonts load", async () => {
    jest.mocked(useAppFonts).mockReturnValue([false, null]);

    await render(<RootLayout />);

    expect(useKeepAwake).toHaveBeenCalled();
    expect(useHiddenSystemBars).toHaveBeenCalled();
  });

  it("carries on with the system font if the fonts fail to load", async () => {
    jest.spyOn(console, "warn").mockImplementation(() => {});
    jest.mocked(useAppFonts).mockReturnValue([false, new Error("font missing")]);

    await render(<RootLayout />);

    expect(screen.getByText("app content")).toBeTruthy();
    expect(SplashScreen.hideAsync).toHaveBeenCalled();
  });
});
