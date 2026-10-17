/**
 * textLimits.ts
 *
 * The longest text each widget can show whole on the tablet at its fixed size
 * (`WIDGET_SIZES`), at the smallest text size the design allows. The
 * caregiver's editor should stop longer text; the tablet cuts it off. The
 * server allows more today.
 */

export const TEXT_LIMITS = {
  /** A message's text, with a photo and a label beside it. */
  message: 80,
  /** One reminder's text, beside its icon and time. */
  reminder: 32,
} as const;
