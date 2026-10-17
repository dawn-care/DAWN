import { Platform } from "react-native";
import * as SecureStore from "expo-secure-store";

/**
 * Small values kept on the device across restarts: `localStorage` on web,
 * `expo-secure-store` on native.
 */
export async function getStoredItem(key: string): Promise<string | null> {
  if (Platform.OS === "web") return localStorage.getItem(key);
  return SecureStore.getItemAsync(key);
}

export async function setStoredItem(key: string, value: string): Promise<void> {
  if (Platform.OS === "web") {
    localStorage.setItem(key, value);
    return;
  }
  await SecureStore.setItemAsync(key, value);
}
