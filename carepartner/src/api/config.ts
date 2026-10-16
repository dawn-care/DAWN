import { Platform } from "react-native";

/**
 * The local development backend's address for a platform. The Android
 * emulator reaches the computer it runs on at 10.0.2.2, not localhost.
 */
export function localApiUrl(os: string): string {
  return os === "android" ? "http://10.0.2.2:8080/api" : "http://localhost:8080/api";
}

// Set EXPO_PUBLIC_API_URL in .env.local to use any other backend, including
// from a real tablet, which needs the computer's network address.
export const API_URL = process.env.EXPO_PUBLIC_API_URL ?? localApiUrl(Platform.OS);
