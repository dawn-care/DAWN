import { connectTablet, OUTBOX_LIMIT, RECONNECT_MAX_MS, tabletSocketUrl, type SocketLike } from "../socket";

/** A WebSocket the test opens, closes and sends messages through. */
class FakeSocket implements SocketLike {
  sent: string[] = [];
  closed = false;
  onopen: (() => void) | null = null;
  onmessage: ((event: { data: unknown }) => void) | null = null;
  onclose: (() => void) | null = null;
  onerror: (() => void) | null = null;

  constructor(public url: string) {}

  send(data: string) {
    this.sent.push(data);
  }
  close() {
    this.closed = true;
    this.onclose?.();
  }
  /** The server accepts the connection. */
  open() {
    this.onopen?.();
  }
  /** The connection fails or drops. */
  drop() {
    this.onerror?.();
    this.onclose?.();
  }
  receive(data: string) {
    this.onmessage?.({ data });
  }
}

const URL = "ws://server/api/realtime/tablet?carepartner_id=7";

function connect() {
  const sockets: FakeSocket[] = [];
  const onMessage = jest.fn();
  const connection = connectTablet({
    url: URL,
    onMessage,
    createSocket: (url) => {
      const socket = new FakeSocket(url);
      sockets.push(socket);
      return socket;
    },
  });
  const latest = () => sockets[sockets.length - 1];
  return { connection, sockets, latest, onMessage };
}

describe("tabletSocketUrl", () => {
  it("turns the API address into the socket address", () => {
    expect(tabletSocketUrl(7, "http://10.0.2.2:8080/api")).toBe("ws://10.0.2.2:8080/api/realtime/tablet?carepartner_id=7");
    expect(tabletSocketUrl(7, "https://example.org/api")).toBe("wss://example.org/api/realtime/tablet?carepartner_id=7");
  });
});

describe("connectTablet", () => {
  beforeEach(() => {
    jest.useFakeTimers();
  });

  afterEach(() => {
    jest.useRealTimers();
  });

  it("connects to the address straight away", () => {
    const { sockets } = connect();
    expect(sockets.map((socket) => socket.url)).toEqual([URL]);
  });

  it("passes on server messages and drops malformed ones with a warning", () => {
    const warn = jest.spyOn(console, "warn").mockImplementation(() => {});
    const { latest, onMessage } = connect();
    latest().open();

    latest().receive(JSON.stringify({ type: "refresh" }));
    latest().receive("not json");

    expect(onMessage).toHaveBeenCalledTimes(1);
    expect(onMessage).toHaveBeenCalledWith({ type: "refresh" });
    expect(warn).toHaveBeenCalledWith("Dropped a malformed message from the server");
    warn.mockRestore();
  });

  it("reconnects after a drop, waiting twice as long after each failed try, up to the limit", () => {
    const { sockets, latest } = connect();
    const waits: number[] = [];

    for (let i = 0; i < 7; i++) {
      latest().drop();
      const before = sockets.length;
      let waited = 0;
      while (sockets.length === before) {
        jest.advanceTimersByTime(1_000);
        waited += 1_000;
      }
      waits.push(waited);
    }

    expect(waits).toEqual([1_000, 2_000, 4_000, 8_000, 16_000, RECONNECT_MAX_MS, RECONNECT_MAX_MS]);
  });

  it("waits the shortest time again after a connection succeeds", () => {
    const { sockets, latest } = connect();
    latest().drop();
    jest.advanceTimersByTime(1_000);
    latest().drop();
    jest.advanceTimersByTime(2_000);
    expect(sockets).toHaveLength(3);

    latest().open();
    latest().drop();
    jest.advanceTimersByTime(1_000);

    expect(sockets).toHaveLength(4);
  });

  it("sends straight away when connected", () => {
    const { connection, latest } = connect();
    latest().open();

    connection.send("hello");

    expect(latest().sent).toEqual(["hello"]);
  });

  it("holds messages sent while disconnected and sends them in order once connected", () => {
    const { connection, latest } = connect();
    connection.send("first");
    latest().drop();
    connection.send("second");

    jest.advanceTimersByTime(1_000);
    latest().open();

    expect(latest().sent).toEqual(["first", "second"]);
  });

  it("keeps only the newest held messages", () => {
    const { connection, latest } = connect();
    for (let i = 0; i <= OUTBOX_LIMIT; i++) connection.send(String(i));

    latest().open();

    expect(latest().sent).toHaveLength(OUTBOX_LIMIT);
    expect(latest().sent[0]).toBe("1");
  });

  it("stops reconnecting once closed", () => {
    const { connection, sockets, latest } = connect();
    latest().open();

    connection.close();
    jest.advanceTimersByTime(RECONNECT_MAX_MS * 2);

    expect(latest().closed).toBe(true);
    expect(sockets).toHaveLength(1);
  });

  it("cancels a pending reconnect when closed", () => {
    const { connection, sockets, latest } = connect();
    latest().drop();

    connection.close();
    jest.advanceTimersByTime(RECONNECT_MAX_MS);

    expect(sockets).toHaveLength(1);
  });
});
