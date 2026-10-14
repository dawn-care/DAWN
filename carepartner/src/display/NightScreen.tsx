import React from "react";
import { StyleSheet, View } from "react-native";
import { colors } from "../theme/theme";

/**
 * The screen shown at night. A plain dark screen for now; the real night
 * screen (dim and warm, with the day and one reassuring line) is #314.
 */
export default function NightScreen() {
  return <View testID="night-screen" style={styles.root} />;
}

const styles = StyleSheet.create({
  root: { flex: 1, backgroundColor: colors.player.background },
});
