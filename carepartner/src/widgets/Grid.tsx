import React, { useState } from "react";
import { LayoutChangeEvent, StyleSheet, View } from "react-native";
import { fitsGrid } from "@shared";
import { cellRect, type Size } from "./layout";
import { lookupRenderer } from "./registry";
import type { WidgetPlacement } from "./types";
import WidgetFrame from "./WidgetFrame";

type Props = {
  widgets: WidgetPlacement[];
  /** Screen size to lay out against before the first layout pass reports it (useful in tests). */
  initialSize?: Size;
};

/**
 * Draws widgets at their grid positions, sized from the space this view gets.
 * Widgets with an unknown renderer, or that don't fit the grid, are skipped.
 */
export default function Grid({ widgets, initialSize }: Props) {
  const [size, setSize] = useState<Size | undefined>(initialSize);

  const onLayout = (e: LayoutChangeEvent) => {
    const { width, height } = e.nativeEvent.layout;
    setSize({ width, height });
  };

  return (
    <View style={styles.root} onLayout={onLayout}>
      {size &&
        widgets.map((widget) => {
          const entry = lookupRenderer(widget.renderer);
          if (!entry || !fitsGrid(widget)) return null;
          const Widget = entry.component;
          return (
            <WidgetFrame
              key={widget.id}
              testID={`widget-${widget.id}`}
              rect={cellRect(widget, size)}
              framed={entry.framed}
            >
              <Widget data={widget.data} />
            </WidgetFrame>
          );
        })}
    </View>
  );
}

const styles = StyleSheet.create({
  root: { flex: 1 },
});
