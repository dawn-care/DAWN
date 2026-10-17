import React, { useState } from "react";
import { LayoutChangeEvent, StyleSheet, Text, View } from "react-native";
import { MaterialCommunityIcons } from "@expo/vector-icons";
import { fonts } from "../../theme/fonts";
import { colors, fontSizes, legibility, space } from "../../theme/theme";
import { fitFontSizeToBox, LINE_HEIGHT, linesAt } from "../fitText";
import type { HeadlineData } from "../types";

const ICON_SIZE = 60;

/** One news headline beside a newspaper icon, as large as fits. */
export default function Headline({ data }: { data: HeadlineData }) {
  const [box, setBox] = useState({ width: 0, height: 0 });
  const onLayout = (e: LayoutChangeEvent) => setBox(e.nativeEvent.layout);
  const size = fitFontSizeToBox(data.title, box.width, box.height, fontSizes.caption, legibility.minText);

  return (
    <View testID="headline" style={styles.root}>
      <MaterialCommunityIcons name="newspaper-variant-outline" size={ICON_SIZE} color={colors.day.text} />
      <View style={styles.body} onLayout={onLayout}>
        <Text
          style={[styles.title, { fontSize: size, lineHeight: size * LINE_HEIGHT }]}
          numberOfLines={linesAt(size, box.height)}
        >
          {data.title}
        </Text>
      </View>
    </View>
  );
}

const styles = StyleSheet.create({
  root: { flex: 1, flexDirection: "row", alignItems: "center", gap: space.sm },
  body: { flex: 1, alignSelf: "stretch", justifyContent: "center" },
  title: { fontFamily: fonts.bold, color: colors.day.text },
});
