import { API_URL } from "../../api/config";
import { fetchModuleItem } from "../staffItem";

describe("fetchModuleItem", () => {
  const realFetch = globalThis.fetch;
  let warn: jest.SpyInstance;

  const respond = (body: unknown, ok = true) => {
    globalThis.fetch = jest.fn(async () => ({ ok, status: ok ? 200 : 500, json: async () => body })) as unknown as typeof fetch;
  };

  beforeEach(() => {
    warn = jest.spyOn(console, "warn").mockImplementation(() => {});
  });

  afterEach(() => {
    globalThis.fetch = realFetch;
    warn.mockRestore();
  });

  it("looks up the module and returns the media it plays", async () => {
    respond({ id: 7, module_type: "video", caregiver_id: 2, carepartner_id: 2, metadata_id: 12 });

    await expect(fetchModuleItem(7)).resolves.toEqual({ module_type: "video", metadata: { id: 12 } });
    expect(globalThis.fetch).toHaveBeenCalledWith(`${API_URL}/modules/7`);
  });

  it.each([
    ["an error response", () => respond("server error", false)],
    ["no media id", () => respond({ id: 7, module_type: "video" })],
    ["no type", () => respond({ id: 7, metadata_id: 12 })],
    ["null", () => respond(null)],
    [
      "a network failure",
      () => {
        globalThis.fetch = jest.fn(async () => {
          throw new Error("offline");
        }) as unknown as typeof fetch;
      },
    ],
  ])("returns null with a warning for %s", async (_name, setUp) => {
    setUp();

    await expect(fetchModuleItem(7)).resolves.toBeNull();
    expect(warn).toHaveBeenCalled();
  });
});
