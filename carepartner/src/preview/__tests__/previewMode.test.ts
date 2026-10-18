import { PREVIEW, parsePreview } from "../previewMode";

describe("parsePreview", () => {
  it("reads the carepartner from a preview address", () => {
    expect(parsePreview("?preview=1&carepartner_id=12")).toEqual({ carepartnerId: 12 });
    expect(parsePreview("?carepartner_id=3&preview=1")).toEqual({ carepartnerId: 3 });
  });

  it.each(["", "?carepartner_id=12", "?preview=0&carepartner_id=12", "?preview=true&carepartner_id=12"])(
    "isn't a preview for %j",
    (search) => {
      expect(parsePreview(search)).toBeNull();
    },
  );

  it.each(["?preview=1", "?preview=1&carepartner_id=", "?preview=1&carepartner_id=0", "?preview=1&carepartner_id=-4", "?preview=1&carepartner_id=4.5", "?preview=1&carepartner_id=abc", "?preview=1&carepartner_id=99999999999999999999"])(
    "is a preview of no carepartner for %j",
    (search) => {
      expect(parsePreview(search)).toEqual({ carepartnerId: null });
    },
  );
});

describe("PREVIEW", () => {
  it("is never set on the tablet", () => {
    // Jest runs as a native platform, like the tablet
    expect(PREVIEW).toBeNull();
  });

  it("is read from the address on web", () => {
    const original = Object.getOwnPropertyDescriptor(globalThis, "location");
    Object.defineProperty(globalThis, "location", { value: { search: "?preview=1&carepartner_id=4" }, configurable: true });
    try {
      jest.isolateModules(() => {
        jest.doMock("react-native", () => ({ Platform: { OS: "web" } }));
        // eslint-disable-next-line @typescript-eslint/no-require-imports -- a fresh copy of the module, read on "web"
        const { PREVIEW: onWeb } = require("../previewMode");
        expect(onWeb).toEqual({ carepartnerId: 4 });
      });
    } finally {
      jest.dontMock("react-native");
      if (original) Object.defineProperty(globalThis, "location", original);
      else delete (globalThis as { location?: unknown }).location;
    }
  });
});
