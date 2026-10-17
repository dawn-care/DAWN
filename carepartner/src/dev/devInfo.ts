/**
 * devInfo.ts
 *
 * What the developer toolbar shows that isn't already in app state: the
 * carepartner's media, whether the server sees the tablet's socket, and each
 * scheduled play's status. Logic only. Developer builds only.
 */

import { getJson, postJson } from "../api/request";
import type { ScheduledPlay } from "../screen/screen";
import { GRACE_MS, playKey, startOf } from "../schedule/schedule";

export type DevMedia = { id: number; kind: string; name: string };

export type PlayStatus =
  /** Started on this tablet. */
  | "played"
  /** In its grace window and not played yet. */
  | "due"
  | "upcoming"
  /** Its grace window passed without it playing. */
  | "missed"
  /** Outside the carepartner's awake hours, so it never plays. */
  | "asleep";

const isObject = (value: unknown): value is Record<string, unknown> => typeof value === "object" && value !== null;

/** The media list from `GET /media`, keeping only entries with an id, kind and name. */
export function readMediaList(raw: unknown): DevMedia[] {
  if (!Array.isArray(raw)) return [];
  return raw.flatMap((item) =>
    isObject(item) && typeof item.id === "number" && typeof item.kind === "string"
      ? [{ id: item.id, kind: item.kind, name: typeof item.display_name === "string" ? item.display_name : "" }]
      : [],
  );
}

/** Where a scheduled play stands at `time`, given the plays that have started. */
export function playStatus(play: ScheduledPlay, time: Date, played: Set<string>): PlayStatus {
  if (play.outsideAwakeHours) return "asleep";
  if (played.has(playKey(play))) return "played";
  const late = time.getTime() - startOf(play).getTime();
  if (late < 0) return "upcoming";
  return late < GRACE_MS ? "due" : "missed";
}

export async function fetchMedia(carepartnerId: number): Promise<DevMedia[]> {
  return readMediaList(await getJson(`/media?carepartner_id=${carepartnerId}`));
}

/** Whether the server has this tablet's socket open. */
export async function fetchConnected(carepartnerId: number): Promise<boolean> {
  const status = await getJson(`/realtime/carepartners/${carepartnerId}/connection`);
  return isObject(status) && status.connected === true;
}

/** Starts a staff play through the server, as the staff app will; returns the Playback id. */
export async function startStaffPlay(carepartnerId: number, mediaId: number): Promise<number | null> {
  const result = await postJson("/playbacks/play", { carepartner_id: carepartnerId, media_id: mediaId });
  const playback = isObject(result) && isObject(result.playback) ? result.playback : null;
  return playback && typeof playback.id === "number" ? playback.id : null;
}
