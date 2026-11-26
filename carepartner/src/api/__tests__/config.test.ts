import appConfig, { allowsPlainHttp, updatesUrl, webBaseUrl } from "../../../app.config";
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

describe("webBaseUrl", () => {
  it.each([undefined, "", "  ", "/"])("serves the web build at the root when unset: %p", (value) => {
    expect(webBaseUrl(value)).toBeUndefined();
  });

  it.each(["/carepartner", "/carepartner/", "carepartner"])("reads %p as /carepartner", (value) => {
    expect(webBaseUrl(value)).toBe("/carepartner");
  });
});

describe("updatesUrl", () => {
  it("is the EAS project's update address", () => {
    expect(updatesUrl("0f1e2d3c-aaaa-bbbb-cccc-123456789abc")).toBe("https://u.expo.dev/0f1e2d3c-aaaa-bbbb-cccc-123456789abc");
  });

  it.each([undefined, "", 7])("is missing without a project ID: %p", (id) => {
    expect(updatesUrl(id)).toBeUndefined();
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

  it("sets the web base path only when CAREPARTNER_WEB_BASE_URL is set", () => {
    const base = process.env.CAREPARTNER_WEB_BASE_URL;
    const make = () =>
      appConfig({ config: { name: "carepartner", slug: "carepartner", experiments: { typedRoutes: true } } } as never);
    try {
      delete process.env.CAREPARTNER_WEB_BASE_URL;
      expect(make().experiments).toEqual({ typedRoutes: true, baseUrl: undefined });

      process.env.CAREPARTNER_WEB_BASE_URL = "/carepartner";
      expect(make().experiments).toEqual({ typedRoutes: true, baseUrl: "/carepartner" });
    } finally {
      if (base === undefined) delete process.env.CAREPARTNER_WEB_BASE_URL;
      else process.env.CAREPARTNER_WEB_BASE_URL = base;
    }
  });

  it("sets the Android plain-HTTP permission from the configured server", () => {
    process.env.EXPO_PUBLIC_API_URL = "https://startup-da.gccis.rit.edu/api";
    expect(buildProperties()).toEqual(["expo-build-properties", { android: { usesCleartextTraffic: false } }]);

    delete process.env.EXPO_PUBLIC_API_URL;
    expect(buildProperties()).toEqual(["expo-build-properties", { android: { usesCleartextTraffic: true } }]);
  });

  it("asks the EAS project for updates once app.json has its ID", () => {
    const config = appConfig({
      config: { name: "carepartner", slug: "carepartner", extra: { eas: { projectId: "abc" } }, updates: { fallbackToCacheTimeout: 0 } },
    } as never);
    expect(config.updates).toEqual({ fallbackToCacheTimeout: 0, url: "https://u.expo.dev/abc" });
  });

  it("leaves updates alone without a project ID", () => {
    const config = appConfig({ config: { name: "carepartner", slug: "carepartner" } } as never);
    expect(config.updates).toBeUndefined();
  });
});
