import React from "react";
import { render, screen } from "@testing-library/react-native";
import * as SplashScreen from "expo-splash-screen";
import RootLayout from "../../../app/_layout";
import { useTabletDisplay } from "../../display/useTabletDisplay";
import { useAppFonts } from "../fonts";

jest.mock("../fonts", () => ({ useAppFonts: jest.fn() }));
jest.mock("../../display/useTabletDisplay", () => ({ useTabletDisplay: jest.fn() }));
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
jest.mock("../../schedule/ScheduleProvider", () => ({
  ScheduleProvider: ({ children }: { children: React.ReactNode }) => <>{children}</>,
}));

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

  it("manages the screen from the start, before the fonts load", async () => {
    jest.mocked(useAppFonts).mockReturnValue([false, null]);

    await render(<RootLayout />);

    expect(useTabletDisplay).toHaveBeenCalled();
  });

  it("carries on with the system font if the fonts fail to load", async () => {
    jest.spyOn(console, "warn").mockImplementation(() => {});
    jest.mocked(useAppFonts).mockReturnValue([false, new Error("font missing")]);

    await render(<RootLayout />);

    expect(screen.getByText("app content")).toBeTruthy();
    expect(SplashScreen.hideAsync).toHaveBeenCalled();
  });
});
