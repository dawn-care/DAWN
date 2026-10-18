/**
 * usePlayReports.ts
 *
 * Reports each scheduled play to the server as a Playback record, following
 * the playback machine the same way staff reports do (`useStaffCommands`):
 *
 *   on screen        → POST /playbacks (source "scheduled"), then acknowledged
 *   actually playing → started (a clip that never loads skips it)
 *   off screen       → finished, failed or stopped
 *
 * Reports go through a report queue, so ones made while offline are sent once
 * the connection returns, in order, each status after the record exists.
 * Only ids, the source and statuses are sent; never a title or other text.
 */

import { useEffect, useRef } from "react";
import { patchJson, postJson } from "../api/request";
import { serverDeadline } from "../clock/clock";
import { createReportQueue, type ReportQueue } from "../reports/queue";
import type { EndReason, Playing } from "./machine";
import { usePlayback } from "./PlaybackProvider";

/**
 * The deadline sent with a new record. The server refuses a record whose
 * deadline has passed, so a report queued during an outage longer than this
 * is lost (#379).
 */
export const REPORT_DEADLINE_MS = 60 * 60_000;

const END_STATUS: Record<EndReason, string> = {
  finished: "finished",
  failed: "failed",
  stopped: "stopped",
  replaced: "stopped",
};

/** A scheduled play with what its report needs; null for anything else. */
function reportable(play: Playing | null | undefined) {
  if (play?.source !== "scheduled" || play.item.id === undefined || play.scheduledItemId === undefined) return null;
  return { playId: play.playId, mediaId: play.item.id, scheduledItemId: play.scheduledItemId };
}

const isObject = (value: unknown): value is Record<string, unknown> => typeof value === "object" && value !== null;

type Reportable = NonNullable<ReturnType<typeof reportable>>;

/** Queues each play's reports, keeping the server's record id for each play once created. */
function createPlayReporter(queue: ReportQueue) {
  /** The server's record id for each play reported, by playId. */
  const recordIds = new Map<number, number>();

  return {
    created({ playId, mediaId, scheduledItemId }: Reportable, carepartnerId: number) {
      queue.add({
        group: `play-${playId}`,
        send: async () => {
          const record = await postJson("/playbacks", {
            carepartner_id: carepartnerId,
            media_id: mediaId,
            scheduled_item_id: scheduledItemId,
            source: "scheduled",
            expires_at: serverDeadline(REPORT_DEADLINE_MS).toISOString(),
          });
          // Retrying would create a second record, so a reply without an id
          // gives up on this play: its statuses are skipped.
          if (!isObject(record) || typeof record.id !== "number") {
            console.warn("Playback response is malformed; not reporting this play");
            return;
          }
          recordIds.set(playId, record.id);
        },
      });
    },
    /** Skipped if the play's record was never created (the server refused it). */
    status(playId: number, carepartnerId: number, status: string, last = false) {
      queue.add({
        group: `play-${playId}`,
        send: async () => {
          const id = recordIds.get(playId);
          if (id === undefined) return;
          await patchJson(`/playbacks/${id}/status`, { carepartner_id: carepartnerId, status });
          if (last) recordIds.delete(playId);
        },
      });
    },
  };
}

export function usePlayReports(): void {
  const { state } = usePlayback();
  const carepartnerId = state.carepartnerId;
  const reporter = useRef<ReturnType<typeof createPlayReporter> | null>(null);

  useEffect(() => {
    const queue = createReportQueue();
    reporter.current = createPlayReporter(queue);
    return () => {
      queue.stop();
      reporter.current = null;
    };
  }, []);

  const onScreen = reportable(state.playing);
  const playId = onScreen?.playId;
  const mediaId = onScreen?.mediaId;
  const scheduledItemId = onScreen?.scheduledItemId;
  useEffect(() => {
    if (playId === undefined || mediaId === undefined || scheduledItemId === undefined || carepartnerId == null) return;
    reporter.current?.created({ playId, mediaId, scheduledItemId }, carepartnerId);
    reporter.current?.status(playId, carepartnerId, "acknowledged");
  }, [playId, mediaId, scheduledItemId, carepartnerId]);

  const started = state.playing?.started === true;
  useEffect(() => {
    if (playId !== undefined && started && carepartnerId != null) reporter.current?.status(playId, carepartnerId, "started");
  }, [playId, started, carepartnerId]);

  const ended = state.lastEnded;
  useEffect(() => {
    const play = reportable(ended?.play);
    if (ended && play && carepartnerId != null) reporter.current?.status(play.playId, carepartnerId, END_STATUS[ended.reason], true);
  }, [ended, carepartnerId]);
}
