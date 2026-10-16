import React from "react";
import { StyleSheet, View } from "react-native";
import { borders, colors, radii, space } from "../theme/theme";
import type { Rect } from "./layout";

type Props = {
  rect: Rect;
  /** Draw the tile background and border; off for widgets that sit on the screen background. */
  framed: boolean;
  testID?: string;
  children: React.ReactNode;
};

/** Positions a widget on the grid and, if framed, draws the shared tile look around it. */
export default function WidgetFrame({ rect, framed, testID, children }: Props) {
  return (
    <View testID={testID} style={[styles.base, rect, framed && styles.framed]}>
      {children}
    </View>
  );
}

const styles = StyleSheet.create({
  base: { position: "absolute" },
  framed: {
    backgroundColor: colors.day.surface,
    borderWidth: borders.tile,
    borderColor: colors.day.border,
    borderRadius: radii.tile,
    padding: space.sm,
  },
});
