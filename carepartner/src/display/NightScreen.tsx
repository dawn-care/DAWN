import React from "react";
import { StyleSheet, Text, View } from "react-native";
import { dayName } from "@shared";
import { useNow } from "../clock/useNow";
import { fonts } from "../theme/fonts";
import { colors, fontSizes, space } from "../theme/theme";

/** The one line the night screen shows under the day. */
export const NIGHT_LINE = "You're safe. Sleep well.";

/** The screen shown at night: dark and warm, with the day and one reassuring line. */
export default function NightScreen() {
  const now = useNow();
  return (
    <View testID="night-screen" style={styles.root}>
      <Text style={styles.day}>{dayName(now)}</Text>
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
  line: { fontFamily: fonts.regular, fontSize: fontSizes.nightLine, color: colors.night.textSecondary, textAlign: "center" },
});
