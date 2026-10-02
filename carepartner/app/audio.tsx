/**
 * audio.tsx
 * 
 * User: Carepartner
 * 
 * Plays audio content within the agenda flow.
 * 
 * Responsibilities:
 * - Stream audio from backend endpoints
 * - Display playback progress and duration
 * - Allow user seeking via progress bar
 * - Notify AgendaView when playback completes
 * 
 * Navigation:
 * - Controlled by AgendaView
 * 
 * Notes:
 * - Uses hidden HTML <audio> element for playback
 * - Progress bar supports click-to-seek interaction
 * - Audio auto-plays when loaded
 */

import React, { useEffect, useMemo, useRef, useState } from "react";
import {
  Image,
  LayoutChangeEvent,
  Pressable,
  StyleSheet,
  Text,
  View,
} from "react-native";
import { Stack } from "expo-router";
import { type ConfigurationWithMetadata, type MediaPlaybackProps, AudioMetadata } from "./agendaView";
import { API_URL } from "../src/api/config";

/* ============================
   Types
============================ */

/**
 * Normalized configuration used internally by this component.
 */
type AudioMediaConfig = {
  audioUrl?: string;
};

/**
 * Props passed from AgendaView.
 */
interface AudioScreenProps {
  configuration: ConfigurationWithMetadata;
  onPlaybackComplete: MediaPlaybackProps["onPlaybackComplete"];
}

/* ============================
   Helpers
============================ */

/**
 * Formats seconds into MM:SS.
 */
function formatTime(totalSeconds: number) {
  const sec = Math.max(0, Math.floor(totalSeconds));
  const m = Math.floor(sec / 60);
  const s = sec % 60;
  return `${m}:${String(s).padStart(2, "0")}`;
}

/**
 * Reusable glass-style container.
 */
function GlassCard({
  children,
  style,
}: {
  children: React.ReactNode;
  style?: any;
}) {
  return <View style={[styles.glassBase, style]}>{children}</View>;
}

/* ============================
   Component
============================ */

/**
 * AudioScreen Component
 * 
 * Handles:
 * - Building audio URL from metadata
 * - Managing playback state (time, duration)
 * - Rendering progress UI
 * - Handling playback completion
 */
export default function AudioScreen({
  configuration: configuration,
  onPlaybackComplete,
  }: AudioScreenProps){

  /**
   * Converts configuration metadata into audio URL.
   */
  const buildConfig = (configuration: ConfigurationWithMetadata): AudioMediaConfig => {
    const metadata = configuration.metadata as AudioMetadata | null;

    if (!metadata) {
      return {
        audioUrl: ""
      };
    }

    return {
      audioUrl: `${API_URL}/audio/file?id=${metadata.id}`
    };
  };

  /* ============================
       STATE
    ============================ */
  const [config, setConfig] = useState<AudioMediaConfig>(
    buildConfig(configuration)
  );

  const audioRef = useRef<HTMLAudioElement>(null);

  const [barWidth, setBarWidth] = useState(1);
  const [durationSec, setDurationSec] = useState(0);
  const [positionSec, setPositionSec] = useState(0);

  /**
   * Playback progress (0 → 1).
   */
  const progress = useMemo(() => {
    if (!durationSec) return 0;
    return Math.min(1, Math.max(0, positionSec / durationSec));
  }, [positionSec, durationSec]);

  /* ============================
       EFFECTS
    ============================ */

  /**
   * Rebuild config when agenda item changes.
   */
  useEffect(() => {
     const next = buildConfig(configuration);
     setConfig(next);
   }, [configuration]);

  /* ============================
       HANDLERS
    ============================ */

  /**
   * Updates playback position and duration.
   */
  const onPlaybackStatusUpdate = (
    event: React.SyntheticEvent<HTMLAudioElement>
  ) => {
    const audio = event.currentTarget;
    const d = audio.duration || 0;
    const p = audio.currentTime || 0;
    setDurationSec(d);
    setPositionSec(p);
  };

  /**
   * Handles user seeking via progress bar.
   */
  const handleSeek = async (newProgress: number) => {
    if (!audioRef.current || !durationSec) return;
    const targetSeconds = newProgress * durationSec;
    try {
      audioRef.current.currentTime = targetSeconds;
    } catch {
      // ignore
    }
  };

  /**
   * Tracks progress bar width for accurate seeking.
   */
  const onBarLayout = (e: LayoutChangeEvent) => {
    setBarWidth(e.nativeEvent.layout.width || 1);
  };

  /* ============================
       RENDER
    ============================ */
  return (
    <>
      <Stack.Screen options={{ headerShown: false }} />

      <View style={styles.safe}>
        <View style={styles.body}>
          <View style={styles.centerWrap}>
            <View style={styles.coverFrame}>
              <Image
                source={require("../assets/images/music-note.jpg")}
                style={styles.coverImage}
              />
              {!config.audioUrl && (
                <View style={styles.coverOverlay}>
                  <Text style={styles.placeholderText}>
                    Audio will load from backend
                  </Text>
                  <Text style={styles.placeholderSubText}>
                    (Set config.audioUrl once your API is wired)
                  </Text>
                </View>
              )}
            </View>
          </View>

          {/* Hidden audio element for playback */}
          {config.audioUrl ? (
            <audio
              ref={audioRef}
              src={config.audioUrl}
              autoPlay
              loop={false}
              onTimeUpdate={onPlaybackStatusUpdate}
              onEnded={onPlaybackComplete}
              onError={(e) => {
                const audio = e.currentTarget as HTMLAudioElement;
                const err = audio?.error;
                console.error("Audio error:", err?.code, err?.message, "src:", config.audioUrl);
              }}
              style={{ display: "none" }}
            />
          ) : null}

          {/* Progress controls (glass row) */}
          <GlassCard style={styles.progressGlass}>
            <View style={styles.progressRow}>
              <GlassCard style={styles.timePill}>
                <Text style={styles.timeText}>{formatTime(positionSec)}</Text>
              </GlassCard>

              <Pressable
                onLayout={onBarLayout}
                style={styles.progressBar}
                onPress={(e) => {
                  const x = e.nativeEvent.locationX;
                  const newProgress = Math.min(1, Math.max(0, x / barWidth));
                  handleSeek(newProgress);
                }}
              >
                <View style={styles.progressTrack} />
                <View
                  style={[styles.progressFill, { width: `${progress * 100}%` }]}
                />
                <View
                  style={[styles.progressDot, { left: `${progress * 100}%` }]}
                />
              </Pressable>

              <GlassCard style={styles.timePill}>
                <Text style={styles.timeText}>{formatTime(durationSec)}</Text>
              </GlassCard>
            </View>
          </GlassCard>
        </View>
      </View>
    </>
  );
}

const styles = StyleSheet.create({
  safe: { flex: 1, backgroundColor: "#d9d9d9", overflow: "hidden" },

  body: { flex: 1, position: "relative", overflow: "hidden" },

  // --- center cover ---
  centerWrap: {
    flex: 1,
    alignItems: "center",
    justifyContent: "center",
    paddingBottom: 90, // keep space for the progress bar
  },
  coverFrame: {
    width: 360,
    height: 360,
    backgroundColor: "#f2f2f2",
    alignItems: "center",
    justifyContent: "center",
    borderWidth: 2,
    borderColor: "rgba(0,0,0,0.25)",
  },
  coverImage: {
    width: 190,
    height: 190,
    resizeMode: "contain",
    opacity: 0.85,
  },
  coverOverlay: {
    position: "absolute",
    bottom: 18,
    left: 18,
    right: 18,
    alignItems: "center",
    gap: 6,
  },
  placeholderText: {
    color: "rgba(0,0,0,0.75)",
    fontSize: 16,
    fontWeight: "700",
  },
  placeholderSubText: {
    color: "rgba(0,0,0,0.6)",
    fontSize: 13,
  },

  // --- glass base ---
  glassBase: {
    borderRadius: 16,
    borderWidth: 1.5,
    borderColor: "rgba(0,0,0,0.25)",
    backgroundColor: "rgba(255,255,255,0.55)",
  },

  // --- arrows ---
  leftWrapper: {
    position: "absolute",
    left: 26,
    top: "45%",
  },
  rightWrapper: {
    position: "absolute",
    right: 26,
    top: "45%",
  },
  arrowGlass: {
    width: 82,
    height: 82,
    borderRadius: 18,
    borderWidth: 2,
    borderColor: "rgba(0, 255, 240, 0.9)",
    alignItems: "center",
    justifyContent: "center",
  },
  arrowText: {
    fontSize: 54,
    color: "rgba(0,0,0,0.75)",
    lineHeight: 54,
    marginTop: -2,
  },

  // --- progress row (glass container) ---
  progressGlass: {
    position: "absolute",
    left: 16,
    right: 16,
    bottom: 16,
    paddingVertical: 12,
    paddingHorizontal: 14,
    borderRadius: 18,
  },
  progressRow: {
    flexDirection: "row",
    alignItems: "center",
    gap: 14,
  },
  timePill: {
    minWidth: 72,
    paddingVertical: 8,
    paddingHorizontal: 12,
    borderRadius: 14,
    backgroundColor: "rgba(255,255,255,0.35)",
    borderWidth: 1.5,
    borderColor: "rgba(0,0,0,0.2)",
    alignItems: "center",
  },
  timeText: {
    fontSize: 22,
    color: "rgba(0,0,0,0.75)",
    fontWeight: "600",
  },

  progressBar: {
    flex: 1,
    height: 26,
    justifyContent: "center",
    position: "relative",
  },
  progressTrack: {
    position: "absolute",
    left: 0,
    right: 0,
    height: 6,
    borderRadius: 999,
    backgroundColor: "rgba(255,255,255,0.35)",
    borderWidth: 1,
    borderColor: "rgba(0,0,0,0.2)",
  },
  progressFill: {
    position: "absolute",
    left: 0,
    height: 6,
    borderRadius: 999,
    backgroundColor: "rgba(0, 255, 240, 0.85)",
  },
  progressDot: {
    position: "absolute",
    width: 26,
    height: 26,
    borderRadius: 999,
    backgroundColor: "rgba(0, 255, 240, 0.95)",
    borderWidth: 2,
    borderColor: "rgba(0,0,0,0.2)",
    marginLeft: -13,
  },
});