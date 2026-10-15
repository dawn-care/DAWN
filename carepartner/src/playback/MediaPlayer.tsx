/**
 * MediaPlayer.tsx
 *
 * Plays a media item from the server with the player for its type, using
 * the metadata checked in media.ts. An item that can't be played renders
 * nothing.
 */

import React from "react";
import { API_URL } from "../api/config";
import { resolveApiUri } from "../api/resolveApiUri";
import AudioPlayer from "./AudioPlayer";
import { DEFAULT_DURATION_SECONDS, readMetadata, type MediaItem } from "./media";
import PhotoPlayer from "./PhotoPlayer";
import VideoPlayer from "./VideoPlayer";

type MediaPlayerProps = {
  item: MediaItem;
  /** Called once when the item has played to the end. */
  onFinished: () => void;
  /** Called once if the item can't be played. */
  onFailed: (reason: string) => void;
};

/**
 * Plays one media item, starting on its own, and renders nothing for an
 * item `isPlayable` rejects. Remount it (change its `key`) to play another.
 */
export default function MediaPlayer({ item, onFinished, onFailed }: MediaPlayerProps) {
  const metadata = readMetadata(item.metadata);
  if (!metadata) return null;
  const uri = `${API_URL}/${item.module_type}/file?id=${metadata.id}`;

  switch (item.module_type) {
    case "video":
      return <VideoPlayer uri={uri} onFinished={onFinished} onFailed={onFailed} />;
    case "audio":
      return (
        <AudioPlayer
          uri={uri}
          title={metadata.displayedName}
          coverUri={resolveApiUri(metadata.coverUri)}
          onFinished={onFinished}
          onFailed={onFailed}
        />
      );
    case "photo":
      return (
        <PhotoPlayer
          uri={uri}
          durationSeconds={item.duration && item.duration > 0 ? item.duration : DEFAULT_DURATION_SECONDS}
          onFinished={onFinished}
          onFailed={onFailed}
        />
      );
    default:
      return null;
  }
}
