import { getCarepartnerId, isPairingComplete, markPairingComplete } from "../carepartnerId";
import { readPairedId } from "../pairingState";

jest.mock("../carepartnerId", () => ({
  getCarepartnerId: jest.fn(),
  isPairingComplete: jest.fn(),
  markPairingComplete: jest.fn(),
}));

function serverSays(answer: { status?: number; body?: unknown } | "offline") {
  globalThis.fetch = jest.fn(async () => {
    if (answer === "offline") throw new Error("offline");
    const { status = 200, body } = answer;
    return { ok: status < 400, status, json: async () => body };
  }) as unknown as typeof fetch;
}

describe("readPairedId", () => {
  const realFetch = globalThis.fetch;

  beforeEach(() => {
    jest.spyOn(console, "warn").mockImplementation(() => {});
    jest.mocked(getCarepartnerId).mockResolvedValue(7);
    jest.mocked(isPairingComplete).mockResolvedValue(false);
  });

  afterEach(() => {
    globalThis.fetch = realFetch;
    jest.restoreAllMocks();
    jest.clearAllMocks();
  });

  it("is unpaired with no stored carepartner", async () => {
    jest.mocked(getCarepartnerId).mockResolvedValue(null);
    serverSays({ body: { paired: true } });

    expect(await readPairedId()).toBeNull();
    expect(globalThis.fetch).not.toHaveBeenCalled();
  });

  it("is paired once pairing is recorded, without asking the server", async () => {
    jest.mocked(isPairingComplete).mockResolvedValue(true);
    serverSays({ body: { paired: false } });

    expect(await readPairedId()).toBe(7);
    expect(globalThis.fetch).not.toHaveBeenCalled();
  });

  it("is unpaired when the caregiver never finished pairing", async () => {
    serverSays({ body: { paired: false } });

    expect(await readPairedId()).toBeNull();
    expect(markPairingComplete).not.toHaveBeenCalled();
  });

  it("asks the server once for a tablet that paired before pairing was recorded", async () => {
    serverSays({ body: { paired: true } });

    expect(await readPairedId()).toBe(7);
    expect(markPairingComplete).toHaveBeenCalled();
    expect(jest.mocked(globalThis.fetch).mock.calls[0][1]).toMatchObject({ body: JSON.stringify({ carepartner_id: 7 }) });
  });

  it.each([
    ["is offline", "offline" as const],
    ["answers with an error", { status: 500, body: {} }],
    ["sends something else", { body: { status: "ok" } }],
  ])("keeps a stored tablet paired when the server %s", async (_name, answer) => {
    serverSays(answer);

    expect(await readPairedId()).toBe(7);
    expect(markPairingComplete).not.toHaveBeenCalled();
  });
});
