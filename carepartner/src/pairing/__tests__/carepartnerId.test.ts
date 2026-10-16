import { Platform } from "react-native";
import * as SecureStore from "expo-secure-store";
import { getCarepartnerId, storeCarepartnerId } from "../carepartnerId";

jest.mock("expo-secure-store", () => {
  const store = new Map<string, string>();
  return {
    getItemAsync: jest.fn(async (key: string) => store.get(key) ?? null),
    setItemAsync: jest.fn(async (key: string, value: string) => {
      store.set(key, value);
    }),
  };
});

describe("carepartnerId", () => {
  afterEach(() => {
    jest.restoreAllMocks();
  });

  describe("on native", () => {
    it("returns null when nothing is stored", async () => {
      expect(await getCarepartnerId()).toBeNull();
    });

    it("round-trips the id through secure storage", async () => {
      await storeCarepartnerId(42);

      expect(SecureStore.setItemAsync).toHaveBeenCalledWith("carepartnerID", "42");
      expect(await getCarepartnerId()).toBe(42);
    });
  });

  describe("on web", () => {
    const storage = new Map<string, string>();

    beforeEach(() => {
      storage.clear();
      jest.replaceProperty(Platform, "OS", "web");
      (globalThis as any).localStorage = {
        getItem: (key: string) => storage.get(key) ?? null,
        setItem: (key: string, value: string) => storage.set(key, value),
      };
    });

    afterEach(() => {
      delete (globalThis as any).localStorage;
    });

    it("returns null when nothing is stored", async () => {
      expect(await getCarepartnerId()).toBeNull();
    });

    it("round-trips the id through localStorage", async () => {
      await storeCarepartnerId(7);

      expect(storage.get("carepartnerID")).toBe("7");
      expect(await getCarepartnerId()).toBe(7);
    });
  });
});
