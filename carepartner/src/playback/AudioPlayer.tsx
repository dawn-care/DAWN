import React, { useEffect, useState } from "react";
import { Image, StyleSheet, Text, View } from "react-native";
import { useAudioPlayer, useAudioPlayerStatus } from "expo-audio";
import { fonts } from "../theme/fonts";
import { colors, fontSizes, radii, space } from "../theme/theme";
import type { PlayerProps } from "./types";
import { useReportOnce } from "./useReportOnce";

/**
 * expo-audio reports no playback errors, so these timeouts stand in for them:
 * a source that never loads, or one that loads but stops moving without
 * finishing (a truncated file can do this on Android), counts as failed.
 */
export const AUDIO_LOAD_TIMEOUT_MS = 15_000;
export const AUDIO_STALL_TIMEOUT_MS = 20_000;

type AudioPlayerProps = PlayerProps & {
  /** Shown under the image while the audio plays. */
  title?: string;
  /**
   * Shown while the audio plays, such as a song's album cover or the sender's
   * photo. Falls back to a music note if missing or if it fails to load.
   */
  coverUri?: string;
};

const FALLBACK_IMAGE = require("../../assets/images/music-note.jpg");

/**
 * Plays one audio file over a still screen, starting on its own.
 * Remount it (change its `key`) to play a different file.
 */
export default function AudioPlayer({ uri, title, coverUri, onStarted, onFinished, onFailed }: AudioPlayerProps) {
  const player = useAudioPlayer(uri);
  const status = useAudioPlayerStatus(player);
  const report = useReportOnce({ onStarted, onFinished, onFailed });
  const [imageFailed, setImageFailed] = useState(false);

  useEffect(() => {
    player.play();
    return () => {
      // On web, expo-audio leaves the sound playing after the player is released.
      try {
        player.pause();
      } catch {
        // On native the player is already released by now, and pausing it throws.
      }
    };
  }, [player]);

  useEffect(() => {
    if (status.playing) report.started();
  }, [status.playing, report]);

  useEffect(() => {
    if (status.didJustFinish) report.finished();
  }, [status.didJustFinish, report]);

  // Restarts whenever the playback position moves, so it only fires once playback is stuck.
  useEffect(() => {
    const [timeout, reason] = status.isLoaded
      ? [AUDIO_STALL_TIMEOUT_MS, "Audio stopped playing"]
      : [AUDIO_LOAD_TIMEOUT_MS, "Audio did not load"];
    const timer = setTimeout(() => report.failed(reason), timeout);
    return () => clearTimeout(timer);
  }, [status.isLoaded, status.currentTime, report]);

  return (
    <View style={styles.root}>
      <Image
        testID="audio-image"
        source={coverUri && !imageFailed ? { uri: coverUri } : FALLBACK_IMAGE}
        onError={() => setImageFailed(true)}
        style={styles.cover}
        resizeMode="cover"
      />
      {!!title && (
        <Text style={styles.title} numberOfLines={2}>
          {title}
        </Text>
      )}
    </View>
  );
}

const styles = StyleSheet.create({
  root: {
    flex: 1,
    alignItems: "center",
    justifyContent: "center",
    gap: space.md,
    padding: space.lg,
    backgroundColor: colors.player.background,
  },
  cover: {
    width: 280,
    height: 280,
    borderRadius: radii.tile,
  },
  title: {
    fontSize: fontSizes.tileWord,
    fontFamily: fonts.bold,
    color: colors.player.text,
    textAlign: "center",
  },
});
