/**
 * pairingState.ts
 *
 * Whether the tablet is paired: it has a stored carepartner id and a
 * caregiver has claimed it (`/users/device-pairing/status`). Logic only.
 */

import { postJson } from "../api/request";
import { getCarepartnerId, isPairingComplete, markPairingComplete } from "./carepartnerId";

/** Whether a caregiver has claimed this carepartner. Throws if the server can't say. */
export async function fetchPairingStatus(carepartnerId: number): Promise<boolean> {
  const status = await postJson("/users/device-pairing/status", { carepartner_id: carepartnerId });
  if (typeof status !== "object" || status === null || typeof (status as { paired?: unknown }).paired !== "boolean") {
    throw new Error("Pairing status response is malformed");
  }
  return (status as { paired: boolean }).paired;
}

/**
 * The paired carepartner's id, or null if the tablet isn't paired.
 *
 * Tablets that paired before pairing was recorded have an id but no record,
 * so for an id without one the server is asked once. If it can't be reached,
 * the tablet keeps counting as paired, so a working tablet isn't sent back to
 * pairing over a dropped connection, and asks again next time.
 */
export async function readPairedId(): Promise<number | null> {
  const id = await getCarepartnerId();
  if (id == null) return null;
  if (await isPairingComplete()) return id;
  try {
    if (!(await fetchPairingStatus(id))) return null;
    await markPairingComplete();
    return id;
  } catch (error) {
    console.warn("Couldn't check whether this tablet is paired", error);
    return id;
  }
}
