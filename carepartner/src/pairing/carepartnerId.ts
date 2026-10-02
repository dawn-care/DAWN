import { Platform } from "react-native";
import * as SecureStore from "expo-secure-store";

export async function getCarepartnerId(): Promise<number | null> {
  if (Platform.OS === "web") {
    const id = localStorage.getItem("carepartnerID");
    return id ? parseInt(id, 10) : null;
  }

  const id = await SecureStore.getItemAsync("carepartnerID");
  return id ? parseInt(id, 10) : null;
}

export async function storeCarepartnerId(id: number): Promise<void> {
  if (Platform.OS === "web") {
    localStorage.setItem("carepartnerID", id.toString());
    return;
  }

  await SecureStore.setItemAsync("carepartnerID", id.toString());
}