/**
 * labels.ts
 *
 * User: Caregiver
 *
 * Short wording for a media item: the badge on its card ("Video 0:48")
 * and its status line ("On her screen", "Not scheduled"). Both are
 * worked out from the item's kind, length and `usage`.
 */

import { KIND_LABELS } from "./constants";
import type { Media } from "./api";

/** Seconds as m:ss, e.g. 48 -> "0:48", 192 -> "3:12" */
export function formatDuration(seconds: number): string {
  const total = Math.round(seconds);
  return `${Math.floor(total / 60)}:${String(total % 60).padStart(2, "0")}`;
}

/** "Video 0:48", "Audio 3:12" or "Photo" */
export function badgeLabel(media: Media): string {
  const kind = KIND_LABELS[media.kind];
  return media.duration_seconds === null ? kind : `${kind} ${formatDuration(media.duration_seconds)}`;
}

export interface MediaStatus {
  text: string;
  /** True when the item is in use, so the card can colour it */
  active: boolean;
}

/**
 * Where the item is used. Photos can be on her screen, scheduled and an audio
 * background all at once; the line names the first that applies. Video and
 * audio can only be scheduled.
 */
export function mediaStatus(media: Media): MediaStatus {
  const { usage } = media;

  if (media.kind === "photo") {
    if (usage.includes("on_screen")) return { text: "On her screen", active: true };
    if (usage.includes("scheduled")) return { text: "Scheduled", active: true };
    if (usage.includes("audio_background")) return { text: "Audio background", active: true };
    return { text: "Saved", active: false };
  }

  return usage.includes("scheduled")
    ? { text: "Scheduled", active: true }
    : { text: "Not scheduled", active: false };
}
