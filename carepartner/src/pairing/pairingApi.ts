/**
 * pairingApi.ts
 *
 * The pair screen's requests, with their responses checked where they enter
 * the app. Logic only. Whether pairing has completed is `fetchPairingStatus`
 * in `pairingState.ts`.
 */

import { getJson, HttpError, postJson } from "../api/request";
import { now } from "../clock/clock";

/** A code a caregiver enters in the family app, and when it stops working. */
export type PairingCode = { code: string; expiresAt: Date };

const isObject = (value: unknown): value is Record<string, unknown> => typeof value === "object" && value !== null;

/**
 * Whether a carepartner with this id exists on the server. Only a 404 counts
 * as missing: any other failure throws, so a tablet that is briefly offline
 * never decides its resident is gone and creates a duplicate.
 */
export async function findCarepartner(id: number): Promise<boolean> {
  try {
    await getJson(`/users/carepartner/${id}`);
    return true;
  } catch (error) {
    if (error instanceof HttpError && error.status === 404) return false;
    throw error;
  }
}

/**
 * Creates a carepartner for this tablet and returns its id. It has no name:
 * the server gives it a placeholder, and the caregiver's name for the
 * resident replaces that when they enter the pairing code.
 */
export async function createCarepartner(): Promise<number> {
  const created = await postJson("/users/carepartner/create-carepartner", {});
  if (!isObject(created) || typeof created.carepartner_id !== "number") {
    throw new Error("Create carepartner response is malformed");
  }
  return created.carepartner_id;
}

/** Asks the server for a new pairing code for this carepartner. */
export async function requestPairingCode(carepartnerId: number): Promise<PairingCode> {
  const response = await postJson("/users/device-pairing/code", { carepartner_id: carepartnerId });
  if (!isObject(response) || typeof response.pairing_code !== "string" || typeof response.expires_in_sec !== "number") {
    throw new Error("Pairing code response is malformed");
  }
  return {
    code: response.pairing_code,
    expiresAt: new Date(now().getTime() + response.expires_in_sec * 1000),
  };
}
