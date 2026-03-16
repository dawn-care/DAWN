// app/audio.tsx (or wherever your route lives)
import React, { useEffect, useMemo, useRef, useState } from "react";
import {
  Image,
  LayoutChangeEvent,
  Pressable,
  SafeAreaView,
  StyleSheet,
  Text,
  View,
} from "react-native";
import { Stack } from "expo-router";
import { Audio, AVPlaybackStatus } from "expo-av";
import AgendaView, { type ConfigurationWithMetadata, AudioMetadata } from "./agendaView";

// ---- STUB TYPES (replace later) ----
type AudioMediaConfig = {
  title: string; // e.g. "Fly Me To the Moon"
  audioUrl?: string; // remote URL from backend
};

interface AudioScreenProps {
  configuration: ConfigurationWithMetadata;  
}


function formatTime(totalSeconds: number) {
  const sec = Math.max(0, Math.floor(totalSeconds));
  const m = Math.floor(sec / 60);
  const s = sec % 60;
  return `${m}:${String(s).padStart(2, "0")}`;
}

// Simple reusable glass container (local to this file for now)
function GlassCard({
  children,
  style,
}: {
  children: React.ReactNode;
  style?: any;
}) {
  return <View style={[styles.glassBase, style]}>{children}</View>;
}

export default function AudioScreen({
    configuration: configuration
  }: AudioScreenProps){
  const soundRef = useRef<Audio.Sound | null>(null);
  const buildConfig = (configuration: ConfigurationWithMetadata): AudioMediaConfig => {
    const metadata = configuration.metadata as AudioMetadata | null;

    if (!metadata) {
      return {
        title: "Audio",
        audioUrl: ""
      };
    }

    return {
      title: metadata.displayed_name,
      audioUrl: metadata.uri
    };
  };

  const [config, setConfig] = useState<AudioMediaConfig>(
    buildConfig(configuration)
  );

  const [durationSec, setDurationSec] = useState(0);
  const [positionSec, setPositionSec] = useState(0);

  const [barWidth, setBarWidth] = useState(1);

  // Fetch config
  useEffect(() => {
     const next = buildConfig(configuration);
     setConfig(next);
   }, [configuration]);

  // Load/unload audio when audioUrl changes
  useEffect(() => {
    let cancelled = false;

    (async () => {
      // Clean up any previous sound
      if (soundRef.current) {
        try {
          soundRef.current.setOnPlaybackStatusUpdate(null);
          await soundRef.current.unloadAsync();
        } catch {
          // ignore
        }
        soundRef.current = null;
      }

      if (!config.audioUrl) {
        setDurationSec(0);
        setPositionSec(0);
        return;
      }

      try {
        await Audio.setAudioModeAsync({
          playsInSilentModeIOS: true,
          staysActiveInBackground: true,
          shouldDuckAndroid: true,
        });

        const { sound } = await Audio.Sound.createAsync(
          { uri: config.audioUrl },
          { shouldPlay: true, isLooping: false },
          (status: AVPlaybackStatus) => {
            if (!status.isLoaded) return;
            const d = status.durationMillis ? status.durationMillis / 1000 : 0;
            const p = status.positionMillis ? status.positionMillis / 1000 : 0;
            if (!cancelled) {
              setDurationSec(d);
              setPositionSec(p);
            }
          }
        );

        if (cancelled) {
          try {
            await sound.unloadAsync();
          } catch {}
          return;
        }

        soundRef.current = sound;
      } catch {
        // TODO: log/report
      }
    })();

    return () => {
      cancelled = true;
    };
  }, [config.audioUrl]);

  // Final cleanup on unmount
  useEffect(() => {
    return () => {
      (async () => {
        if (soundRef.current) {
          try {
            soundRef.current.setOnPlaybackStatusUpdate(null);
            await soundRef.current.unloadAsync();
          } catch {}
          soundRef.current = null;
        }
      })();
    };
  }, []);

  const progress = useMemo(() => {
    if (!durationSec) return 0;
    return Math.min(1, Math.max(0, positionSec / durationSec));
  }, [positionSec, durationSec]);

  const handleSeek = async (newProgress: number) => {
    if (!soundRef.current || !durationSec) return;
    const targetMillis = newProgress * durationSec * 1000;
    try {
      await soundRef.current.setPositionAsync(targetMillis);
    } catch {
      // ignore
    }
  };

  const onBarLayout = (e: LayoutChangeEvent) => {
    setBarWidth(e.nativeEvent.layout.width || 1);
  };

  const handlePrev = () => {
    // TODO: hook to your playlist navigation
    console.log("Prev pressed");
  };

  const handleNext = () => {
    // TODO: hook to your playlist navigation or “continue”
    console.log("Next pressed");
  };

  return (
    <>
      <Stack.Screen options={{ headerShown: false }} />

      <SafeAreaView style={styles.safe}>
        {/* Top title strip */}
        <View style={styles.topBar}>
          <Text style={styles.topTitle}>{config.title}</Text>
        </View>

        {/* Main area */}
        <View style={styles.body}>
          {/* Center “album” / placeholder */}
          <View style={styles.centerWrap}>
            <View style={styles.coverFrame}>
              <Image
                // Put your default music note image here:
                // e.g. /assets/images/music-note.png
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

          {/* Left arrow */}
          <Pressable
            onPress={handlePrev}
            style={({ pressed }) => [
              styles.leftWrapper,
              pressed && { transform: [{ scale: 0.98 }] },
            ]}
          >
            <GlassCard style={styles.arrowGlass}>
              <Text style={styles.arrowText}>‹</Text>
            </GlassCard>
          </Pressable>

          {/* Right arrow */}
          <Pressable
            onPress={handleNext}
            style={({ pressed }) => [
              styles.rightWrapper,
              pressed && { transform: [{ scale: 0.98 }] },
            ]}
          >
            <GlassCard style={styles.arrowGlass}>
              <Text style={styles.arrowText}>›</Text>
            </GlassCard>
          </Pressable>

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
      </SafeAreaView>
    </>
  );
}

const styles = StyleSheet.create({
  safe: { flex: 1, backgroundColor: "#d9d9d9" },

  // --- top strip ---
  topBar: {
    backgroundColor: "#d7fbff",
    borderBottomWidth: 2,
    borderBottomColor: "rgba(0,0,0,0.6)",
    paddingHorizontal: 16,
    paddingVertical: 14,
    alignItems: "center",
    justifyContent: "center",
  },
  topTitle: {
    fontSize: 54,
    fontWeight: "300",
    color: "rgba(0,0,0,0.85)",
  },

  body: { flex: 1, position: "relative" },

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