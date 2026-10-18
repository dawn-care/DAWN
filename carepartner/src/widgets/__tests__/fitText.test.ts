import { fitFontSize, fitFontSizeToBox, GLYPH_WIDTH, LINE_HEIGHT, linesAt } from "../fitText";
import { TEXT_LIMITS } from "@shared";
import { fontSizes, legibility, space } from "../../theme/theme";

describe("fitFontSize", () => {
  it("keeps the full size when the text fits", () => {
    expect(fitFontSize("Monday", 1000, 152, 76)).toBe(152);
  });

  it("shrinks long text to fit the width", () => {
    const size = fitFontSize("Wednesday", 576, 152, 76);

    expect(size).toBeLessThan(152);
    expect(size * GLYPH_WIDTH * "Wednesday".length).toBeLessThanOrEqual(576);
  });

  it("never goes below the floor", () => {
    expect(fitFontSize("Wednesday", 100, 152, 76)).toBe(76);
  });

  it("uses the full size before the width is known", () => {
    expect(fitFontSize("Wednesday", 0, 152, 76)).toBe(152);
  });
});

describe("fitFontSizeToBox", () => {
  it("keeps the full size when the text fits on one line", () => {
    expect(fitFontSizeToBox("Love you", 600, 200, 42, 26)).toBe(42);
  });

  it("shrinks text so its wrapped lines fit the height", () => {
    const text = "Remember to drink some water after your walk this afternoon";
    const size = fitFontSizeToBox(text, 300, 150, 42, 26);
    const perLine = Math.floor(300 / (size * GLYPH_WIDTH));

    expect(size).toBeLessThan(42);
    expect(Math.ceil(text.length / perLine) * size * LINE_HEIGHT).toBeLessThanOrEqual(150);
  });

  it("never goes below the floor", () => {
    expect(fitFontSizeToBox("word ".repeat(100), 200, 60, 42, 26)).toBe(26);
  });

  it("uses the full size before the box is measured", () => {
    expect(fitFontSizeToBox("Love you", 0, 0, 42, 26)).toBe(42);
  });
});

describe("linesAt", () => {
  it("counts the whole lines that fit, at least one", () => {
    expect(linesAt(25, 100)).toBe(3);
    expect(linesAt(40, 10)).toBe(1);
  });
});

describe("TEXT_LIMITS", () => {
  // Characters that fit a box at the smallest text size, counting whole lines.
  const capacity = (width: number, height: number) =>
    Math.floor(width / (legibility.minText * GLYPH_WIDTH)) * linesAt(legibility.minText, height);

  // Boxes measured on the 1280x800 design, inside a 2x1 tile.
  it("leaves room for the longest message beside a photo and under a label", () => {
    const labelHeight = fontSizes.label * LINE_HEIGHT + space.xs / 2;
    expect(capacity(343, 167 - labelHeight)).toBeGreaterThanOrEqual(TEXT_LIMITS.message);
  });

  it("leaves room for the longest reminder on two lines beside its icon and time", () => {
    expect(capacity(300, 2 * legibility.minText * LINE_HEIGHT)).toBeGreaterThanOrEqual(TEXT_LIMITS.reminder);
  });
});
