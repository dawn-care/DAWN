import React from "react";
import { act, fireEvent, render, screen } from "@testing-library/react-native";
import { setClock } from "../../clock/clock";
import { usePlayback } from "../../playback/PlaybackProvider";
import { loadPlayed } from "../../schedule/playedStore";
import { useScreen } from "../../screen/ScreenProvider";
import DevToolbar from "../DevToolbar";
import { fetchConnected, fetchMedia, startStaffPlay } from "../devInfo";

jest.mock("../../playback/PlaybackProvider", () => ({ usePlayback: jest.fn() }));
jest.mock("../../screen/ScreenProvider", () => ({ useScreen: jest.fn() }));
jest.mock("../../schedule/playedStore", () => ({ loadPlayed: jest.fn() }));
jest.mock("../devInfo", () => ({
  ...jest.requireActual("../devInfo"),
  fetchMedia: jest.fn(),
  fetchConnected: jest.fn(),
  startStaffPlay: jest.fn(),
}));

const stop = jest.fn();
const refresh = jest.fn();

function showing(playing: boolean) {
  jest.mocked(usePlayback).mockReturnValue({
    screen: playing ? "playing" : "alwaysOn",
    state: {
      pairing: "paired",
      carepartnerId: 4,
      playing: playing ? { playId: 1, item: { kind: "video", url: "/media/1/file", title: "Birthday" }, source: "staff" } : null,
    },
    stop,
  } as unknown as ReturnType<typeof usePlayback>);
}

async function openPanel() {
  await render(<DevToolbar />);
  await fireEvent.press(screen.getByTestId("dev-menu"));
  await act(async () => {});
}

describe("DevToolbar", () => {
  beforeEach(() => {
    setClock(new Date("2026-10-17T09:59:00"));
    showing(false);
    jest.mocked(useScreen).mockReturnValue({
      screen: {
        awakeHours: { wake: "07:00", bed: "21:00" }, chime: true,
        widgets: [],
        schedule: [
          { scheduledItemId: 9, date: "2026-10-17", time: "10:00", outsideAwakeHours: false, item: { kind: "video", url: "/m" } },
        ],
      },
      refresh,
      fetchedAt: new Date("2026-10-17T09:58:00"),
      failed: false,
    });
    jest.mocked(loadPlayed).mockResolvedValue(new Set());
    jest.mocked(fetchMedia).mockResolvedValue([{ id: 3, kind: "photo", name: "sunset" }]);
    jest.mocked(fetchConnected).mockResolvedValue(true);
    jest.mocked(startStaffPlay).mockResolvedValue(31);
  });

  afterEach(() => {
    setClock(null);
    jest.clearAllMocks();
  });

  it("shows only the menu button while nothing plays, with the panel closed", async () => {
    await render(<DevToolbar />);

    expect(screen.getByTestId("dev-menu")).toBeTruthy();
    expect(screen.queryByTestId("dev-skip")).toBeNull();
    expect(screen.queryByTestId("dev-panel")).toBeNull();
  });

  it("skips what is playing", async () => {
    showing(true);
    await render(<DevToolbar />);

    await fireEvent.press(screen.getByTestId("dev-skip"));

    expect(stop).toHaveBeenCalledWith();
  });

  it("shows the app's state in the panel", async () => {
    await openPanel();

    expect(screen.getByText(/7:00 to 21:00, day now/)).toBeTruthy();
    expect(screen.getByText(/4 \(paired\), socket connected/)).toBeTruthy();
    expect(screen.getByText(/today 10:00 · item 9 · video/)).toBeTruthy();
    expect(screen.getByText(/upcoming$/)).toBeTruthy();
    expect(screen.getByText("#3 · photo · sunset")).toBeTruthy();
  });

  it("staff-plays a media item when tapped", async () => {
    await openPanel();

    await fireEvent.press(screen.getByTestId("dev-media-3"));

    expect(startStaffPlay).toHaveBeenCalledWith(4, 3);
    expect(await screen.findByText("Sent staff play of media 3 (playback 31)")).toBeTruthy();
  });

  it("refetches the screen when asked", async () => {
    await openPanel();

    await fireEvent.press(screen.getByText("Refetch screen"));

    expect(refresh).toHaveBeenCalled();
  });
});
