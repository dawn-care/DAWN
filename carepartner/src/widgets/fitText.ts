/**
 * Average glyph width as a share of the font size for Atkinson Hyperlegible
 * Next bold and extra bold, measured from the day name at 152px. Errs wide so
 * text never overflows.
 */
export const GLYPH_WIDTH = 0.64;

/**
 * The largest font size, up to `max`, at which `text` fits in `width`, but
 * never below `min` (a legibility floor). Works on every platform, unlike
 * React Native's adjustsFontSizeToFit, which the web build ignores.
 */
export function fitFontSize(text: string, width: number, max: number, min: number): number {
  if (!text.length || width <= 0) return max;
  const fitting = width / (text.length * GLYPH_WIDTH);
  return Math.max(min, Math.min(max, Math.floor(fitting)));
}
