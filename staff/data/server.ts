/**
 * Calls to the real server. Everything without a backend yet (family folders,
 * sleeping hours, ratings, requests to the caregiver) stays in data/mock.ts.
 */

import { api, ApiError, apiWithStatus } from "../lib/api";
import type { Clip, Resident } from "./types";

interface ServerCarepartner {
  id: number;
  name: string;
  timezone: string;
}

interface ServerMedia {
  id: number;
  kind: "photo" | "video" | "audio";
  display_name: string;
  duration_seconds: number | null;
  created_at: string;
  url: string;
}

/** The parts of the server's Playback record the staff app uses. */
export interface ServerPlayback {
  id: number;
  carepartner_id: number;
  media_id: number | null;
  media_display_name: string;
  status: string;
  expires_at: string;
}

interface CommandResponse {
  playback: ServerPlayback;
  /** Whether the tablet had an open connection to receive the command. */
  delivered: boolean;
}

const toResident = ({ id, name }: ServerCarepartner): Resident => ({ id, name });

/** Every carepartner. For the demo, all residents are visible to all staff. */
export async function getResidents(): Promise<Resident[]> {
  // The server sends null rather than [] when there are none.
  const rows = await api<ServerCarepartner[] | null>("/users/carepartner");
  return (rows ?? []).map(toResident);
}

export async function getResident(id: number): Promise<Resident | undefined> {
  try {
    return toResident(await api<ServerCarepartner>(`/users/carepartner/${id}`));
  } catch (err) {
    if (err instanceof ApiError && err.status === 404) return undefined;
    throw err;
  }
}

/** A resident's videos and audio, newest first. Photos aren't played as clips. */
export async function getClips(residentId: number): Promise<Clip[]> {
  const media = await api<ServerMedia[] | null>(`/media?carepartner_id=${residentId}`);
  return (media ?? [])
    .filter((m): m is ServerMedia & { kind: "video" | "audio" } => m.kind !== "photo")
    .map((m) => ({
      id: m.id,
      kind: m.kind,
      title: m.display_name,
      duration: m.duration_seconds,
      addedAt: m.created_at,
    }));
}

export async function getClip(id: number): Promise<Clip | undefined> {
  try {
    const m = await api<ServerMedia>(`/media/${id}`);
    if (m.kind === "photo") return undefined;
    return { id: m.id, kind: m.kind, title: m.display_name, duration: m.duration_seconds, addedAt: m.created_at };
  } catch (err) {
    if (err instanceof ApiError && err.status === 404) return undefined;
    throw err;
  }
}

/**
 * Plays a clip on the resident's tablet. The server records the command and
 * pushes it over the tablet's WebSocket; `delivered` is false when the tablet
 * isn't connected (it then starts if the tablet reconnects within 30 seconds).
 */
export function playOnTablet(residentId: number, mediaId: number): Promise<CommandResponse> {
  return api<CommandResponse>("/playbacks/play", {
    method: "POST",
    body: JSON.stringify({ carepartner_id: residentId, media_id: mediaId }),
  });
}

/**
 * What happened when staff pressed Stop:
 * - "stopped": it had not reached the tablet yet, or was already stopped
 * - "stopping": the tablet has it and was told to stop
 * - "finished": it had already ended on the tablet, so there was nothing to stop
 * - "unreachable": the tablet may be playing but isn't connected, so it can't be stopped
 */
export type StopResult = "stopped" | "stopping" | "finished" | "unreachable";

export async function stopOnTablet(residentId: number, playbackId: number): Promise<StopResult> {
  try {
    const { status } = await apiWithStatus<CommandResponse>(`/playbacks/${playbackId}/stop`, {
      method: "POST",
      body: JSON.stringify({ carepartner_id: residentId }),
    });
    return status === 202 ? "stopping" : "stopped";
  } catch (err) {
    if (err instanceof ApiError && err.status === 409) {
      // 409 means either the tablet can't be reached, or the playback already
      // ended (finished, failed or expired), so there is nothing left to stop.
      return err.message === "tablet is not connected" ? "unreachable" : "finished";
    }
    throw err;
  }
}
