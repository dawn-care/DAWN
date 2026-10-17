/**
 * socket.ts
 *
 * The tablet's WebSocket to the server, which pushes staff commands the moment
 * staff act. It reconnects by itself after a drop, waiting longer after each
 * failed try, and holds messages sent while disconnected until it reconnects.
 * The server's heartbeat pings are answered by the WebSocket itself.
 *
 * Logic only; `useStaffCommands` decides what to do with the messages.
 */

import { API_URL } from "../api/config";
import { readServerMessage, type ServerMessage } from "./socketMessages";

/** The wait before the first reconnect; it doubles after each failed try. */
export const RECONNECT_MIN_MS = 1_000;
/** The longest wait between reconnects. */
export const RECONNECT_MAX_MS = 30_000;
/** Messages held while disconnected, beyond which the oldest are dropped. */
export const OUTBOX_LIMIT = 100;

/** The parts of a WebSocket this uses, so tests can pass a fake. */
export type SocketLike = {
  send(data: string): void;
  close(): void;
  onopen: (() => void) | null;
  onmessage: ((event: { data: unknown }) => void) | null;
  onclose: (() => void) | null;
  onerror: (() => void) | null;
};

export type TabletConnection = {
  /** Sends now if connected, otherwise once the connection is back. */
  send(text: string): void;
  /** Closes the connection for good; it doesn't reconnect. */
  close(): void;
};

/** The tablet's socket address, from the API address (`http` becomes `ws`, `https` becomes `wss`). */
export function tabletSocketUrl(carepartnerId: number, apiUrl: string = API_URL): string {
  return `${apiUrl.replace(/^http/, "ws")}/realtime/tablet?carepartner_id=${carepartnerId}`;
}

export function connectTablet({
  url,
  onMessage,
  createSocket = (address) => new WebSocket(address) as unknown as SocketLike,
}: {
  url: string;
  onMessage: (message: ServerMessage) => void;
  createSocket?: (url: string) => SocketLike;
}): TabletConnection {
  let socket: SocketLike | null = null;
  let open = false;
  let closed = false;
  let delay = RECONNECT_MIN_MS;
  let retry: ReturnType<typeof setTimeout> | null = null;
  const outbox: string[] = [];

  const connect = () => {
    retry = null;
    const current = createSocket(url);
    socket = current;
    current.onopen = () => {
      open = true;
      delay = RECONNECT_MIN_MS;
      while (outbox.length) current.send(outbox.shift()!);
    };
    current.onmessage = (event) => {
      const message = readServerMessage(String(event.data));
      if (message) onMessage(message);
      else console.warn("Dropped a malformed message from the server");
    };
    // A failed connection reports an error and then closes, so only close retries.
    current.onerror = () => {};
    current.onclose = () => {
      if (socket !== current) return;
      open = false;
      socket = null;
      if (closed) return;
      retry = setTimeout(connect, delay);
      delay = Math.min(delay * 2, RECONNECT_MAX_MS);
    };
  };

  connect();

  return {
    send(text) {
      if (open && socket) {
        socket.send(text);
        return;
      }
      outbox.push(text);
      if (outbox.length > OUTBOX_LIMIT) outbox.shift();
    },
    close() {
      closed = true;
      if (retry) clearTimeout(retry);
      socket?.close();
      socket = null;
      open = false;
    },
  };
}
