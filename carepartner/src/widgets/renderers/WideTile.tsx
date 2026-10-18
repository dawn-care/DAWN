import React from "react";
import { StyleSheet, Text, View } from "react-native";
import { MaterialCommunityIcons } from "@expo/vector-icons";
import { fonts } from "../../theme/fonts";
import { colors, fontSizes, space } from "../../theme/theme";
import type { WideTileData } from "../types";
import { tileIcon } from "./icons";

const ICON_SIZE = 72;

/** A 2×1 data widget: the icon, then the label, the word in large type and an optional detail line. */
export default function WideTile({ data }: { data: WideTileData }) {
  return (
    <View testID="wide-tile" style={styles.root}>
      <MaterialCommunityIcons name={tileIcon(data.icon)} size={ICON_SIZE} color={colors.day.text} />
      <View style={styles.body}>
        <Text style={styles.label} numberOfLines={1}>
          {data.label}
        </Text>
        <Text style={styles.word} numberOfLines={1}>
          {data.word}
        </Text>
        {!!data.detail && (
          <Text style={styles.detail} numberOfLines={1}>
            {data.detail}
          </Text>
        )}
      </View>
    </View>
  );
}

const styles = StyleSheet.create({
  root: { flex: 1, flexDirection: "row", alignItems: "center", gap: space.sm },
  body: { flex: 1, justifyContent: "center" },
  label: { fontFamily: fonts.regular, fontSize: fontSizes.label, color: colors.day.textSecondary },
  word: { fontFamily: fonts.bold, fontSize: fontSizes.tileWord, color: colors.day.text },
  detail: { fontFamily: fonts.regular, fontSize: fontSizes.label, color: colors.day.text },
});
