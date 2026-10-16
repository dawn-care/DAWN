import { fitFontSize, GLYPH_WIDTH } from "../fitText";

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
