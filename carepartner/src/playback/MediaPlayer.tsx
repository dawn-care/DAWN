/**
 * MediaPlayer.tsx
 *
 * Plays a media item from the server with the player for its type. This is
 * where server data enters playback, so the item's metadata is checked here,
 * and an item that can't be played renders nothing.
 */

import React from "react";
import { API_URL } from "../api/config";
import { resolveApiUri } from "../api/resolveApiUri";
import AudioPlayer from "./AudioPlayer";
import PhotoPlayer from "./PhotoPlayer";
import VideoPlayer from "./VideoPlayer";

/**
 * A module and its media, as the server sends it: the parts that say what to
 * play. Today these come from scheduled configurations; staff and tap plays
 * will use the same shape.
 */
export type MediaItem = {
  module_type: string;
  /**
   * Seconds to show a photo. It comes from the schedule entry, so a play that
   * isn't scheduled may have none and gets the default.
   */
  duration?: number | null;
  /**
   * The module's media details, unchecked server JSON (possibly null), so it
   * is checked once in `readMetadata` before use. It can be typed precisely
   * once responses are checked where they are fetched (#287).
   */
  metadata: unknown;
};

type MediaPlayerProps = {
  item: MediaItem;
  /** Called once when the item has played to the end. */
  onFinished: () => void;
  /** Called once if the item can't be played. */
  onFailed: (reason: string) => void;
};

/**
 * The server doesn't yet guarantee a duration above 0 (#344), so a missing or
 * non-positive one gets the caregiver form's default. Remove once #344 lands.
 */
export const DEFAULT_DURATION_SECONDS = 30;

const PLAYABLE_TYPES = ["video", "audio", "photo"];

/** Whether `MediaPlayer` can play this item. */
export function isPlayable(item: MediaItem): boolean {
  return PLAYABLE_TYPES.includes(item.module_type) && readMetadata(item.metadata) !== null;
}

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

/** The media metadata fields playback uses, or null if there's no usable id. */
function readMetadata(value: unknown): { id: number; displayedName?: string; coverUri?: string } | null {
  if (typeof value !== "object" || value === null) return null;
  const { id, displayed_name, cover_uri } = value as Record<string, unknown>;
  if (typeof id !== "number") return null;
  return {
    id,
    displayedName: typeof displayed_name === "string" ? displayed_name : undefined,
    coverUri: typeof cover_uri === "string" ? cover_uri : undefined,
  };
}
