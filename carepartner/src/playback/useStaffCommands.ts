/**
 * useStaffCommands.ts
 *
 * Connects the tablet to the server's WebSocket once it is paired, carries out
 * staff play and stop commands through the playback machine, and reports each
 * step of a staff play back to the server, so its Playback record always says
 * what the tablet actually did:
 *
 *   play command → acknowledged (or expired, if it arrived too late)
 *   on screen    → started
 *   off screen   → finished, failed or stopped
 *
 * A staff play pre-empts anything (see `canPlay`). The server resends a live
 * command every time the tablet reconnects, so a command already handled is
 * ignored.
 */

import { useEffect, useRef } from "react";
import { now } from "../clock/clock";
import { useSchedule } from "../schedule/ScheduleProvider";
import type { EndReason } from "./machine";
import { isPlayable } from "./media";
import { usePlayback } from "./PlaybackProvider";
import { connectTablet, tabletSocketUrl, type TabletConnection } from "./socket";
import { statusMessage, type PlaybackStatus, type ServerMessage } from "./socketMessages";
import { fetchModuleItem } from "./staffItem";

/** The status reported when a staff play leaves the screen. */
const END_STATUS: Record<EndReason, PlaybackStatus> = {
  finished: "finished",
  failed: "failed",
  stopped: "stopped",
  replaced: "stopped",
};

export function useStaffCommands(): void {
  const { state, play, stop } = usePlayback();
  const { refresh } = useSchedule();
  const carepartnerId = state.carepartnerId;

  const connection = useRef<TabletConnection | null>(null);
  /** Plays already handled, so a command resent on reconnect is ignored. */
  const handled = useRef(new Set<number>());
  /** Plays acknowledged and being looked up, not yet on screen. */
  const pending = useRef(new Set<number>());

  const report = (playbackId: number, status: PlaybackStatus) =>
    connection.current?.send(statusMessage(playbackId, status));

  const handle = async (message: ServerMessage) => {
    switch (message.type) {
      case "play": {
        const { playbackId, moduleId, expiresAt } = message;
        if (handled.current.has(playbackId)) return;
        handled.current.add(playbackId);
        if (expiresAt <= now()) {
          report(playbackId, "expired");
          return;
        }
        report(playbackId, "acknowledged");
        pending.current.add(playbackId);
        const item = await fetchModuleItem(moduleId);
        // Gone if it was stopped, or the server said it expired, while looking it up.
        if (!pending.current.delete(playbackId)) return;
        if (!item || !isPlayable(item)) report(playbackId, "failed");
        else if (expiresAt <= now()) report(playbackId, "expired");
        else latest.current.play(item, "staff", playbackId);
        return;
      }
      case "stop":
        if (pending.current.delete(message.playbackId)) report(message.playbackId, "stopped");
        // The machine ignores it unless it is the play on screen.
        else latest.current.stop(message.playbackId);
        return;
      case "refresh":
        latest.current.refresh();
        return;
      case "status_result":
        // The tablet's clock can be wrong, so the server decides what expired.
        if (message.status === "expired") pending.current.delete(message.playbackId);
        if (message.error) console.warn(`Server refused a status for playback ${message.playbackId}: ${message.error}`);
        return;
      case "error":
        console.warn("Server couldn't read a message from the tablet:", message.error);
        return;
    }
  };

  // The socket calls the handler from this render, with this render's
  // play, stop and refresh, without reconnecting when they change.
  const latest = useRef({ handle, play, stop, refresh });
  useEffect(() => {
    latest.current = { handle, play, stop, refresh };
  });

  useEffect(() => {
    if (carepartnerId == null) return;
    const opened = connectTablet({
      url: tabletSocketUrl(carepartnerId),
      onMessage: (message) => latest.current.handle(message),
    });
    connection.current = opened;
    return () => {
      opened.close();
      if (connection.current === opened) connection.current = null;
    };
  }, [carepartnerId]);

  // Reports follow the machine, whatever put a play on or off the screen. Each
  // runs only when its play changes; `report` reads the connection from a ref.
  useEffect(() => {
    const playbackId = state.playing?.playbackId;
    if (playbackId !== undefined) report(playbackId, "started");
  }, [state.playing]);

  useEffect(() => {
    const ended = state.lastEnded;
    if (ended?.play.playbackId !== undefined) report(ended.play.playbackId, END_STATUS[ended.reason]);
  }, [state.lastEnded]);
}
