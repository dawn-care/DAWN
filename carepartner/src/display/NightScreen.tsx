import React from "react";
import { StyleSheet, Text, View } from "react-native";
import { MaterialCommunityIcons } from "@expo/vector-icons";
import { dayName } from "@shared";
import { useNow } from "../clock/useNow";
import { fonts } from "../theme/fonts";
import { colors, fontSizes, space } from "../theme/theme";

/** The one line the night screen shows under the day. */
export const NIGHT_LINE = "You're safe. Sleep well.";
export const NIGHT_WORD = "Nighttime";

/** The screen shown at night: dark and warm, with the day, a moon and "Nighttime", and one reassuring line. */
export default function NightScreen() {
  const now = useNow();
  return (
    <View testID="night-screen" style={styles.root}>
      <Text style={styles.day}>{dayName(now)}</Text>
      <View style={styles.nightRow}>
        <MaterialCommunityIcons testID="night-moon" name="weather-night" size={fontSizes.nightWord} color={colors.night.text} />
        <Text style={styles.nightWord}>{NIGHT_WORD}</Text>
      </View>
      <Text style={styles.line}>{NIGHT_LINE}</Text>
    </View>
  );
}

const styles = StyleSheet.create({
  root: {
    flex: 1,
    alignItems: "center",
    justifyContent: "center",
    gap: space.md,
    padding: space.lg,
    backgroundColor: colors.night.background,
  },
  day: { fontFamily: fonts.heavy, fontSize: fontSizes.nightDay, color: colors.night.text, textAlign: "center" },
  nightRow: { flexDirection: "row", alignItems: "center", gap: space.sm },
  nightWord: { fontFamily: fonts.bold, fontSize: fontSizes.nightWord, color: colors.night.text, textAlign: "center" },
  line: { fontFamily: fonts.regular, fontSize: fontSizes.nightLine, color: colors.night.textSecondary, textAlign: "center" },
});
