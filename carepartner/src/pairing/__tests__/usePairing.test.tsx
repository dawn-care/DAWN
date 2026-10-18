import { act, renderHook } from "@testing-library/react-native";
import { now, setClock } from "../../clock/clock";
import { usePlayback } from "../../playback/PlaybackProvider";
import { getCarepartnerId, markPairingComplete, storeCarepartnerId } from "../carepartnerId";
import { createCarepartner, findCarepartner, requestPairingCode } from "../pairingApi";
import { fetchPairingStatus } from "../pairingState";
import { PAIRING_POLL_MS, usePairing } from "../usePairing";

jest.mock("../../playback/PlaybackProvider", () => ({ usePlayback: jest.fn() }));
jest.mock("../carepartnerId", () => ({
  getCarepartnerId: jest.fn(),
  storeCarepartnerId: jest.fn(async () => {}),
  markPairingComplete: jest.fn(async () => {}),
}));
jest.mock("../pairingApi", () => ({
  findCarepartner: jest.fn(),
  createCarepartner: jest.fn(),
  requestPairingCode: jest.fn(),
}));
jest.mock("../pairingState", () => ({ fetchPairingStatus: jest.fn() }));

const checkPairing = jest.fn(async () => {});

/** A code from the server, lasting `seconds` from now by the app clock. */
const codeLasting = (code: string, seconds = 600) => ({ code, expiresAt: new Date(now().getTime() + seconds * 1000) });

const wait = (ms: number) => act(() => jest.advanceTimersByTimeAsync(ms));

async function start() {
  const view = await renderHook(() => usePairing());
  await wait(0);
  return view;
}

describe("usePairing", () => {
  beforeEach(() => {
    jest.useFakeTimers();
    setClock(new Date("2026-10-17T10:00:00"));
    jest.spyOn(console, "warn").mockImplementation(() => {});
    jest.mocked(usePlayback).mockReturnValue({ checkPairing } as unknown as ReturnType<typeof usePlayback>);
    jest.mocked(getCarepartnerId).mockResolvedValue(12);
    jest.mocked(findCarepartner).mockResolvedValue(true);
    jest.mocked(fetchPairingStatus).mockResolvedValue(false);
    jest.mocked(requestPairingCode).mockImplementation(async () => codeLasting("K7M4QP"));
  });

  afterEach(() => {
    setClock(null);
    jest.useRealTimers();
    jest.restoreAllMocks();
    jest.clearAllMocks();
  });

  it("shows a code for the stored carepartner without creating another", async () => {
    const { result } = await start();

    expect(result.current.view).toMatchObject({ kind: "showing", code: "K7M4QP" });
    expect(requestPairingCode).toHaveBeenCalledWith(12);
    expect(createCarepartner).not.toHaveBeenCalled();
  });

  it("creates and stores a carepartner when none is stored, or the server says it's gone", async () => {
    jest.mocked(findCarepartner).mockResolvedValue(false);
    jest.mocked(createCarepartner).mockResolvedValue(31);

    const { result } = await start();

    expect(storeCarepartnerId).toHaveBeenCalledWith(31);
    expect(requestPairingCode).toHaveBeenCalledWith(31);
    expect(result.current.view.kind).toBe("showing");
  });

  it("never creates a carepartner while the server can't be reached, and tries again", async () => {
    jest.mocked(findCarepartner).mockRejectedValueOnce(new Error("Network request failed"));

    const { result } = await start();
    expect(result.current.view.kind).toBe("offline");
    expect(createCarepartner).not.toHaveBeenCalled();

    await wait(PAIRING_POLL_MS);

    expect(result.current.view).toMatchObject({ kind: "showing", code: "K7M4QP" });
    expect(createCarepartner).not.toHaveBeenCalled();
  });

  it("finishes by itself once a caregiver enters the code", async () => {
    const { result } = await start();
    jest.mocked(fetchPairingStatus).mockResolvedValue(true);

    await wait(PAIRING_POLL_MS);

    expect(markPairingComplete).toHaveBeenCalled();
    expect(checkPairing).toHaveBeenCalled();
    expect(result.current.view.kind).toBe("paired");
  });

  it("gets a new code when the old one expires, checking first that it wasn't just used", async () => {
    jest.mocked(requestPairingCode).mockImplementationOnce(async () => codeLasting("OLD111", 3));
    const { result } = await start();
    expect(result.current.view).toMatchObject({ code: "OLD111" });
    jest.mocked(fetchPairingStatus).mockClear();

    await wait(3_000);

    expect(fetchPairingStatus).toHaveBeenCalledTimes(1);
    expect(result.current.view).toMatchObject({ kind: "showing", code: "K7M4QP" });
  });

  it("finishes instead of getting a new code when the code was used just before it expired", async () => {
    jest.mocked(requestPairingCode).mockImplementationOnce(async () => codeLasting("OLD111", 3));
    const { result } = await start();
    jest.mocked(fetchPairingStatus).mockResolvedValue(true);

    await wait(3_000);

    expect(requestPairingCode).toHaveBeenCalledTimes(1);
    expect(result.current.view.kind).toBe("paired");
  });

  it("gets a new code straight away when asked", async () => {
    jest.mocked(requestPairingCode).mockImplementationOnce(async () => codeLasting("OLD111"));
    const { result } = await start();

    await act(async () => result.current.newCode());
    await wait(0);

    expect(result.current.view).toMatchObject({ code: "K7M4QP" });
  });

  it("keeps showing a code that still works while the server is briefly unreachable", async () => {
    const { result } = await start();
    jest.mocked(fetchPairingStatus).mockRejectedValueOnce(new Error("Network request failed"));

    await wait(PAIRING_POLL_MS);

    expect(result.current.view).toMatchObject({ kind: "showing", code: "K7M4QP" });
  });

  it("stops checking once unmounted", async () => {
    const { unmount } = await start();
    jest.mocked(fetchPairingStatus).mockClear();

    await act(async () => unmount());
    await wait(3 * PAIRING_POLL_MS);

    expect(fetchPairingStatus).not.toHaveBeenCalled();
  });
});
