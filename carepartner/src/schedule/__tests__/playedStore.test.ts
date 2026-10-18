import { setStoredItem } from "../../storage/storage";
import { loadPlayed, savePlayed } from "../playedStore";

jest.mock("../../storage/storage", () => {
  const store = new Map<string, string>();
  return {
    getStoredItem: jest.fn(async (key: string) => store.get(key) ?? null),
    setStoredItem: jest.fn(async (key: string, value: string) => {
      store.set(key, value);
    }),
  };
});

describe("playedStore", () => {
  beforeEach(() => {
    jest.spyOn(console, "warn").mockImplementation(() => {});
  });

  afterEach(() => {
    jest.mocked(console.warn).mockRestore();
  });

  it("has nothing played before anything is stored", async () => {
    expect(await loadPlayed()).toEqual(new Set());
  });

  it("keeps today's and tomorrow's plays and drops earlier days", async () => {
    await savePlayed(new Set(["1@2026-10-11", "1@2026-10-12", "2@2026-10-13"]), "2026-10-12");

    expect(await loadPlayed()).toEqual(new Set(["1@2026-10-12", "2@2026-10-13"]));
  });

  it("has nothing played if what is stored can't be read", async () => {
    await setStoredItem("playedScheduledItems", "not json");

    expect(await loadPlayed()).toEqual(new Set());
    expect(console.warn).toHaveBeenCalled();
  });

  it("carries on if storing fails", async () => {
    jest.mocked(setStoredItem).mockRejectedValueOnce(new Error("full"));

    await expect(savePlayed(new Set(["1@2026-10-12"]), "2026-10-12")).resolves.toBeUndefined();
    expect(console.warn).toHaveBeenCalled();
  });
});
