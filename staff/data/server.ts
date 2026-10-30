/**
 * Calls to the real server. Everything without a backend yet (sleeping hours,
 * ratings, requests to the caregiver) stays in data/mock.ts.
 */

import { api, ApiError, apiWithStatus, mediaUrl } from "../lib/api";
import type { Clip, FamilyFolder, PlayedItem, Rating, Resident, SituationId } from "./types";

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
  situations: string[];
  family_member_ids: number[];
}

/** The parts of the server's Playback record the staff app uses. */
export interface ServerPlayback {
  id: number;
  carepartner_id: number;
  media_id: number | null;
  media_display_name: string;
  media_kind: "photo" | "video" | "audio";
  source: "staff" | "scheduled" | "tap";
  status: string;
  started_at: string | null;
  ended_at: string | null;
  expires_at: string;
  /** Until when staff can rate it; null if it never reached the screen. */
  rating_closes_at: string | null;
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

interface ServerFamilyMember {
  id: number;
  name: string;
  relationship: string;
  guidance_note: string | null;
}

const toFolder = (m: ServerFamilyMember): FamilyFolder => ({
  id: m.id,
  name: m.name,
  relationship: m.relationship,
  guidanceNote: m.guidance_note ?? undefined,
});

/** A resident's family, in the order the caregiver added them. */
export async function getFamilyFolders(residentId: number): Promise<FamilyFolder[]> {
  const members = await api<ServerFamilyMember[] | null>(`/family-members?carepartner_id=${residentId}`);
  return (members ?? []).map(toFolder);
}

/** "Asking for someone" isn't one of the server's situations: staff pick the family member instead. */
const SERVER_SITUATIONS: SituationId[] = [
  "wants_to_go_home",
  "doesnt_know_where",
  "anxious_or_upset",
  "starting_the_day",
  "winding_down",
];

/** Which clips to show: the ones about a family member, or for a situation. */
export type ClipFilter = { familyMemberId?: number; situation?: string };

/**
 * A resident's videos, audio and photos: what staff rated best first, newest
 * first among equal ratings (see rateClip).
 * With a family member, only clips about them; otherwise, with a situation the
 * caregiver can tag, only clips for it.
 */
export async function getClips(residentId: number, filter: ClipFilter = {}): Promise<Clip[]> {
  // What staff rated best comes first; the scores themselves aren't sent
  let query = `carepartner_id=${residentId}&sort=rank`;
  if (filter.familyMemberId) {
    query += `&family_member_id=${filter.familyMemberId}`;
  } else if (SERVER_SITUATIONS.includes(filter.situation as SituationId)) {
    query += `&situation=${filter.situation}`;
  }
  const media = await api<ServerMedia[] | null>(`/media?${query}`);
  return (media ?? [])
    // With no situation and no family member, the caregiver kept it out of the staff app
    .filter((m) => m.situations.length > 0 || m.family_member_ids.length > 0)
    .map(toClip);
}

/** How long the tablet shows a photo staff play (DEFAULT_DURATION_SECONDS in carepartner/src/playback/media.ts). */
const PHOTO_SECONDS = 30;

const toClip = (m: ServerMedia): Clip => ({
  id: m.id,
  kind: m.kind,
  title: m.display_name,
  duration: m.kind === "photo" ? PHOTO_SECONDS : m.duration_seconds,
  fileUrl: mediaUrl(m.url),
  addedAt: m.created_at,
});

export async function getClip(id: number): Promise<Clip | undefined> {
  try {
    return toClip(await api<ServerMedia>(`/media/${id}`));
  } catch (err) {
    if (err instanceof ApiError && err.status === 404) return undefined;
    throw err;
  }
}

/** What most recently played on the resident's tablet, or null if nothing has. */
export async function getLastPlayed(residentId: number): Promise<PlayedItem | null> {
  // The server lists only playbacks that started, most recent first.
  const [latest] = await api<ServerPlayback[]>(`/playbacks?carepartner_id=${residentId}&limit=1`);
  if (!latest?.started_at) return null;
  return {
    title: latest.media_display_name,
    kind: latest.media_kind,
    source: latest.source,
    status: latest.status as PlayedItem["status"],
    startedAt: latest.started_at,
  };
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

/** What the tablet has done with a play, as far as the Playing screen needs. */
export interface PlayState {
  /** When it appeared on the tablet's screen (ms since epoch); null if it hasn't. */
  startedAt: number | null;
  /** The tablet reported it finished, was stopped, or failed. */
  ended: boolean;
  /** The tablet couldn't play it to the end (the file didn't load, or playback stalled). */
  failed: boolean;
  /** Someone stopped it, from this phone or another. */
  stopped: boolean;
  /** Until when it can be rated (ms since epoch); null if it never reached the screen. */
  ratingClosesAt: number | null;
}

/** The server's record of a play, polled while the Playing screen is open. */
export async function getPlayState(residentId: number, playbackId: number): Promise<PlayState> {
  const p = await api<ServerPlayback>(`/playbacks/${playbackId}?carepartner_id=${residentId}`);
  return {
    startedAt: p.started_at ? Date.parse(p.started_at) : null,
    ended: p.ended_at !== null,
    failed: p.status === "failed",
    stopped: p.status === "stopped",
    ratingClosesAt: p.rating_closes_at ? Date.parse(p.rating_closes_at) : null,
  };
}

/**
 * What happened when staff rated a play:
 * - "saved": recorded; it ranks the resident's media for every staff member
 * - "closed": it can't be rated anymore. The screen only offers rating while it
 *   can be, so this only happens right at the end of the window.
 */
export type RateResult = "saved" | "closed";

/**
 * Saves how a play went. Rating it again inside the window replaces the
 * earlier rating.
 */
export async function rateClip(residentId: number, playbackId: number, rating: Rating): Promise<RateResult> {
  try {
    await api(`/playbacks/${playbackId}/rating`, {
      method: "POST",
      body: JSON.stringify({ carepartner_id: residentId, rating }),
    });
    return "saved";
  } catch (err) {
    if (err instanceof ApiError && err.status === 409) return "closed";
    throw err;
  }
}
