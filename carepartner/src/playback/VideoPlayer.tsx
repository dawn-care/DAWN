import React, { useEffect } from "react";
import { AppState, StyleSheet, View } from "react-native";
import { useEventListener } from "expo";
import { useVideoPlayer, VideoView } from "expo-video";
import { colors } from "../theme/theme";
import type { PlayerProps } from "./types";
import { useReportOnce } from "./useReportOnce";

/**
 * Plays one video full size with no controls, starting on its own.
 * Remount it (change its `key`) to play a different video.
 */
export default function VideoPlayer({ uri, onFinished, onFailed }: PlayerProps) {
  const player = useVideoPlayer(uri, (p) => {
    p.loop = false;
  });

  // Start from an effect, not the setup callback above: on web, play() only
  // reaches video elements that are already mounted, and VideoView mounts first.
  useEffect(() => {
    player.play();
  }, [player]);

  // Android pauses the video when the app goes to the background and doesn't
  // resume it, which would leave a frozen frame that never finishes.
  useEffect(() => {
    const subscription = AppState.addEventListener("change", (state) => {
      if (state === "active") player.play();
    });
    return () => subscription.remove();
  }, [player]);

  const report = useReportOnce({ onFinished, onFailed });
  useEventListener(player, "playToEnd", report.finished);
  useEventListener(player, "statusChange", ({ status, error }) => {
    if (status === "error") report.failed(error?.message ?? "Video could not be played");
  });

  return (
    <View style={styles.root}>
      <VideoView
        player={player}
        style={styles.video}
        contentFit="contain"
        nativeControls={false}
        fullscreenOptions={{ enable: false }}
        allowsPictureInPicture={false}
      />
    </View>
  );
}

const styles = StyleSheet.create({
  root: { flex: 1, backgroundColor: colors.player.background },
  video: { flex: 1 },
});
