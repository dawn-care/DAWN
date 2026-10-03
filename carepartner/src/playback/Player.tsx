import React from "react";
import AudioPlayer from "./AudioPlayer";
import VideoPlayer from "./VideoPlayer";
import type { MediaKind, MediaPlayerProps } from "./types";

type PlayerProps = MediaPlayerProps & {
  kind: MediaKind;
  title?: string;
  /** Audio only: the image shown while it plays. */
  imageUri?: string;
};

/** Plays one piece of media with the player for its kind. */
export default function Player({ kind, title, imageUri, ...media }: PlayerProps) {
  if (kind === "audio") return <AudioPlayer title={title} imageUri={imageUri} {...media} />;
  return <VideoPlayer {...media} />;
}
