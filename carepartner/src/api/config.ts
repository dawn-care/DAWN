// Set EXPO_PUBLIC_API_URL in .env.local to point a build at another backend.
// On an Android emulator the dev machine is 10.0.2.2, not localhost.
export const API_URL = process.env.EXPO_PUBLIC_API_URL ?? "http://localhost:8080/api";
