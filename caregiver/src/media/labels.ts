/**
 * labels.ts
 *
 * User: Caregiver
 *
 * Short wording for a media item: the badge on its card ("Video 0:48",
 * from its kind and length), its status line ("On screen", "Not
 * scheduled", from its kind and `usage`) and how staff find it.
 */

import { KIND_LABELS, SITUATIONS } from "./constants";
import type { FamilyMember, Media, ScheduledItem } from "./api";
import { describeSchedule } from "./schedule";

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
 * Where the item is used. Photos can be on screen, scheduled and an audio
 * background all at once; the line names the first that applies. Video and
 * audio can only be scheduled.
 */
export function mediaStatus(media: Media): MediaStatus {
  const { usage } = media;

  if (media.kind === "photo") {
    if (usage.includes("on_screen")) return { text: "On screen", active: true };
    if (usage.includes("scheduled")) return { text: "Scheduled", active: true };
    if (usage.includes("audio_background")) return { text: "Audio background", active: true };
    return { text: "Saved", active: false };
  }

  return usage.includes("scheduled")
    ? { text: "Scheduled", active: true }
    : { text: "Not scheduled", active: false };
}

/**
 * When the item plays, e.g. "Every day at 8:00 am", or null if it isn't
 * scheduled. `scheduled` is null when the schedule couldn't be loaded; the
 * item's `usage` still says it is scheduled, so fall back to a plain
 * "Scheduled".
 */
export function scheduleLine(media: Media, scheduled: ScheduledItem[] | null): string | null {
  const line = scheduled ? describeSchedule(scheduled, media.id) : null;
  return line ?? (media.usage.includes("scheduled") ? "Scheduled" : null);
}

/** "Every situation", "None", or the chosen ones, e.g. "Wants to go home, Anxious or upset" */
export function situationsLine(media: Media): string {
  if (media.situations.length === 0) return "None";
  if (media.situations.length === SITUATIONS.length) return "Every situation";
  return SITUATIONS.filter(({ value }) => media.situations.includes(value))
    .map(({ label }) => label)
    .join(", ");
}

/** "Elena, Daniel", "No one in particular", or a count if the family couldn't be loaded */
export function familyLine(media: Media, family: FamilyMember[] | null): string {
  const ids = media.family_member_ids;
  if (ids.length === 0) return "No one in particular";
  if (family === null) return ids.length === 1 ? "1 family member" : `${ids.length} family members`;
  return family
    .filter((member) => ids.includes(member.id))
    .map((member) => member.name)
    .join(", ");
}
