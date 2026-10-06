import appConfig, { allowsPlainHttp } from "../../../app.config";
import { localApiUrl } from "../config";

describe("localApiUrl", () => {
  it("uses 10.0.2.2 on Android, where the emulator reaches the computer", () => {
    expect(localApiUrl("android")).toBe("http://10.0.2.2:8080/api");
  });

  it.each(["web", "ios"])("uses localhost on %s", (os) => {
    expect(localApiUrl(os)).toBe("http://localhost:8080/api");
  });
});

describe("allowsPlainHttp", () => {
  it.each([undefined, "", "http://10.0.2.2:8080/api", "HTTP://192.168.1.20:8080/api"])(
    "allows plain HTTP for a local server: %p",
    (url) => {
      expect(allowsPlainHttp(url)).toBe(true);
    },
  );

  it("keeps plain HTTP blocked for an https server", () => {
    expect(allowsPlainHttp("https://startup-da.gccis.rit.edu/api")).toBe(false);
  });
});

describe("app config", () => {
  const original = process.env.EXPO_PUBLIC_API_URL;

  afterEach(() => {
    if (original === undefined) delete process.env.EXPO_PUBLIC_API_URL;
    else process.env.EXPO_PUBLIC_API_URL = original;
  });

  const buildProperties = () => {
    const config = appConfig({ config: { name: "carepartner", slug: "carepartner", plugins: ["expo-router"] } } as never);
    return config.plugins?.find((p) => Array.isArray(p) && p[0] === "expo-build-properties");
  };

  it("keeps the plugins from app.json", () => {
    delete process.env.EXPO_PUBLIC_API_URL;
    const config = appConfig({ config: { name: "carepartner", slug: "carepartner", plugins: ["expo-router"] } } as never);

    expect(config.plugins?.[0]).toBe("expo-router");
  });

  it("sets the Android plain-HTTP permission from the configured server", () => {
    process.env.EXPO_PUBLIC_API_URL = "https://startup-da.gccis.rit.edu/api";
    expect(buildProperties()).toEqual(["expo-build-properties", { android: { usesCleartextTraffic: false } }]);

    delete process.env.EXPO_PUBLIC_API_URL;
    expect(buildProperties()).toEqual(["expo-build-properties", { android: { usesCleartextTraffic: true } }]);
  });
});
