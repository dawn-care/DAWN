import React, { useState } from "react";
import { Image, StyleSheet, Text, View } from "react-native";
import { MaterialCommunityIcons } from "@expo/vector-icons";
import { fonts } from "../../theme/fonts";
import { colors, fontSizes, radii, space } from "../../theme/theme";
import type { PhotoData } from "../types";

/** A large family photo with the person's name across the bottom. */
export default function Photo({ data }: { data: PhotoData }) {
  const [failed, setFailed] = useState(false);
  const showImage = !!data.uri && !failed;

  return (
    <View style={styles.root}>
      {showImage ? (
        <Image
          testID="photo-image"
          source={{ uri: data.uri! }}
          style={StyleSheet.absoluteFill}
          resizeMode="cover"
          onError={() => setFailed(true)}
        />
      ) : (
        <View testID="photo-placeholder" style={styles.placeholder}>
          <MaterialCommunityIcons name="image-outline" size={120} color={colors.day.placeholderIcon} />
        </View>
      )}
      {!!data.name && (
        <View style={styles.caption}>
          <Text style={styles.name} numberOfLines={1}>
            {data.name}
          </Text>
        </View>
      )}
    </View>
  );
}

const styles = StyleSheet.create({
  root: {
    flex: 1,
    borderRadius: radii.tile,
    overflow: "hidden",
    backgroundColor: colors.day.placeholder,
  },
  placeholder: { flex: 1, alignItems: "center", justifyContent: "center" },
  caption: {
    position: "absolute",
    left: 0,
    right: 0,
    bottom: 0,
    paddingVertical: space.xs,
    paddingHorizontal: space.sm,
    backgroundColor: colors.day.photoCaption,
  },
  name: {
    fontFamily: fonts.bold,
    fontSize: fontSizes.name,
    color: colors.player.text,
    textAlign: "center",
  },
});
