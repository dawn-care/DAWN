/**
 * pairingErrors.ts
 *
 * User: Caregiver
 *
 * The errors the server's POST /users/device-pairing/verify can answer with
 * (VerifyPairingCode in device_pairing_service.go), and the wording the
 * caregiver sees for each. If the server's text changes, change it here only.
 */

export const PAIRING_SERVER_ERROR = {
  NOT_FOUND: "pairing code not found",
  ALREADY_USED: "pairing code already used",
  EXPIRED: "pairing code expired",
} as const;

export const PAIRING_ERROR_FALLBACK = "Couldn't pair the tablet. Please try again.";

const PAIRING_ERROR_WORDING: Record<string, string> = {
  [PAIRING_SERVER_ERROR.NOT_FOUND]: "That code didn't work. Check the code on the tablet and try again.",
  [PAIRING_SERVER_ERROR.ALREADY_USED]: "That code has already been used. Get a new code on the tablet.",
  [PAIRING_SERVER_ERROR.EXPIRED]: "That code has expired. Get a new code on the tablet.",
};

/** The caregiver's wording for a server error message */
export const pairingErrorWording = (serverMessage: string): string =>
  PAIRING_ERROR_WORDING[serverMessage] ?? PAIRING_ERROR_FALLBACK;
