import { getStoredItem, setStoredItem } from "../storage/storage";

/** Where the paired carepartner's id is stored, on web and on native. */
const STORAGE_KEY = "carepartnerID";

/**
 * Whether a caregiver has claimed that id. The pair screen stores the id
 * before pairing finishes, so having one doesn't mean the tablet is paired.
 */
const PAIRED_KEY = "pairingComplete";

export async function getCarepartnerId(): Promise<number | null> {
  const id = await getStoredItem(STORAGE_KEY);
  return id ? parseInt(id, 10) : null;
}

/** Stores a new id, not yet paired. */
export async function storeCarepartnerId(id: number): Promise<void> {
  await setStoredItem(PAIRED_KEY, "false");
  await setStoredItem(STORAGE_KEY, id.toString());
}

export async function isPairingComplete(): Promise<boolean> {
  return (await getStoredItem(PAIRED_KEY)) === "true";
}

/** Records that a caregiver has claimed the stored id. */
export async function markPairingComplete(): Promise<void> {
  await setStoredItem(PAIRED_KEY, "true");
}
