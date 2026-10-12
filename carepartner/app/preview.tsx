import React from "react";
import { StyleSheet, View } from "react-native";
import Grid from "../src/widgets/Grid";
import { sampleScreen } from "../src/widgets/sampleScreen";
import { colors } from "../src/theme/theme";

/**
 * Temporary: the always-on screen's widgets on sample data, for review. Open
 * at /preview on web or carepartner://preview on Android. Goes away when the
 * always-on screen becomes the home screen.
 */
export default function Preview() {
  return (
    <View style={styles.root}>
      <Grid widgets={sampleScreen} />
    </View>
  );
}

const styles = StyleSheet.create({
  root: { flex: 1, backgroundColor: colors.day.background },
});
