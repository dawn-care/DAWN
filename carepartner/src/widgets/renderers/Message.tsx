import React, { useState } from "react";
import { Image, LayoutChangeEvent, StyleSheet, Text, View } from "react-native";
import { fonts } from "../../theme/fonts";
import { colors, fontSizes, legibility, radii, space } from "../../theme/theme";
import { fitFontSizeToBox, LINE_HEIGHT, linesAt } from "../fitText";
import type { MessageData } from "../types";

/** A caregiver's short message, as large as fits, with an optional photo and role. */
export default function Message({ data }: { data: MessageData }) {
  const [box, setBox] = useState({ width: 0, height: 0 });
  const [photoFailed, setPhotoFailed] = useState(false);
  const onLayout = (e: LayoutChangeEvent) => setBox(e.nativeEvent.layout);

  const textHeight = box.height - (data.label ? fontSizes.label * LINE_HEIGHT + space.xs / 2 : 0);
  const size = fitFontSizeToBox(data.text, box.width, textHeight, fontSizes.caption, legibility.minText);

  return (
    <View testID="message" style={styles.root}>
      {!!data.photoUri && !photoFailed && (
        <Image
          testID="message-photo"
          source={{ uri: data.photoUri }}
          style={styles.photo}
          resizeMode="cover"
          onError={() => setPhotoFailed(true)}
        />
      )}
      <View testID="message-body" style={styles.body} onLayout={onLayout}>
        {!!data.label && (
          <Text style={styles.label} numberOfLines={1}>
            {data.label}
          </Text>
        )}
        {/* Hidden until measured, so it doesn't show at full size and then shrink. */}
        <Text
          style={[styles.text, { fontSize: size, lineHeight: size * LINE_HEIGHT, opacity: box.width ? 1 : 0 }]}
          numberOfLines={linesAt(size, textHeight)}
        >
          {data.text}
        </Text>
      </View>
    </View>
  );
}

const styles = StyleSheet.create({
  root: { flex: 1, flexDirection: "row", alignItems: "center", gap: space.sm },
  photo: { height: "100%", aspectRatio: 1, borderRadius: radii.tile, backgroundColor: colors.day.placeholder },
  body: { flex: 1, alignSelf: "stretch", justifyContent: "center", gap: space.xs / 2 },
  label: {
    fontFamily: fonts.regular,
    fontSize: fontSizes.label,
    lineHeight: fontSizes.label * LINE_HEIGHT,
    color: colors.day.textSecondary,
  },
  text: { fontFamily: fonts.bold, color: colors.day.text },
});
