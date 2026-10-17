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

/**
 * The largest font size, up to `max`, at which `text` wrapped over several
 * lines fits a `width` × `height` box, but never below `min`. Counts whole
 * lines of average glyphs, so it errs large for text with long words; cap the
 * lines with `linesAt` so anything left over is cut off rather than overflow.
 */
export function fitFontSizeToBox(text: string, width: number, height: number, max: number, min: number): number {
  if (!text.length || width <= 0 || height <= 0) return max;
  for (let size = max; size > min; size--) {
    if (linesNeeded(text, width, size) * size * LINE_HEIGHT <= height) return size;
  }
  return min;
}

/** How many lines of text at `size` fit in `height`; at least one. */
export function linesAt(size: number, height: number): number {
  return Math.max(1, Math.floor(height / (size * LINE_HEIGHT)));
}

/** Line height as a share of the font size, for text set with `fitFontSizeToBox`. */
export const LINE_HEIGHT = 1.2;

function linesNeeded(text: string, width: number, size: number): number {
  const perLine = Math.max(1, Math.floor(width / (size * GLYPH_WIDTH)));
  return Math.ceil(text.length / perLine);
}
