/**
 * textLimits.ts
 *
 * The longest text each widget can show whole on the tablet at its fixed size
 * (`WIDGET_SIZES`), at the smallest text size the design allows. The
 * caregiver's editor should stop longer text; the tablet cuts it off. The
 * server allows more today.
 */

export const TEXT_LIMITS = {
  /**
   * A message's text, with a photo and a label beside it. About 80 fit by
   * average glyph width; this leaves room for lines that break early.
   */
  message: 64,
  /** One reminder's text, beside its icon and time. */
  reminder: 32,
  /**
   * A photo's caption, on one line across the bottom of its 2x2 tile at the
   * name size. The tile is 576 wide on the 1280x800 design, less 20 of padding
   * each side; this is how many average glyphs fit in what is left. The server
   * enforces the same number (`MaxPhotoCaption` in screen_widgets/config.go).
   */
  photoCaption: 20,
} as const;

/**
 * A photo's caption as the screen shows it: text longer than the limit is cut
 * and ends in "…", so the result is never longer than `TEXT_LIMITS.photoCaption`.
 * Text within the limit is returned as it is. Counts as a text field's
 * `maxLength` does, and never leaves half of a character that takes two units.
 */
export function fitCaption(text: string): string {
  const limit = TEXT_LIMITS.photoCaption;
  if (text.length <= limit) return text;

  let end = limit - 1;
  const last = text.charCodeAt(end - 1);
  if (last >= 0xd800 && last <= 0xdbff) end -= 1;
  return text.slice(0, end).replace(/\s+$/, "") + "…";
}
