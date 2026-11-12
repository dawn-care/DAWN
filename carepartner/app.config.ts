import type { ConfigContext, ExpoConfig } from "expo/config";

/**
 * Android release builds block plain http:// connections. Allow them only
 * when the build talks to an http:// server: local development, which is the
 * default when EXPO_PUBLIC_API_URL is unset. A build pointed at an https://
 * server keeps them blocked.
 */
export function allowsPlainHttp(apiUrl: string | undefined): boolean {
  return !apiUrl || apiUrl.trim().toLowerCase().startsWith("http://");
}

/**
 * The path the web build is served under, from CAREPARTNER_WEB_BASE_URL (the
 * server image sets `/carepartner`, see server/Dockerfile). Unset everywhere
 * else, so `npm run web` stays at `/` and the Android app is untouched.
 */
export function webBaseUrl(value: string | undefined): string | undefined {
  const path = value?.trim().replace(/\/+$/, "");
  return path ? (path.startsWith("/") ? path : `/${path}`) : undefined;
}

// Adds build settings that depend on the environment to the static app.json.
export default ({ config }: ConfigContext): ExpoConfig =>
  ({
    ...config,
    experiments: { ...config.experiments, baseUrl: webBaseUrl(process.env.CAREPARTNER_WEB_BASE_URL) },
    plugins: [
      ...(config.plugins ?? []),
      [
        "expo-build-properties",
        { android: { usesCleartextTraffic: allowsPlainHttp(process.env.EXPO_PUBLIC_API_URL) } },
      ],
    ],
  }) as ExpoConfig;
