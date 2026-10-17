import React from "react";
import { fireEvent, render, screen } from "@testing-library/react-native";
import { router } from "expo-router";
import PairScreen from "../../../app/pair";
import { usePlayback } from "../../playback/PlaybackProvider";
import { getCarepartnerId, markPairingComplete, storeCarepartnerId } from "../carepartnerId";

jest.mock("expo-router", () => ({
  Stack: { Screen: () => null },
  router: { push: jest.fn(), replace: jest.fn() },
}));

jest.mock("../../playback/PlaybackProvider", () => ({ usePlayback: jest.fn() }));

jest.mock("../carepartnerId", () => ({
  getCarepartnerId: jest.fn(),
  storeCarepartnerId: jest.fn(),
  markPairingComplete: jest.fn(),
}));

/** Answers the pair screen's requests by URL suffix. */
function mockBackend(routes: Record<string, { status?: number; body: unknown }>) {
  const fetchMock = jest.fn(async (url: string) => {
    const match = Object.keys(routes).find((suffix) => url.endsWith(suffix));
    if (!match) throw new Error(`Unexpected request: ${url}`);
    const { status = 200, body } = routes[match];
    return { ok: status < 400, status, json: async () => body };
  });
  globalThis.fetch = fetchMock as unknown as typeof fetch;
  return fetchMock;
}

describe("PairScreen", () => {
  const realFetch = globalThis.fetch;
  const checkPairing = jest.fn(async () => {});

  beforeEach(() => {
    jest.mocked(usePlayback).mockReturnValue({ checkPairing } as unknown as ReturnType<typeof usePlayback>);
  });

  afterEach(() => {
    globalThis.fetch = realFetch;
    jest.clearAllMocks();
  });

  it("shows a pairing code for the stored carepartner", async () => {
    jest.mocked(getCarepartnerId).mockResolvedValue(12);
    const fetchMock = mockBackend({
      "/users/carepartner/12": { body: { id: 12 } },
      "/users/device-pairing/code": { body: { pairing_code: "K7M4QP", expires_in_sec: 600 } },
    });

    await render(<PairScreen />);

    expect(await screen.findByText("K7M4QP")).toBeTruthy();
    expect(screen.getByText("Expires in 10:00")).toBeTruthy();
    expect(storeCarepartnerId).not.toHaveBeenCalled();
    expect(fetchMock).not.toHaveBeenCalledWith(
      expect.stringContaining("create-carepartner"),
      expect.anything(),
    );
  });

  it("creates and stores a carepartner when none is stored", async () => {
    jest.mocked(getCarepartnerId).mockResolvedValue(null);
    mockBackend({
      "/users/carepartner/create-carepartner": { body: { carepartner_id: 31 } },
      "/users/device-pairing/code": { body: { pairing_code: "ABC123", expires_in_sec: 600 } },
    });

    await render(<PairScreen />);

    expect(await screen.findByText("ABC123")).toBeTruthy();
    expect(storeCarepartnerId).toHaveBeenCalledWith(31);
  });

  it("shows the server's error when a code can't be generated", async () => {
    jest.mocked(getCarepartnerId).mockResolvedValue(12);
    mockBackend({
      "/users/carepartner/12": { body: { id: 12 } },
      "/users/device-pairing/code": { status: 500, body: { error: "Code service unavailable" } },
    });

    await render(<PairScreen />);

    expect(await screen.findByText("Code service unavailable")).toBeTruthy();
  });

  async function verifyWith(status: { status?: number; body: unknown }) {
    jest.mocked(getCarepartnerId).mockResolvedValue(12);
    mockBackend({
      "/users/carepartner/12": { body: { id: 12 } },
      "/users/device-pairing/code": { body: { pairing_code: "K7M4QP", expires_in_sec: 600 } },
      "/users/device-pairing/status": status,
    });
    await render(<PairScreen />);
    await screen.findByText("K7M4QP");
    // The instructions also say "Verify"; the button is the last one.
    await fireEvent.press(screen.getAllByText("Verify").at(-1)!);
  }

  it.each([
    ["the caregiver hasn't entered the code", { body: { paired: false } }, "Not paired yet. Enter the code in the caregiver app, then tap Verify."],
    ["the status can't be checked", { status: 500, body: {} }, "Something went wrong. Please check your connection and try again."],
  ])("stays on the pair screen when %s", async (_name, status, message) => {
    await verifyWith(status);

    expect(await screen.findByText(message)).toBeTruthy();
    expect(markPairingComplete).not.toHaveBeenCalled();
    expect(checkPairing).not.toHaveBeenCalled();
    expect(router.replace).not.toHaveBeenCalled();
  });

  it("records pairing, tells playback and goes home once the caregiver has paired", async () => {
    await verifyWith({ body: { paired: true } });

    expect(markPairingComplete).toHaveBeenCalled();
    expect(checkPairing).toHaveBeenCalled();
    expect(router.replace).toHaveBeenCalledWith("/");
  });

});
