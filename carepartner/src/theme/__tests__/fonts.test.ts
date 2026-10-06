import { fontFiles, fonts } from "../fonts";

describe("fonts", () => {
  it("loads a file for every font family the theme uses", () => {
    // A family name with no loaded file silently falls back to the system font on Android.
    for (const family of Object.values(fonts)) {
      expect(fontFiles[family]).toBeDefined();
    }
  });
});
