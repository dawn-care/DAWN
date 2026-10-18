/**
 * MediaPlayer.tsx
 *
 * Plays a media item with the player for its kind.
 */

import React from "react";
import { resolveApiUri } from "../api/resolveApiUri";
import AudioPlayer from "./AudioPlayer";
import { DEFAULT_DURATION_SECONDS, type MediaItem } from "./media";
import PhotoPlayer from "./PhotoPlayer";
import VideoPlayer from "./VideoPlayer";

type MediaPlayerProps = {
  item: MediaItem;
  /** Called once when the item is actually playing. */
  onStarted?: () => void;
  /** Called once when the item has played to the end. */
  onFinished: () => void;
  /** Called once if the item can't be played. */
  onFailed: (reason: string) => void;
};

/**
 * Plays one media item, starting on its own. Remount it (change its `key`) to
 * play another.
 */
export default function MediaPlayer({ item, onStarted, onFinished, onFailed }: MediaPlayerProps) {
  const uri = resolveApiUri(item.url)!;

  switch (item.kind) {
    case "video":
      return <VideoPlayer uri={uri} onStarted={onStarted} onFinished={onFinished} onFailed={onFailed} />;
    case "audio":
      return (
        <AudioPlayer
          uri={uri}
          title={item.title}
          coverUri={resolveApiUri(item.coverUrl)}
          onStarted={onStarted}
          onFinished={onFinished}
          onFailed={onFailed}
        />
      );
    case "photo":
      return (
        <PhotoPlayer
          uri={uri}
          durationSeconds={item.durationSeconds && item.durationSeconds > 0 ? item.durationSeconds : DEFAULT_DURATION_SECONDS}
          onStarted={onStarted}
          onFinished={onFinished}
          onFailed={onFailed}
        />
      );
  }
}
