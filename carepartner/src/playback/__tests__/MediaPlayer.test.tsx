import React from "react";
import { render } from "@testing-library/react-native";
import { API_URL } from "../../api/config";
import AudioPlayer from "../AudioPlayer";
import { DEFAULT_DURATION_SECONDS, readMedia, type MediaItem } from "../media";
import MediaPlayer from "../MediaPlayer";
import PhotoPlayer from "../PhotoPlayer";
import VideoPlayer from "../VideoPlayer";

// The players themselves are tested in players.test.tsx; here only which one
// is chosen, and with what, matters.
jest.mock("../VideoPlayer", () => ({ __esModule: true, default: jest.fn(() => null) }));
jest.mock("../AudioPlayer", () => ({ __esModule: true, default: jest.fn(() => null) }));
jest.mock("../PhotoPlayer", () => ({ __esModule: true, default: jest.fn(() => null) }));

const onFinished = jest.fn();
const onFailed = jest.fn();
const play = (item: MediaItem) => render(<MediaPlayer item={item} onFinished={onFinished} onFailed={onFailed} />);
const callbacks = { onFinished, onFailed };

describe("MediaPlayer", () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  it("plays a video from its file", async () => {
    await play({ kind: "video", url: "/media/5/file" });

    expect(VideoPlayer).toHaveBeenCalledWith({ uri: `${API_URL}/media/5/file`, ...callbacks }, undefined);
  });

  it("plays audio with its title and cover", async () => {
    await play({ kind: "audio", url: "/media/9/file", title: "Moon River", coverUrl: "/media/4/file" });

    expect(AudioPlayer).toHaveBeenCalledWith(
      { uri: `${API_URL}/media/9/file`, title: "Moon River", coverUri: `${API_URL}/media/4/file`, ...callbacks },
      undefined,
    );
  });

  it("plays audio with no title or cover when it has none", async () => {
    await play({ kind: "audio", url: "/media/9/file" });

    expect(AudioPlayer).toHaveBeenCalledWith(
      expect.objectContaining({ title: undefined, coverUri: undefined }),
      undefined,
    );
  });

  it("shows a photo for its duration", async () => {
    await play({ kind: "photo", url: "/media/4/file", durationSeconds: 45 });

    expect(PhotoPlayer).toHaveBeenCalledWith(
      { uri: `${API_URL}/media/4/file`, durationSeconds: 45, ...callbacks },
      undefined,
    );
  });

  it.each([undefined, null, 0, -3])("shows a photo for the default duration when the duration is %p", async (durationSeconds) => {
    await play({ kind: "photo", url: "/media/4/file", durationSeconds });

    expect(PhotoPlayer).toHaveBeenCalledWith(
      expect.objectContaining({ durationSeconds: DEFAULT_DURATION_SECONDS }),
      undefined,
    );
  });
});

describe("readMedia", () => {
  it("reads the server's media, keeping its id for play reports", () => {
    expect(readMedia({ id: 9, kind: "audio", display_name: "Moon River", url: "/media/9/file", duration_seconds: 214 })).toEqual({
      id: 9,
      kind: "audio",
      url: "/media/9/file",
      title: "Moon River",
    });
  });

  it.each([
    ["null", null],
    ["a string", "video.mp4"],
    ["no kind", { url: "/media/1/file" }],
    ["an unknown kind", { kind: "weather", url: "/media/1/file" }],
    ["no file", { kind: "video" }],
    ["an empty file", { kind: "video", url: "" }],
  ])("rejects %s", (_name, value) => {
    expect(readMedia(value)).toBeNull();
  });
});
