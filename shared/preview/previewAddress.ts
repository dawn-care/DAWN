/**
 * previewAddress.ts
 *
 * The address the caregiver app opens the carepartner web app at to show one
 * carepartner's screen, `?preview=1&carepartner_id=N`. The caregiver builds it
 * and the carepartner reads it, so both use these names: if they ever differed,
 * the iframe would run as a real tablet (pairing, play reports, touches).
 */

export const PREVIEW_ADDRESS = {
  /** Present with the value `on` in a preview */
  flag: "preview",
  on: "1",
  /** The carepartner whose screen to show */
  carepartnerId: "carepartner_id",
} as const;
