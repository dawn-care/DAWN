import React from "react";
import { act, render, screen } from "@testing-library/react-native";
import Index from "../../../app/index";
import MediaPlayer from "../MediaPlayer";
import { usePlayback } from "../PlaybackProvider";
import type { Screen } from "../machine";

jest.mock("expo-router", () => {
  const { Text } = jest.requireActual("react-native");
  return { Redirect: ({ href }: { href: string }) => <Text>redirect:{href}</Text> };
});
jest.mock("../PlaybackProvider", () => ({ usePlayback: jest.fn() }));
jest.mock("../MediaPlayer", () => ({ __esModule: true, default: jest.fn(() => null) }));
jest.mock("../../display/NightScreen", () => {
  const { Text } = jest.requireActual("react-native");
  return { __esModule: true, default: () => <Text>night</Text> };
});
jest.mock("../../screen/AlwaysOnScreen", () => {
  const { Text } = jest.requireActual("react-native");
  return { __esModule: true, default: () => <Text>always-on</Text> };
});

const ended = jest.fn();
const item = { kind: "video" as const, url: "/media/5/file" };

function showing(screenName: Screen) {
  jest.mocked(usePlayback).mockReturnValue({
    screen: screenName,
    state: { playing: screenName === "playing" ? { playId: 3, item, source: "scheduled" } : null },
    ended,
  } as unknown as ReturnType<typeof usePlayback>);
}

describe("index route", () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  it("shows nothing while pairing is being checked", async () => {
    showing("starting");
    await render(<Index />);
    expect(screen.toJSON()).toBeNull();
  });

  it("sends an unpaired tablet to pairing", async () => {
    showing("unpaired");
    await render(<Index />);
    expect(screen.getByText("redirect:/pair")).toBeTruthy();
  });

  it("shows the always-on screen when idle by day", async () => {
    showing("alwaysOn");
    await render(<Index />);
    expect(screen.getByText("always-on")).toBeTruthy();
  });

  it("shows the night screen when idle at night", async () => {
    showing("night");
    await render(<Index />);
    expect(screen.getByText("night")).toBeTruthy();
  });

  it("plays the current item full screen and reports how it ended", async () => {
    jest.spyOn(console, "warn").mockImplementation(() => {});
    showing("playing");
    await render(<Index />);

    const props = jest.mocked(MediaPlayer).mock.calls[0][0];
    expect(props.item).toBe(item);

    await act(async () => props.onFinished());
    await act(async () => props.onFailed("404"));
    expect(ended).toHaveBeenNthCalledWith(1, 3, "finished");
    expect(ended).toHaveBeenNthCalledWith(2, 3, "failed");
  });
});
