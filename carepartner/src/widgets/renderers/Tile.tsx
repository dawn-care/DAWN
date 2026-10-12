import React from "react";
import { StyleSheet, Text, View } from "react-native";
import { MaterialCommunityIcons } from "@expo/vector-icons";
import { fonts } from "../../theme/fonts";
import { colors, fontSizes, space } from "../../theme/theme";
import type { TileData } from "../types";
import { tileIcon } from "./icons";

/** Smaller when a label shares the tile, so icon, label and word fit one cell. */
export const TILE_ICON_SIZE = { plain: 88, withLabel: 60 } as const;

/** One icon, an optional short label and one word, such as weather or a game result. */
export default function Tile({ data }: { data: TileData }) {
  return (
    <View style={styles.root}>
      <MaterialCommunityIcons
        testID="tile-icon"
        name={tileIcon(data.icon)}
        size={data.label ? TILE_ICON_SIZE.withLabel : TILE_ICON_SIZE.plain}
        color={colors.day.text}
      />
      {!!data.label && (
        <Text style={styles.label} numberOfLines={1}>
          {data.label}
        </Text>
      )}
      <Text style={styles.word} numberOfLines={1}>
        {data.word}
      </Text>
    </View>
  );
}

const styles = StyleSheet.create({
  root: { flex: 1, alignItems: "center", justifyContent: "center", gap: space.xs / 2 },
  label: {
    fontFamily: fonts.regular,
    fontSize: fontSizes.label,
    lineHeight: fontSizes.label * 1.2,
    color: colors.day.textSecondary,
  },
  word: {
    fontFamily: fonts.bold,
    fontSize: fontSizes.tileWord,
    lineHeight: fontSizes.tileWord * 1.2,
    color: colors.day.text,
  },
});
