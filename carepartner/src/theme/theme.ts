/**
 * theme.ts
 *
 * Palette, type sizes and spacing for the tablet, taken from the
 * visual design (drawn at 1280×800).
 *
 * The day colours follow the D.A.W.N. brand palette the staff and family apps
 * use (night #1E2140, cream #FFF8EF, dusk #5B6091, ember #B4533A). Where a
 * brand token falls below the tablet's stricter floors it is darkened or
 * swapped: secondary text is dusk darkened toward night (dusk itself is 5.6:1
 * on cream, under the 7:1 floor), tile borders are dusk (the brand's border is
 * nearly invisible on a white tile), and the accent is ember (apricot is too
 * pale for a marker). The night colours stay warm, for sleep.
 */

export const colors = {
  day: {
    background: "#FFF8EF",
    surface: "#FFFFFF",
    text: "#1E2140",
    textSecondary: "#494D79",
    border: "#5B6091",
    accent: "#B4533A",
    placeholder: "#EDE3D6",
    placeholderIcon: "#5B6091",
    positive: "#2F5D3A",
    photoCaption: "rgba(30,33,64,0.84)",
  },
  night: {
    background: "#15110D",
    text: "#E8DCC8",
    textSecondary: "#BFB19C",
    accent: "#C9A56A",
  },
  player: {
    background: "#2A241D",
    text: "#FFFFFF",
    senderPill: "rgba(0,0,0,0.72)",
    captionBar: "rgba(0,0,0,0.82)",
  },
} as const;

/** Legibility floors every tablet screen must meet. */
export const legibility = {
  minOrientationText: 76,
  minText: 26,
  minContrast: 7,
} as const;

export const fontSizes = {
  day: 152,
  nightDay: 120,
  timeOfDay: 88,
  nightLine: 56,
  tileWord: 44,
  caption: 42,
  name: 40,
  label: 28,
  small: 26,
} as const;

export const space = {
  xs: 12,
  sm: 20,
  md: 32,
  lg: 48,
  xl: 80,
} as const;

export const radii = {
  tile: 24,
  pill: 999,
} as const;

export const borders = {
  tile: 3,
  accent: 4,
} as const;
