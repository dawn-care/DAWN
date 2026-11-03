import { fitCaption, TEXT_LIMITS } from "../textLimits";

describe("fitCaption", () => {
  const limit = TEXT_LIMITS.photoCaption;

  it("returns a caption within the limit as it is", () => {
    expect(fitCaption("ny_giants_logo")).toBe("ny_giants_logo");
    expect(fitCaption("a".repeat(limit))).toBe("a".repeat(limit));
  });

  it("returns an empty caption as it is", () => {
    expect(fitCaption("")).toBe("");
  });

  it("cuts a longer caption to the limit, ending in an ellipsis", () => {
    const cut = fitCaption("a".repeat(limit + 1));

    expect(cut).toBe("a".repeat(limit - 1) + "…");
    expect(cut).toHaveLength(limit);
  });

  it("cuts a very long caption the same way", () => {
    expect(fitCaption("b".repeat(200))).toBe("b".repeat(limit - 1) + "…");
  });

  it("drops a space left before the ellipsis", () => {
    expect(fitCaption("x".repeat(limit - 2) + " " + "y".repeat(10))).toBe("x".repeat(limit - 2) + "…");
  });

  it("does not leave half of a two-unit character", () => {
    const cut = fitCaption("a".repeat(limit - 2) + "😀😀");

    expect(cut).toBe("a".repeat(limit - 2) + "…");
    expect(cut).not.toMatch(/[\ud800-\udbff](?![\udc00-\udfff])/);
  });

  it("never returns more than the limit", () => {
    for (const text of ["a".repeat(limit + 5), "😀".repeat(limit), "word ".repeat(limit)]) {
      expect(fitCaption(text).length).toBeLessThanOrEqual(limit);
    }
  });
});
