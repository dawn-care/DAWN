/**
 * socketMessages.ts
 *
 * The messages on the tablet's WebSocket (see the server's
 * `internal/realtime/messages.go`). Every message is one JSON object with a
 * `type`; fields that don't apply to a type are left out. Messages from the
 * server are checked here, where they enter the app, and malformed ones are
 * dropped.
 */

/** The statuses the tablet reports for a play. Matches the server's `Playback.status`. */
export type PlaybackStatus = "acknowledged" | "started" | "finished" | "stopped" | "failed" | "expired";

export type ServerMessage =
  /** Play a media item now. The server decides whether it is still in time (see `useStaffCommands`). */
  | { type: "play"; playbackId: number; mediaId: number; expiresAt: Date }
  /** Stop the play with this record. */
  | { type: "stop"; playbackId: number }
  /** Fetch the schedule again. */
  | { type: "refresh" }
  /** The server's answer to a status report: the status it stored, or why it refused it. */
  | { type: "status_result"; playbackId: number; status?: string; error?: string }
  /** The server couldn't understand a message the tablet sent. */
  | { type: "error"; error?: string };

/** A server message, or null if it is malformed or of a type the tablet doesn't know. */
export function readServerMessage(raw: string): ServerMessage | null {
  let value: unknown;
  try {
    value = JSON.parse(raw);
  } catch {
    return null;
  }
  if (typeof value !== "object" || value === null) return null;
  const message = value as Record<string, unknown>;
  const optionalText = (field: unknown) => (typeof field === "string" ? field : undefined);

  switch (message.type) {
    case "play": {
      const expiresAt = typeof message.expires_at === "string" ? new Date(message.expires_at) : null;
      if (!isId(message.playback_id) || !isId(message.media_id) || !expiresAt || isNaN(expiresAt.getTime())) {
        return null;
      }
      return { type: "play", playbackId: message.playback_id, mediaId: message.media_id, expiresAt };
    }
    case "stop":
      return isId(message.playback_id) ? { type: "stop", playbackId: message.playback_id } : null;
    case "refresh":
      return { type: "refresh" };
    case "status_result":
      if (!isId(message.playback_id)) return null;
      return {
        type: "status_result",
        playbackId: message.playback_id,
        status: optionalText(message.status),
        error: optionalText(message.error),
      };
    case "error":
      return { type: "error", error: optionalText(message.error) };
    default:
      return null;
  }
}

/** The message reporting a play's status to the server. */
export function statusMessage(playbackId: number, status: PlaybackStatus): string {
  return JSON.stringify({ type: "status", playback_id: playbackId, status });
}

function isId(value: unknown): value is number {
  return typeof value === "number" && Number.isInteger(value) && value > 0;
}
