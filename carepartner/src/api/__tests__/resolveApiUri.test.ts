import { API_URL } from "../config";
import { resolveApiUri } from "../resolveApiUri";

describe("resolveApiUri", () => {
  it.each([undefined, null, "", "   "])("returns undefined for %p", (path) => {
    expect(resolveApiUri(path)).toBeUndefined();
  });

  it("prefixes a path relative to the API", () => {
    expect(resolveApiUri("/audio/cover?id=3")).toBe(`${API_URL}/audio/cover?id=3`);
  });

  it("adds the slash when the relative path has none", () => {
    expect(resolveApiUri("audio/cover?id=3")).toBe(`${API_URL}/audio/cover?id=3`);
  });

  it.each(["https://cdn.example.com/c.jpg", "http://10.0.2.2:8080/api/x", "file:///data/cache/c.jpg"])(
    "leaves a full address unchanged: %s",
    (uri) => {
      expect(resolveApiUri(uri)).toBe(uri);
    },
  );
});
