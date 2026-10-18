import React, { useMemo } from "react";
import { StyleSheet, View } from "react-native";
import { colors, space } from "../theme/theme";
import Grid from "../widgets/Grid";
import Orientation from "../widgets/renderers/Orientation";
import { placementsFor } from "./placements";
import { useScreen } from "./ScreenProvider";

/**
 * The always-on screen: the caregiver's widgets on the grid. Nothing on it
 * needs a tap. When there is nothing to draw (before the first screen
 * arrives, with no widgets, or only ones the tablet can't draw) it shows the
 * day, time of day and date, so the screen is never blank.
 */
export default function AlwaysOnScreen() {
  const { screen } = useScreen();
  const widgets = useMemo(() => (screen ? placementsFor(screen) : []), [screen]);

  return (
    <View testID="always-on-screen" style={styles.root}>
      {widgets.length ? (
        <Grid widgets={widgets} />
      ) : (
        <View testID="default-screen" style={styles.default}>
          <Orientation data={{ showDate: true }} />
        </View>
      )}
    </View>
  );
}

const styles = StyleSheet.create({
  root: { flex: 1, backgroundColor: colors.day.background },
  default: { flex: 1, padding: space.xl },
});
