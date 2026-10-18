import React from "react";
import { act, render, screen } from "@testing-library/react-native";
import { useAudioPlayer, useAudioPlayerStatus } from "expo-audio";
import ChimeScreen, {
  CHIME_LOAD_TIMEOUT_MS,
  CHIME_MAX_MS,
  CHIME_PAUSE_MS,
  CHIME_VOLUME,
  NIGHT_CHIME_VOLUME,
} from "../ChimeScreen";

jest.mock("expo-audio", () => ({ useAudioPlayer: jest.fn(), useAudioPlayerStatus: jest.fn() }));

type Status = { isLoaded: boolean; didJustFinish: boolean };
const status = (overrides: Partial<Status> = {}) =>
  ({ isLoaded: false, didJustFinish: false, ...overrides }) as ReturnType<typeof useAudioPlayerStatus>;

describe("ChimeScreen", () => {
  const onDone = jest.fn();
  let chime: { play: jest.Mock; pause: jest.Mock; volume: number };

  const element = (props: Partial<React.ComponentProps<typeof ChimeScreen>> = {}) => (
    <ChimeScreen kind="video" sound night={false} onDone={onDone} {...props} />
  );
  const wait = (ms: number) => act(async () => jest.advanceTimersByTime(ms));

  beforeEach(() => {
    jest.useFakeTimers();
    jest.clearAllMocks();
    chime = { play: jest.fn(), pause: jest.fn(), volume: 1 };
    jest.mocked(useAudioPlayer).mockReturnValue(chime as never);
    jest.mocked(useAudioPlayerStatus).mockReturnValue(status());
  });

  afterEach(() => jest.useRealTimers());

  it("shows a calm screen with the media's kind, and plays the chime at the day volume", async () => {
    await render(element({ kind: "audio" }));

    expect(screen.getByTestId("chime-icon-audio")).toBeTruthy();
    expect(chime.play).toHaveBeenCalledTimes(1);
    expect(chime.volume).toBe(CHIME_VOLUME);
  });

  it("plays it quieter at night", async () => {
    await render(element({ night: true }));

    expect(chime.volume).toBe(NIGHT_CHIME_VOLUME);
  });

  it("is done a pause after the chime finishes", async () => {
    const view = await render(element());
    jest.mocked(useAudioPlayerStatus).mockReturnValue(status({ isLoaded: true }));
    await view.rerender(element());
    jest.mocked(useAudioPlayerStatus).mockReturnValue(status({ isLoaded: true, didJustFinish: true }));
    await view.rerender(element());

    await wait(CHIME_PAUSE_MS - 1);
    expect(onDone).not.toHaveBeenCalled();
    await wait(1);
    expect(onDone).toHaveBeenCalledTimes(1);
  });

  it("goes on without the chime if it doesn't load", async () => {
    await render(element());

    await wait(CHIME_LOAD_TIMEOUT_MS);
    await wait(CHIME_PAUSE_MS);

    expect(onDone).toHaveBeenCalledTimes(1);
  });

  it("goes on if the chime never finishes", async () => {
    const view = await render(element());
    jest.mocked(useAudioPlayerStatus).mockReturnValue(status({ isLoaded: true }));
    await view.rerender(element());

    await wait(CHIME_MAX_MS);
    await wait(CHIME_PAUSE_MS);

    expect(onDone).toHaveBeenCalledTimes(1);
  });

  it("keeps the pause but plays no sound when the chime is switched off", async () => {
    await render(element({ sound: false }));

    expect(chime.play).not.toHaveBeenCalled();
    await wait(CHIME_PAUSE_MS);
    expect(onDone).toHaveBeenCalledTimes(1);
  });

  it("stops the chime and isn't done once taken off screen", async () => {
    const view = await render(element());

    await view.unmount();
    await wait(CHIME_MAX_MS + CHIME_PAUSE_MS);

    expect(chime.pause).toHaveBeenCalled();
    expect(onDone).not.toHaveBeenCalled();
  });
});
