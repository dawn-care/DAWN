/**
 * theme.ts
 *
 * Palette, type sizes and spacing for the tablet, taken from
 * mockups/DAWN Carepartner UI.html (drawn at 1280×800).
 */

export const colors = {
  day: {
    background: "#F2EADB",
    surface: "#FFFFFF",
    text: "#1E1A15",
    textSecondary: "#4A4238",
    border: "#8A7B66",
    accent: "#9A3412",
    placeholder: "#E8DDCB",
    placeholderIcon: "#6B5E4E",
    positive: "#2F5D3A",
    photoCaption: "rgba(30,26,21,0.84)",
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

/** Legibility floors from NFR-D in docs/features.md. */
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

export const fontWeights = {
  bold: "700",
  heavy: "800",
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
