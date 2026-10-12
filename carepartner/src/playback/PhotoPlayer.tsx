import React, { useCallback, useEffect, useRef, useState } from "react";
import { Image, StyleSheet, View } from "react-native";
import { colors } from "../theme/theme";
import type { PlayerProps } from "./types";

/**
 * A photo that has neither loaded nor failed by now counts as failed. On
 * Android a stuck request may never report either.
 */
export const PHOTO_LOAD_TIMEOUT_MS = 15_000;

type PhotoPlayerProps = PlayerProps & {
  /** Seconds to show the photo once it has loaded; the item's configured duration. */
  durationSeconds: number;
};

/**
 * Shows one photo full screen, whole and uncropped, then reports finished.
 * The time starts once the photo has loaded, so a slow download doesn't use
 * it up. Remount it (change its `key`) to show a different photo.
 */
export default function PhotoPlayer({ uri, durationSeconds, onFinished, onFailed }: PhotoPlayerProps) {
  const [loaded, setLoaded] = useState(false);
  const ended = useRef(false);

  // Read the callbacks through a ref, so a parent passing new functions on
  // every render doesn't restart the timers below.
  const callbacks = useRef({ onFinished, onFailed });
  useEffect(() => {
    callbacks.current = { onFinished, onFailed };
  });

  /** Reports finished, or failed when given a reason; only the first report counts. */
  const report = useCallback((failure?: string) => {
    if (ended.current) return;
    ended.current = true;
    if (failure === undefined) callbacks.current.onFinished();
    else callbacks.current.onFailed(failure);
  }, []);

  useEffect(() => {
    const timer = loaded
      ? setTimeout(() => report(), durationSeconds * 1000)
      : setTimeout(() => report("Photo did not load"), PHOTO_LOAD_TIMEOUT_MS);
    return () => clearTimeout(timer);
  }, [loaded, durationSeconds, report]);

  return (
    <View style={styles.root}>
      <Image
        testID="photo-player-image"
        source={{ uri }}
        style={styles.photo}
        resizeMode="contain"
        onLoad={() => setLoaded(true)}
        onError={() => report("Photo could not be loaded")}
      />
    </View>
  );
}

const styles = StyleSheet.create({
  root: { flex: 1, backgroundColor: colors.player.background },
  photo: { flex: 1 },
});
