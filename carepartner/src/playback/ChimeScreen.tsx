/**
 * ChimeScreen.tsx
 *
 * What the tablet shows before anything plays, so a resident isn't startled:
 * a calm screen with one large icon for the kind of media, a gentle chime,
 * then `CHIME_PAUSE_MS` of quiet before the media itself (#303).
 *
 * - The chime is quieter at night. When staff have switched it off for the
 *   resident (`sound` false), the pause still comes, without the sound.
 * - The chime never holds up a play: if it hasn't loaded within
 *   `CHIME_LOAD_TIMEOUT_MS`, or hasn't finished within `CHIME_MAX_MS`, the
 *   pause starts anyway.
 * - `onDone` is called once, when the pause is over.
 *
 * The sound is `assets/sounds/chime.mp3`, from Pixabay (free to use under the
 * Pixabay Content License, no attribution needed).
 */

import React, { useEffect, useRef, useState } from "react";
import { StyleSheet, View } from "react-native";
import { MaterialCommunityIcons } from "@expo/vector-icons";
import { useAudioPlayer, useAudioPlayerStatus } from "expo-audio";
import { colors } from "../theme/theme";
import type { MediaKind } from "./media";

/** The chime's volume by day and at night, from 0 to 1 of the tablet's volume. Tune by ear. */
export const CHIME_VOLUME = 1.0;
export const NIGHT_CHIME_VOLUME = 0.3;
/** Quiet after the chime, before the media starts. */
export const CHIME_PAUSE_MS = 1_000;
/** A chime that hasn't loaded by now is skipped. */
export const CHIME_LOAD_TIMEOUT_MS = 1_500;
/** A chime still going by now (it is about 1.4 s) counts as over. */
export const CHIME_MAX_MS = 3_000;

const CHIME = require("../../assets/sounds/chime.mp3");

const ICONS: Record<MediaKind, React.ComponentProps<typeof MaterialCommunityIcons>["name"]> = {
  video: "play-circle-outline",
  audio: "music-note",
  photo: "image-outline",
};

type Props = {
  kind: MediaKind;
  /** Whether to play the chime; the pause comes either way. */
  sound: boolean;
  night: boolean;
  onDone: () => void;
};

export default function ChimeScreen({ kind, sound, night, onDone }: Props) {
  const player = useAudioPlayer(sound ? CHIME : null);
  const status = useAudioPlayerStatus(player);
  const [soundOver, setSoundOver] = useState(!sound);
  const done = useRef(onDone);
  // Set once when the chime starts; night beginning mid-chime doesn't matter.
  const volume = useRef(night ? NIGHT_CHIME_VOLUME : CHIME_VOLUME);
  useEffect(() => {
    done.current = onDone;
  });

  useEffect(() => {
    if (!sound) return;
    player.volume = volume.current;
    player.play();
    return () => {
      // On web, expo-audio leaves the sound playing after the player is released.
      try {
        player.pause();
      } catch {
        // On native the player is already released by now, and pausing it throws.
      }
    };
  }, [player, sound]);

  useEffect(() => {
    if (status.didJustFinish) setSoundOver(true);
  }, [status.didJustFinish]);

  // Never let the chime hold up the play.
  useEffect(() => {
    if (soundOver) return;
    const timer = setTimeout(() => setSoundOver(true), status.isLoaded ? CHIME_MAX_MS : CHIME_LOAD_TIMEOUT_MS);
    return () => clearTimeout(timer);
  }, [soundOver, status.isLoaded]);

  useEffect(() => {
    if (!soundOver) return;
    const timer = setTimeout(() => done.current(), CHIME_PAUSE_MS);
    return () => clearTimeout(timer);
  }, [soundOver]);

  return (
    <View testID="chime-screen" style={styles.root}>
      <MaterialCommunityIcons testID={`chime-icon-${kind}`} name={ICONS[kind]} size={160} color={colors.player.text} />
    </View>
  );
}

const styles = StyleSheet.create({
  root: { flex: 1, alignItems: "center", justifyContent: "center", backgroundColor: colors.player.background },
});
