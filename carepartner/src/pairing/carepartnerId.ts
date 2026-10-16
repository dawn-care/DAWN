import { getStoredItem, setStoredItem } from "../storage/storage";

/** Where the paired carepartner's id is stored, on web and on native. */
const STORAGE_KEY = "carepartnerID";

export async function getCarepartnerId(): Promise<number | null> {
  const id = await getStoredItem(STORAGE_KEY);
  return id ? parseInt(id, 10) : null;
}

export async function storeCarepartnerId(id: number): Promise<void> {
  await setStoredItem(STORAGE_KEY, id.toString());
}
