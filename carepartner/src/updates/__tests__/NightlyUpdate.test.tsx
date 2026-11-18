import React from "react";
import { act, render } from "@testing-library/react-native";
import * as Updates from "expo-updates";
import type { Screen } from "../../playback/machine";
import { usePlayback } from "../../playback/PlaybackProvider";
import { colors } from "../../theme/theme";
import NightlyUpdate, { RELOAD_SCREEN, UPDATE_CHECK_MS } from "../NightlyUpdate";

jest.mock("expo-updates", () => ({
  isEnabled: true,
  checkForUpdateAsync: jest.fn(),
  fetchUpdateAsync: jest.fn(),
  reloadAsync: jest.fn(),
}));
jest.mock("../../playback/PlaybackProvider", () => ({ usePlayback: jest.fn() }));

const updates = jest.mocked(Updates);
const settle = (ms = 0) => act(() => jest.advanceTimersByTimeAsync(ms));

let rerender: (element: React.ReactElement) => Promise<void>;
async function showScreen(screen: Screen) {
  jest.mocked(usePlayback).mockReturnValue({ screen } as ReturnType<typeof usePlayback>);
  if (rerender) await rerender(<NightlyUpdate />);
  else ({ rerender } = await render(<NightlyUpdate />));
  await settle();
}

function updateAvailable() {
  updates.checkForUpdateAsync.mockResolvedValue({ isAvailable: true, isRollBackToEmbedded: false } as never);
  updates.fetchUpdateAsync.mockResolvedValue({ isNew: true, isRollBackToEmbedded: false } as never);
}

describe("NightlyUpdate", () => {
  beforeEach(() => {
    jest.useFakeTimers();
    jest.spyOn(console, "warn").mockImplementation(() => {});
    (updates as { isEnabled: boolean }).isEnabled = true;
    updates.checkForUpdateAsync.mockResolvedValue({ isAvailable: false, isRollBackToEmbedded: false } as never);
    updates.reloadAsync.mockResolvedValue();
    rerender = undefined as never;
  });

  afterEach(() => {
    jest.useRealTimers();
    jest.restoreAllMocks();
    jest.clearAllMocks();
  });

  it.each<Screen>(["alwaysOn", "playing", "unpaired", "starting"])("doesn't check on the %s screen", async (screen) => {
    updateAvailable();
    await showScreen(screen);
    await settle(UPDATE_CHECK_MS * 3);

    expect(updates.checkForUpdateAsync).not.toHaveBeenCalled();
  });

  it("downloads an update at night and restarts into it", async () => {
    updateAvailable();
    await showScreen("night");

    expect(updates.fetchUpdateAsync).toHaveBeenCalledTimes(1);
    expect(updates.reloadAsync).toHaveBeenCalledTimes(1);
  });

  it("restarts on a dark screen with no spinner, not the default white one", async () => {
    updateAvailable();
    await showScreen("night");

    expect(updates.reloadAsync).toHaveBeenCalledWith({ reloadScreenOptions: RELOAD_SCREEN });
    expect(RELOAD_SCREEN).toEqual({ backgroundColor: colors.night.background, spinner: { enabled: false }, fade: false });
  });

  it("restarts to roll back to the build's own code", async () => {
    updates.checkForUpdateAsync.mockResolvedValue({ isAvailable: false, isRollBackToEmbedded: true } as never);
    updates.fetchUpdateAsync.mockResolvedValue({ isNew: false, isRollBackToEmbedded: true } as never);
    await showScreen("night");

    expect(updates.reloadAsync).toHaveBeenCalledTimes(1);
  });

  it("checks again every hour through the night", async () => {
    await showScreen("night");
    expect(updates.checkForUpdateAsync).toHaveBeenCalledTimes(1);

    await settle(UPDATE_CHECK_MS);
    expect(updates.checkForUpdateAsync).toHaveBeenCalledTimes(2);

    updateAvailable();
    await settle(UPDATE_CHECK_MS);
    expect(updates.reloadAsync).toHaveBeenCalledTimes(1);
  });

  it("stops checking when the night ends", async () => {
    await showScreen("night");
    await showScreen("alwaysOn");
    await settle(UPDATE_CHECK_MS * 3);

    expect(updates.checkForUpdateAsync).toHaveBeenCalledTimes(1);
  });

  it("tries again later when the check fails", async () => {
    updates.checkForUpdateAsync.mockRejectedValueOnce(new Error("offline"));
    await showScreen("night");
    expect(updates.reloadAsync).not.toHaveBeenCalled();

    updateAvailable();
    await settle(UPDATE_CHECK_MS);
    expect(updates.reloadAsync).toHaveBeenCalledTimes(1);
  });

  it("waits for the next quiet night screen if a play starts while it downloads", async () => {
    updates.checkForUpdateAsync.mockResolvedValue({ isAvailable: true, isRollBackToEmbedded: false } as never);
    let finishDownload!: (result: unknown) => void;
    updates.fetchUpdateAsync.mockReturnValue(new Promise((resolve) => (finishDownload = resolve)) as never);
    await showScreen("night");

    await showScreen("playing");
    await act(async () => finishDownload({ isNew: true, isRollBackToEmbedded: false }));
    expect(updates.reloadAsync).not.toHaveBeenCalled();

    await showScreen("night");
    expect(updates.reloadAsync).toHaveBeenCalledTimes(1);
    // Downloaded once, not again.
    expect(updates.fetchUpdateAsync).toHaveBeenCalledTimes(1);
  });

  it("does nothing when updates are switched off", async () => {
    (updates as { isEnabled: boolean }).isEnabled = false;
    updateAvailable();
    await showScreen("night");

    expect(updates.checkForUpdateAsync).not.toHaveBeenCalled();
  });
});
