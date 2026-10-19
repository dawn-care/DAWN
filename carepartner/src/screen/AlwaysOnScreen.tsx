import React, { useMemo } from "react";
import { StyleSheet, View } from "react-native";
import { TouchArea } from "../telemetry/TouchTelemetry";
import { colors, space } from "../theme/theme";
import Grid from "../widgets/Grid";
import Orientation from "../widgets/renderers/Orientation";
import { DAY_TIME_PLACEMENT, placementsFor } from "./placements";
import { useScreen } from "./ScreenProvider";

/**
 * The always-on screen: the day and time in the top left, and the caregiver's
 * widgets around it on the grid. Nothing on it needs a tap. Until the first
 * screen arrives it shows the day, time of day and date on their own, so the
 * screen is never blank. Touches are recorded (see `TouchTelemetry`), on a
 * widget or an empty cell.
 */
export default function AlwaysOnScreen() {
  const { screen } = useScreen();
  const widgets = useMemo(() => (screen ? [DAY_TIME_PLACEMENT, ...placementsFor(screen)] : []), [screen]);

  const grid = widgets.length > 0;
  return (
    <TouchArea
      testID="always-on-screen"
      style={styles.root}
      screen={grid ? "always_on" : "default"}
      target={grid ? { type: "empty" } : null}
    >
      {grid ? (
        <Grid widgets={widgets} />
      ) : (
        <View testID="default-screen" style={styles.default}>
          <Orientation data={{ showDate: true }} />
        </View>
      )}
    </TouchArea>
  );
}

const styles = StyleSheet.create({
  root: { flex: 1, backgroundColor: colors.day.background },
  default: { flex: 1, padding: space.xl },
});
