import React from "react";
import { StyleSheet, Text, View } from "react-native";
import { fonts } from "../theme/fonts";
import { colors, fontSizes, space } from "../theme/theme";

/**
 * Shown in preview mode in place of a screen: when the address names no
 * carepartner (instead of the pair screen), or when their screen couldn't be loaded.
 */
export default function PreviewUnavailable({ message = "No screen to show." }: { message?: string }) {
  return (
    <View testID="preview-unavailable" style={styles.root}>
      <Text style={styles.line}>{message}</Text>
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
