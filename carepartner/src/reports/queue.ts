/**
 * queue.ts
 *
 * Reports for the server (plays now, taps later), sent one at a time in the
 * order they were made and kept until they succeed. Logic only.
 *
 * - A network error or a server error (5xx) leaves the report at the front
 *   and tries again after `RETRY_FIRST_MS`, doubling to `RETRY_MAX_MS`, so
 *   reports made while offline go out once the connection returns.
 * - A refusal (4xx) won't succeed by retrying, so the report is dropped, and
 *   with it the rest of its group: the later steps of the same play would be
 *   refused too.
 * - At most `QUEUE_LIMIT` reports are kept; past that the oldest are dropped.
 *
 * Kept in memory, so an outage survives but an app restart doesn't. Warnings
 * give counts only, never a report's contents.
 */

import { HttpError } from "../api/request";

export const RETRY_FIRST_MS = 2_000;
export const RETRY_MAX_MS = 60_000;
export const QUEUE_LIMIT = 200;

export type Report = {
  /** Reports about one thing, such as one play; a refused report drops the rest of its group. */
  group: string;
  /** Sends the report; throws if it didn't get through. */
  send: () => Promise<void>;
};

export type ReportQueue = {
  add: (report: Report) => void;
  /** How many reports are waiting. */
  size: () => number;
  /** Stops sending and retrying; for unmounting. */
  stop: () => void;
};

export function createReportQueue(): ReportQueue {
  let reports: Report[] = [];
  let sending = false;
  let stopped = false;
  let retryMs = RETRY_FIRST_MS;
  let retry: ReturnType<typeof setTimeout> | undefined;

  async function flush() {
    if (sending || retry || stopped) return;
    sending = true;
    while (reports.length && !stopped) {
      const next = reports[0];
      try {
        await next.send();
        if (reports[0] === next) reports.shift();
        retryMs = RETRY_FIRST_MS;
      } catch (error) {
        if (error instanceof HttpError && error.status >= 400 && error.status < 500) {
          const before = reports.length;
          reports = reports.filter((report) => report.group !== next.group);
          console.warn(`The server refused a report (${error.status}); dropped ${before - reports.length}`);
          continue;
        }
        retry = setTimeout(() => {
          retry = undefined;
          flush();
        }, retryMs);
        retryMs = Math.min(retryMs * 2, RETRY_MAX_MS);
        break;
      }
    }
    sending = false;
  }

  return {
    add(report) {
      if (stopped) return;
      reports.push(report);
      if (reports.length > QUEUE_LIMIT) {
        reports.splice(0, reports.length - QUEUE_LIMIT);
        console.warn(`Too many unsent reports; dropped the oldest`);
      }
      flush();
    },
    size: () => reports.length,
    stop() {
      stopped = true;
      clearTimeout(retry);
    },
  };
}
