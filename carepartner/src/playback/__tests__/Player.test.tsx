import React from "react";
import { act, fireEvent, render, screen } from "@testing-library/react-native";
import { useAudioPlayer, useAudioPlayerStatus } from "expo-audio";
import { useVideoPlayer } from "expo-video";
import { AUDIO_LOAD_TIMEOUT_MS, AUDIO_STALL_TIMEOUT_MS } from "../AudioPlayer";
import Player from "../Player";

jest.mock("expo-video", () => {
  const { View } = jest.requireActual("react-native");
  return {
    useVideoPlayer: jest.fn(),
    VideoView: (props: { nativeControls?: boolean }) => (
      <View testID="video-view" accessibilityHint={`controls:${props.nativeControls}`} />
    ),
  };
});

jest.mock("expo-audio", () => ({
  useAudioPlayer: jest.fn(),
  useAudioPlayerStatus: jest.fn(),
}));

/** Stands in for the native video player: records listeners so tests can fire events. */
function fakeVideoPlayer() {
  const listeners = new Map<string, Set<(payload?: unknown) => void>>();
  return {
    loop: true,
    play: jest.fn(),
    addListener(event: string, listener: (payload?: unknown) => void) {
      if (!listeners.has(event)) listeners.set(event, new Set());
      listeners.get(event)!.add(listener);
      return { remove: () => listeners.get(event)!.delete(listener) };
    },
    emit(event: string, payload?: unknown) {
      listeners.get(event)?.forEach((listener) => listener(payload));
    },
  };
}

const audioStatus = (overrides: { isLoaded?: boolean; didJustFinish?: boolean; currentTime?: number } = {}) =>
  ({ isLoaded: false, didJustFinish: false, currentTime: 0, ...overrides }) as ReturnType<typeof useAudioPlayerStatus>;

describe("Player", () => {
  const onFinished = jest.fn();
  const onFailed = jest.fn();

  // Cleared before each test, not after: the previous test's player is unmounted
  // after its afterEach hooks, and that unmount calls into the mocks.
  beforeEach(() => {
    jest.clearAllMocks();
  });

  afterEach(() => {
    jest.useRealTimers();
  });

  describe("video", () => {
    let video: ReturnType<typeof fakeVideoPlayer>;

    beforeEach(() => {
      video = fakeVideoPlayer();
      jest.mocked(useVideoPlayer).mockImplementation((_source, setup) => {
        setup?.(video as never);
        return video as never;
      });
    });

    const renderVideo = () =>
      render(<Player kind="video" uri="http://server/video/file?id=5" onFinished={onFinished} onFailed={onFailed} />);

    it("starts the given video on its own, once, with no controls", async () => {
      await renderVideo();

      expect(useVideoPlayer).toHaveBeenCalledWith("http://server/video/file?id=5", expect.any(Function));
      expect(video.play).toHaveBeenCalledTimes(1);
      expect(video.loop).toBe(false);
      expect(screen.getByTestId("video-view").props.accessibilityHint).toBe("controls:false");
    });

    it("reports finished when the video plays to the end", async () => {
      await renderVideo();

      await act(async () => video.emit("playToEnd"));

      expect(onFinished).toHaveBeenCalledTimes(1);
      expect(onFailed).not.toHaveBeenCalled();
    });

    it("reports failed with the player's message on an error", async () => {
      await renderVideo();

      await act(async () => video.emit("statusChange", { status: "error", error: { message: "404" } }));

      expect(onFailed).toHaveBeenCalledWith("404");
      expect(onFinished).not.toHaveBeenCalled();
    });

    it("ignores status changes that aren't errors", async () => {
      await renderVideo();

      await act(async () => video.emit("statusChange", { status: "readyToPlay" }));

      expect(onFailed).not.toHaveBeenCalled();
    });
  });

  describe("audio", () => {
    const audio = { play: jest.fn(), pause: jest.fn() };

    beforeEach(() => {
      jest.mocked(useAudioPlayer).mockReturnValue(audio as never);
      jest.mocked(useAudioPlayerStatus).mockReturnValue(audioStatus());
    });

    // A fresh element each time, so rerender() actually renders again.
    const audioElement = () => (
      <Player
        kind="audio"
        uri="http://server/audio/file?id=9"
        title="Moon River"
        onFinished={onFinished}
        onFailed={onFailed}
      />
    );

    it("starts the given audio on its own and shows its title", async () => {
      await render(audioElement());

      expect(useAudioPlayer).toHaveBeenCalledWith("http://server/audio/file?id=9");
      expect(audio.play).toHaveBeenCalledTimes(1);
      expect(screen.getByText("Moon River")).toBeTruthy();
      expect(screen.queryByTestId("video-view")).toBeNull();
    });

    it("shows the music note when there is no image", async () => {
      await render(audioElement());

      expect(screen.getByTestId("audio-image").props.source).not.toEqual(
        expect.objectContaining({ uri: expect.anything() }),
      );
    });

    it("shows the given image, and falls back to the music note if it fails to load", async () => {
      const element = (
        <Player
          kind="audio"
          uri="http://server/audio/file?id=9"
          imageUri="http://server/audio/cover?id=9"
          onFinished={onFinished}
          onFailed={onFailed}
        />
      );
      await render(element);
      const image = () => screen.getByTestId("audio-image");
      expect(image().props.source).toEqual({ uri: "http://server/audio/cover?id=9" });

      await act(async () => fireEvent(image(), "error"));

      expect(image().props.source).not.toEqual(expect.objectContaining({ uri: expect.anything() }));
    });

    it("stops the audio when it is taken off screen", async () => {
      const view = await render(audioElement());
      expect(audio.pause).not.toHaveBeenCalled();

      await view.unmount();

      expect(audio.pause).toHaveBeenCalledTimes(1);
    });

    it("survives a player that is already released when taken off screen", async () => {
      audio.pause.mockImplementationOnce(() => {
        throw new Error("Unable to find the native shared object");
      });
      const view = await render(audioElement());

      await expect(view.unmount()).resolves.not.toThrow();
    });

    it("reports finished once when the audio ends", async () => {
      const view = await render(audioElement());

      jest.mocked(useAudioPlayerStatus).mockReturnValue(audioStatus({ isLoaded: true, didJustFinish: true }));
      await view.rerender(audioElement());
      await view.rerender(audioElement());

      expect(onFinished).toHaveBeenCalledTimes(1);
      expect(onFailed).not.toHaveBeenCalled();
    });

    it("reports failed if the audio never loads", async () => {
      jest.useFakeTimers();
      await render(audioElement());

      await act(async () => jest.advanceTimersByTime(AUDIO_LOAD_TIMEOUT_MS - 1));
      expect(onFailed).not.toHaveBeenCalled();

      await act(async () => jest.advanceTimersByTime(1));
      expect(onFailed).toHaveBeenCalledWith("Audio did not load");
    });

    it("does not report failed while loaded audio keeps moving", async () => {
      jest.useFakeTimers();
      const view = await render(audioElement());

      for (let second = 1; second <= 60; second++) {
        jest.mocked(useAudioPlayerStatus).mockReturnValue(audioStatus({ isLoaded: true, currentTime: second }));
        await view.rerender(audioElement());
        await act(async () => jest.advanceTimersByTime(1000));
      }

      expect(onFailed).not.toHaveBeenCalled();
    });

    it("reports failed if loaded audio stops moving without finishing", async () => {
      jest.useFakeTimers();
      const view = await render(audioElement());

      jest.mocked(useAudioPlayerStatus).mockReturnValue(audioStatus({ isLoaded: true, currentTime: 18 }));
      await view.rerender(audioElement());
      await act(async () => jest.advanceTimersByTime(AUDIO_STALL_TIMEOUT_MS - 1));
      expect(onFailed).not.toHaveBeenCalled();

      await act(async () => jest.advanceTimersByTime(1));
      expect(onFailed).toHaveBeenCalledWith("Audio stopped playing");
      expect(onFinished).not.toHaveBeenCalled();
    });

    it("does not report failed after the audio has finished", async () => {
      jest.useFakeTimers();
      const view = await render(audioElement());

      jest.mocked(useAudioPlayerStatus).mockReturnValue(audioStatus({ isLoaded: true, didJustFinish: true, currentTime: 30 }));
      await view.rerender(audioElement());
      await act(async () => jest.advanceTimersByTime(AUDIO_STALL_TIMEOUT_MS * 2));

      expect(onFinished).toHaveBeenCalledTimes(1);
      expect(onFailed).not.toHaveBeenCalled();
    });
  });
});
