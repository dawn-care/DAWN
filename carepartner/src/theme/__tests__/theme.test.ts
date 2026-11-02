import { colors, fontSizes, legibility } from "../theme";

/** WCAG relative luminance of a #RRGGBB colour. */
function luminance(hex: string) {
  const [r, g, b] = [1, 3, 5].map((i) => {
    const c = parseInt(hex.slice(i, i + 2), 16) / 255;
    return c <= 0.03928 ? c / 12.92 : ((c + 0.055) / 1.055) ** 2.4;
  });
  return 0.2126 * r + 0.7152 * g + 0.0722 * b;
}

function contrast(a: string, b: string) {
  const [hi, lo] = [luminance(a), luminance(b)].sort((x, y) => y - x);
  return (hi + 0.05) / (lo + 0.05);
}

describe("theme legibility", () => {
  it.each([
    ["day text on background", colors.day.text, colors.day.background],
    ["day text on surface", colors.day.text, colors.day.surface],
    ["day secondary text on background", colors.day.textSecondary, colors.day.background],
    ["day secondary text on surface", colors.day.textSecondary, colors.day.surface],
    ["night text on background", colors.night.text, colors.night.background],
    ["night secondary text on background", colors.night.textSecondary, colors.night.background],
    ["player text on background", colors.player.text, colors.player.background],
  ])("%s meets the contrast floor", (_name, foreground, background) => {
    expect(contrast(foreground, background)).toBeGreaterThanOrEqual(legibility.minContrast);
  });

  it.each([
    ["tile border on surface", colors.day.border, colors.day.surface],
    ["tile border on background", colors.day.border, colors.day.background],
    ["accent marker on surface", colors.day.accent, colors.day.surface],
    ["photo placeholder icon on placeholder", colors.day.placeholderIcon, colors.day.placeholder],
  ])("%s stands out (3:1, for shapes that aren't text)", (_name, foreground, background) => {
    expect(contrast(foreground, background)).toBeGreaterThanOrEqual(3);
  });

  it("keeps every text size at or above the floor", () => {
    for (const size of Object.values(fontSizes)) {
      expect(size).toBeGreaterThanOrEqual(legibility.minText);
    }
  });

  it("keeps orientation text at or above its floor", () => {
    for (const size of [fontSizes.day, fontSizes.nightDay, fontSizes.nightWord, fontSizes.timeOfDay]) {
      expect(size).toBeGreaterThanOrEqual(legibility.minOrientationText);
    }
  });
});
