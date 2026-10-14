import React from "react";
import { render } from "@testing-library/react-native";
import { API_URL } from "../../api/config";
import AudioPlayer from "../AudioPlayer";
import { DEFAULT_DURATION_SECONDS, isPlayable, type MediaItem } from "../media";
import MediaPlayer from "../MediaPlayer";
import PhotoPlayer from "../PhotoPlayer";
import VideoPlayer from "../VideoPlayer";

// The players themselves are tested in players.test.tsx; here only which one
// is chosen, and with what, matters.
jest.mock("../VideoPlayer", () => ({ __esModule: true, default: jest.fn(() => null) }));
jest.mock("../AudioPlayer", () => ({ __esModule: true, default: jest.fn(() => null) }));
jest.mock("../PhotoPlayer", () => ({ __esModule: true, default: jest.fn(() => null) }));

const item = (module_type: string, metadata: unknown, duration: number | null = 20): MediaItem => ({
  module_type,
  duration,
  metadata,
});

const onFinished = jest.fn();
const onFailed = jest.fn();
const play = (media: MediaItem) =>
  render(<MediaPlayer item={media} onFinished={onFinished} onFailed={onFailed} />);
const callbacks = { onFinished, onFailed };

describe("MediaPlayer", () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  it("plays a video from its file", async () => {
    await play(item("video", { id: 5, displayed_name: "Birthday" }));

    expect(VideoPlayer).toHaveBeenCalledWith({ uri: `${API_URL}/video/file?id=5`, ...callbacks }, undefined);
  });

  it("plays audio with its name as the title and its cover", async () => {
    await play(item("audio", { id: 9, displayed_name: "Moon River", cover_uri: "/audio/cover?id=9" }));

    expect(AudioPlayer).toHaveBeenCalledWith(
      {
        uri: `${API_URL}/audio/file?id=9`,
        title: "Moon River",
        coverUri: `${API_URL}/audio/cover?id=9`,
        ...callbacks,
      },
      undefined,
    );
  });

  it("plays audio with no title or cover when the server sends none", async () => {
    await play(item("audio", { id: 9, cover_uri: null }));

    expect(AudioPlayer).toHaveBeenCalledWith(
      expect.objectContaining({ title: undefined, coverUri: undefined }),
      undefined,
    );
  });

  it("shows a photo for the item's duration", async () => {
    await play(item("photo", { id: 4 }, 45));

    expect(PhotoPlayer).toHaveBeenCalledWith(
      { uri: `${API_URL}/photo/file?id=4`, durationSeconds: 45, ...callbacks },
      undefined,
    );
  });

  it.each([null, 0, -3])("shows a photo for the default duration when the duration is %p", async (duration) => {
    await play(item("photo", { id: 4 }, duration));

    expect(PhotoPlayer).toHaveBeenCalledWith(
      expect.objectContaining({ durationSeconds: DEFAULT_DURATION_SECONDS }),
      undefined,
    );
  });

  it("shows a photo for the default duration when the item has no duration, as an unscheduled play may not", async () => {
    await play({ module_type: "photo", metadata: { id: 4 } });

    expect(PhotoPlayer).toHaveBeenCalledWith(
      expect.objectContaining({ durationSeconds: DEFAULT_DURATION_SECONDS }),
      undefined,
    );
  });

  it.each(["weather", "news", "messages", "event", "something-new"])("plays nothing for %s items", async (type) => {
    await play(item(type, { id: 1 }));

    expect(isPlayable(item(type, { id: 1 }))).toBe(false);
    expect(VideoPlayer).not.toHaveBeenCalled();
    expect(AudioPlayer).not.toHaveBeenCalled();
    expect(PhotoPlayer).not.toHaveBeenCalled();
  });

  it.each([null, undefined, "photo.jpg", {}, { id: "4" }])("plays nothing for unusable metadata %p", async (metadata) => {
    await play(item("photo", metadata));

    expect(isPlayable(item("photo", metadata))).toBe(false);
    expect(PhotoPlayer).not.toHaveBeenCalled();
  });

  it.each(["video", "audio", "photo"])("counts %s items with usable metadata as playable", (type) => {
    expect(isPlayable(item(type, { id: 1 }))).toBe(true);
  });
});
