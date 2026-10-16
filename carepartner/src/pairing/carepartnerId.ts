import { Platform } from "react-native";
import * as SecureStore from "expo-secure-store";

/** Where the paired carepartner's id is stored, on web and on native. */
const STORAGE_KEY = "carepartnerID";

export async function getCarepartnerId(): Promise<number | null> {
  if (Platform.OS === "web") {
    const id = localStorage.getItem(STORAGE_KEY);
    return id ? parseInt(id, 10) : null;
  }

  const id = await SecureStore.getItemAsync(STORAGE_KEY);
  return id ? parseInt(id, 10) : null;
}

export async function storeCarepartnerId(id: number): Promise<void> {
  if (Platform.OS === "web") {
    localStorage.setItem(STORAGE_KEY, id.toString());
    return;
  }

  await SecureStore.setItemAsync(STORAGE_KEY, id.toString());
}