/**
 * StaffCommands.tsx
 *
 * Connects the tablet to the server's WebSocket once it is paired, carries out
 * staff play and stop commands through the playback machine, and reports each
 * step of a staff play back to the server, so its Playback record always says
 * what the tablet actually did:
 *
 *   play command        → acknowledged
 *   actually playing    → started (a clip that never loads skips it)
 *   off screen          → finished, failed or stopped
 *
 * The server decides whether a command is still in time: the tablet's clock
 * can be wrong, so a play starts only once the server has accepted its
 * acknowledgement. The media lookup runs meanwhile.
 *
 * A staff play pre-empts anything (see `canPlay`), and a newer command cancels
 * an older one still waiting to start. The server resends a live command every
 * time the tablet reconnects, so a command already handled isn't started twice.
 */

import { useEffect, useRef } from "react";
import { useScreen } from "../screen/ScreenProvider";
import type { EndReason } from "./machine";
import type { MediaItem } from "./media";
import { usePlayback } from "./PlaybackProvider";
import { connectTablet, tabletSocketUrl, type TabletConnection } from "./socket";
import { statusMessage, type PlaybackStatus, type ServerMessage } from "./socketMessages";
import { fetchMediaItem } from "./staffItem";

/** The status reported when a staff play leaves the screen. */
const END_STATUS: Record<EndReason, PlaybackStatus> = {
  finished: "finished",
  failed: "failed",
  stopped: "stopped",
  replaced: "stopped",
};

/** How many handled commands to remember; the server only ever resends the latest. */
const HANDLED_LIMIT = 50;

/** A command acknowledged but not yet on screen. */
type Pending = {
  /** The server accepted the acknowledgement, so the command is still in time. */
  accepted: boolean;
  /** The media, once looked up. */
  item?: MediaItem;
};

/** Rendered in `_layout.tsx` inside both providers; draws nothing. */
export default function StaffCommands() {
  const { state, play, stop } = usePlayback();
  const { refresh } = useScreen();
  const carepartnerId = state.carepartnerId;

  const connection = useRef<TabletConnection | null>(null);
  /** Plays already handled, so a command resent on reconnect isn't started twice. */
  const handled = useRef(new Set<number>());
  /** Commands acknowledged and not yet on screen. At most one, since a newer command cancels older ones. */
  const pending = useRef(new Map<number, Pending>());

  const report = (playbackId: number, status: PlaybackStatus) =>
    connection.current?.send(statusMessage(playbackId, status));

  /** Plays a pending command once it is both accepted and looked up. */
  const startIfReady = (playbackId: number) => {
    const command = pending.current.get(playbackId);
    if (!command?.accepted || !command.item) return;
    pending.current.delete(playbackId);
    latest.current.play(command.item, "staff", { playbackId });
  };

  const handle = async (message: ServerMessage) => {
    switch (message.type) {
      case "play": {
        const { playbackId, mediaId } = message;
        if (handled.current.has(playbackId)) {
          // A resend after reconnecting. If the server's answer to the
          // acknowledgement was lost, ask again; repeating it is harmless.
          if (pending.current.get(playbackId)?.accepted === false) report(playbackId, "acknowledged");
          return;
        }
        handled.current.add(playbackId);
        if (handled.current.size > HANDLED_LIMIT) handled.current.delete(handled.current.values().next().value!);

        // Staff replaced an earlier command before it started.
        for (const older of pending.current.keys()) {
          pending.current.delete(older);
          report(older, "stopped");
        }
        pending.current.set(playbackId, { accepted: false });
        report(playbackId, "acknowledged");

        const item = await fetchMediaItem(mediaId);
        const command = pending.current.get(playbackId);
        // Gone if it was stopped, replaced or refused while looking it up.
        if (!command) return;
        if (!item) {
          pending.current.delete(playbackId);
          report(playbackId, "failed");
          return;
        }
        command.item = item;
        startIfReady(playbackId);
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
      case "status_result": {
        const command = pending.current.get(message.playbackId);
        if (command && !command.accepted) {
          // The answer to the acknowledgement: play only if the server took it.
          // Anything else (expired, or refused) means it must not start.
          if (message.status === "acknowledged") {
            command.accepted = true;
            startIfReady(message.playbackId);
          } else {
            pending.current.delete(message.playbackId);
          }
        }
        if (message.error) console.warn(`Server refused a status for playback ${message.playbackId}: ${message.error}`);
        return;
      }
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
  const startedPlaybackId = state.playing?.started ? state.playing.playbackId : undefined;
  useEffect(() => {
    if (startedPlaybackId !== undefined) report(startedPlaybackId, "started");
  }, [startedPlaybackId]);

  useEffect(() => {
    const ended = state.lastEnded;
    if (ended?.play.playbackId !== undefined) report(ended.play.playbackId, END_STATUS[ended.reason]);
  }, [state.lastEnded]);

  return null;
}
