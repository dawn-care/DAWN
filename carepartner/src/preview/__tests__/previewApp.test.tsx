/**
 * Preview mode across the app: the root layout leaves out what only a tablet
 * does, pairing comes from the address, and a preview of no carepartner never
 * opens the pair screen.
 */

import React from "react";
import { act, render, screen } from "@testing-library/react-native";
import { useKeepAwake } from "expo-keep-awake";
import RootLayout from "../../../app/_layout";
import Index from "../../../app/index";
import DevToolbar from "../../dev/DevToolbar";
import { useHiddenSystemBars } from "../../display/useHiddenSystemBars";
import { readPairedId } from "../../pairing/pairingState";
import { PlaybackProvider, usePlayback } from "../../playback/PlaybackProvider";
import PlayReports from "../../playback/PlayReports";
import StaffCommands from "../../playback/StaffCommands";
import SchedulePlayer from "../../schedule/SchedulePlayer";
import { TouchTelemetryProvider } from "../../telemetry/TouchTelemetry";
import { useAppFonts } from "../../theme/fonts";
import type { Preview } from "../previewMode";

const mockPreview: { current: Preview | null } = { current: { carepartnerId: 4 } };
jest.mock("../previewMode", () => ({
  get PREVIEW() {
    return mockPreview.current;
  },
}));

jest.mock("../../theme/fonts", () => ({ useAppFonts: jest.fn(), fonts: {} }));
jest.mock("expo-keep-awake", () => ({ useKeepAwake: jest.fn() }));
jest.mock("../../display/useHiddenSystemBars", () => ({ useHiddenSystemBars: jest.fn() }));
jest.mock("../../display/TabletDisplay", () => ({ __esModule: true, default: () => null }));
jest.mock("expo-splash-screen", () => ({ preventAutoHideAsync: jest.fn(), hideAsync: jest.fn() }));
jest.mock("expo-router", () => {
  const { Text } = jest.requireActual("react-native");
  return {
    Stack: () => <Text>app content</Text>,
    Redirect: ({ href }: { href: string }) => <Text>redirect:{href}</Text>,
  };
});
jest.mock("../../screen/ScreenProvider", () => ({
  ScreenProvider: ({ children }: { children: React.ReactNode }) => <>{children}</>,
  useScreen: () => ({ screen: null, refresh: jest.fn(), fetchedAt: null }),
}));
jest.mock("../../schedule/SchedulePlayer", () => ({ __esModule: true, default: jest.fn(() => null) }));
jest.mock("../../playback/StaffCommands", () => ({ __esModule: true, default: jest.fn(() => null) }));
jest.mock("../../playback/PlayReports", () => ({ __esModule: true, default: jest.fn(() => null) }));
jest.mock("../../telemetry/TouchTelemetry", () => ({
  TouchTelemetryProvider: jest.fn(({ children }: { children: React.ReactNode }) => <>{children}</>),
}));
jest.mock("../../playback/ChimeScreen", () => ({ __esModule: true, default: () => null }));
jest.mock("../../dev/DevToolbar", () => ({ __esModule: true, default: jest.fn(() => null) }));
jest.mock("../../pairing/pairingState", () => ({ readPairedId: jest.fn() }));
jest.mock("../../playback/MediaPlayer", () => ({ __esModule: true, default: () => null }));
jest.mock("../../display/NightScreen", () => ({ __esModule: true, default: () => null }));
jest.mock("../../screen/AlwaysOnScreen", () => ({ __esModule: true, default: () => null }));

let playback: ReturnType<typeof usePlayback>;
function Probe() {
  playback = usePlayback();
  return null;
}

describe("preview mode", () => {
  beforeEach(() => {
    jest.clearAllMocks();
    mockPreview.current = { carepartnerId: 4 };
    jest.mocked(useAppFonts).mockReturnValue([true, null]);
    jest.mocked(readPairedId).mockResolvedValue(7);
  });

  it("leaves out keeping the screen on, hiding the bars, staff commands, play reports, touch recording, the schedule and the toolbar", async () => {
    await render(<RootLayout />);
    await act(async () => {});

    expect(screen.getByText("app content")).toBeTruthy();
    expect(useKeepAwake).not.toHaveBeenCalled();
    expect(useHiddenSystemBars).not.toHaveBeenCalled();
    expect(StaffCommands).not.toHaveBeenCalled();
    expect(PlayReports).not.toHaveBeenCalled();
    expect(TouchTelemetryProvider).not.toHaveBeenCalled();
    expect(SchedulePlayer).not.toHaveBeenCalled();
    expect(DevToolbar).not.toHaveBeenCalled();
  });

  it("still does all of them on the tablet", async () => {
    mockPreview.current = null;
    await render(<RootLayout />);
    await act(async () => {});

    expect(useKeepAwake).toHaveBeenCalled();
    expect(useHiddenSystemBars).toHaveBeenCalled();
    expect(StaffCommands).toHaveBeenCalled();
    expect(PlayReports).toHaveBeenCalled();
    expect(TouchTelemetryProvider).toHaveBeenCalled();
    expect(SchedulePlayer).toHaveBeenCalled();
  });

  it("shows the carepartner its address names, without reading the stored pairing", async () => {
    await render(
      <PlaybackProvider>
        <Probe />
      </PlaybackProvider>,
    );
    await act(async () => {});

    expect(readPairedId).not.toHaveBeenCalled();
    expect(playback.state.carepartnerId).toBe(4);
    expect(playback.screen).toBe("alwaysOn");
  });

  it("shows a calm message instead of the pair screen when it names no carepartner", async () => {
    mockPreview.current = { carepartnerId: null };
    await render(
      <PlaybackProvider>
        <Index />
      </PlaybackProvider>,
    );
    await act(async () => {});

    expect(screen.getByTestId("preview-unavailable")).toBeTruthy();
    expect(screen.getByText("No screen to show.")).toBeTruthy();
    expect(screen.queryByText("redirect:/pair")).toBeNull();
  });
});
