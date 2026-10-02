import React from "react";
import { render, screen } from "@testing-library/react-native";
import PairScreen from "../../../app/pair";
import { getCarepartnerId, storeCarepartnerId } from "../carepartnerId";

jest.mock("expo-router", () => ({
  Stack: { Screen: () => null },
  router: { push: jest.fn(), replace: jest.fn() },
}));

jest.mock("../carepartnerId", () => ({
  getCarepartnerId: jest.fn(),
  storeCarepartnerId: jest.fn(),
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
});
