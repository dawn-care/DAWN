/**
 * styles.ts
 *
 * User: Caregiver
 *
 * Tailwind class strings shared by the form controls and buttons in
 * components/ui, so the focus ring and input look stay identical.
 */

// Visible keyboard focus ring for every interactive control
export const FOCUS_RING =
  "focus-visible:outline focus-visible:outline-[3px] focus-visible:outline-offset-2 focus-visible:outline-dawn-accent";

// Bordered, 44px-tall field used by text inputs and selects
export const FIELD_CLASSES = `h-11 px-3 rounded-[10px] border-2 border-dawn-control bg-white text-base text-dawn-ink ${FOCUS_RING}`;
