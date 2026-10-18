import React from "react";
import { StyleSheet, Text, View } from "react-native";
import { fonts } from "../theme/fonts";
import { colors, fontSizes, space } from "../theme/theme";

/** Shown in preview mode when the address names no carepartner, in place of the pair screen. */
export default function PreviewUnavailable() {
  return (
    <View testID="preview-unavailable" style={styles.root}>
      <Text style={styles.line}>No screen to show.</Text>
    </View>
  );
}

const styles = StyleSheet.create({
  root: {
    flex: 1,
    alignItems: "center",
    justifyContent: "center",
    padding: space.lg,
    backgroundColor: colors.day.background,
  },
  line: { fontFamily: fonts.bold, fontSize: fontSizes.name, color: colors.day.textSecondary, textAlign: "center" },
});
