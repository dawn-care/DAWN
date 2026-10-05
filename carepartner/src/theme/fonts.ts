import {
  AtkinsonHyperlegibleNext_400Regular,
  AtkinsonHyperlegibleNext_700Bold,
  AtkinsonHyperlegibleNext_800ExtraBold,
  useFonts,
} from "@expo-google-fonts/atkinson-hyperlegible-next";

/**
 * Font family names for styles. On Android a custom font's weight comes from
 * its family name, not from `fontWeight`, so use the family for the weight you
 * want instead of setting `fontWeight`.
 */
export const fonts = {
  regular: "AtkinsonHyperlegibleNext_400Regular",
  bold: "AtkinsonHyperlegibleNext_700Bold",
  heavy: "AtkinsonHyperlegibleNext_800ExtraBold",
} as const;

/** The font files to load, keyed by the family names above. */
export const fontFiles = {
  [fonts.regular]: AtkinsonHyperlegibleNext_400Regular,
  [fonts.bold]: AtkinsonHyperlegibleNext_700Bold,
  [fonts.heavy]: AtkinsonHyperlegibleNext_800ExtraBold,
};

/** Loads the app's fonts. Returns [loaded, error], as expo-font's useFonts does. */
export function useAppFonts() {
  return useFonts(fontFiles);
}
