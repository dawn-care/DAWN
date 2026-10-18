import { HttpError } from "../../api/request";
import { setClock } from "../../clock/clock";
import { createCarepartner, findCarepartner, requestPairingCode } from "../pairingApi";

jest.mock("../../api/request", () => ({
  ...jest.requireActual("../../api/request"),
  getJson: jest.fn(),
  postJson: jest.fn(),
}));

const { getJson, postJson } = jest.requireMock("../../api/request");

describe("pairingApi", () => {
  afterEach(() => {
    setClock(null);
    jest.clearAllMocks();
  });

  describe("findCarepartner", () => {
    it("is true when the server has the carepartner", async () => {
      getJson.mockResolvedValue({ id: 12 });

      await expect(findCarepartner(12)).resolves.toBe(true);
      expect(getJson).toHaveBeenCalledWith("/users/carepartner/12");
    });

    it("is false only when the server says it's gone", async () => {
      getJson.mockRejectedValue(new HttpError("/users/carepartner/12", 404));

      await expect(findCarepartner(12)).resolves.toBe(false);
    });

    it.each([
      ["a network failure", new Error("Network request failed")],
      ["a server error", new HttpError("/users/carepartner/12", 500)],
    ])("throws on %s, so being offline never counts as missing", async (_name, error) => {
      getJson.mockRejectedValue(error);

      await expect(findCarepartner(12)).rejects.toBe(error);
    });
  });

  describe("createCarepartner", () => {
    it("creates one with no name, for the caregiver to name, and returns its id", async () => {
      postJson.mockResolvedValue({ carepartner_id: 31 });

      await expect(createCarepartner()).resolves.toBe(31);
      expect(postJson).toHaveBeenCalledWith("/users/carepartner/create-carepartner", {});
    });

    it("throws on a malformed response", async () => {
      postJson.mockResolvedValue({ id: "31" });

      await expect(createCarepartner()).rejects.toThrow("malformed");
    });
  });

  describe("requestPairingCode", () => {
    it("returns the code and when it expires by the app clock", async () => {
      setClock(new Date("2026-10-17T10:00:00"));
      postJson.mockResolvedValue({ pairing_code: "K7M4QP", expires_in_sec: 600 });

      const { code, expiresAt } = await requestPairingCode(12);

      expect(code).toBe("K7M4QP");
      expect(expiresAt.getTime()).toBeCloseTo(new Date("2026-10-17T10:10:00").getTime(), -2);
      expect(postJson).toHaveBeenCalledWith("/users/device-pairing/code", { carepartner_id: 12 });
    });

    it("throws on a malformed response", async () => {
      postJson.mockResolvedValue({ pairing_code: "K7M4QP" });

      await expect(requestPairingCode(12)).rejects.toThrow("malformed");
    });
  });
});
