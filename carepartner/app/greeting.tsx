import React, { useEffect, useMemo, useRef, useState } from "react";
import {
  Image,
  Pressable,
  SafeAreaView,
  StyleSheet,
  Text,
  View,
  LayoutChangeEvent,
} from "react-native";
import { Stack, router } from "expo-router";
import { Video, ResizeMode, AVPlaybackStatus } from "expo-av";
import AgendaView, { type ConfigurationWithMetadata, VideoMetadata } from "./agendaView";

// ---- STUB TYPES (replace later) ----
type GreetingMediaConfig = {
  title: string; // "Good Morning, Grandma"
  videoUrl?: string; // remote URL from backend
};

interface GreetingScreenProps {
  configuration: ConfigurationWithMetadata;  
}

const API_URL = "http://localhost:8080"

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

export default function GreetingScreen({
  configuration: configuration
  }: GreetingScreenProps) {
  const videoRef = useRef<HTMLVideoElement>(null);
  const buildConfig = (configuration: ConfigurationWithMetadata): GreetingMediaConfig => {
      const metadata = configuration.metadata as VideoMetadata | null;
  
      if (!metadata) {
        return {
          title: "Video",
          videoUrl: "",
        };
      }
  
      return {
        title: metadata.displayed_name,
        videoUrl: `${API_URL}/video/file?id=${metadata.id}`,
      };
    };
  
    const [config, setConfig] = useState<GreetingMediaConfig>(
      buildConfig(configuration)
    );

  const [durationSec, setDurationSec] = useState(0);
  const [positionSec, setPositionSec] = useState(0);

  const [barWidth, setBarWidth] = useState(1);

  useEffect(() => {
    const next = buildConfig(configuration);
    setConfig(next);
  }, [configuration]);

  const progress = useMemo(() => {
    if (!durationSec) return 0;
    return Math.min(1, Math.max(0, positionSec / durationSec));
  }, [positionSec, durationSec]);

  const onPlaybackStatusUpdate = (
    event: React.SyntheticEvent<HTMLVideoElement>
  ) => {
    const video = event.currentTarget;

    const d = video.duration || 0;
    const p = video.currentTime || 0;

    setDurationSec(d);
    setPositionSec(p);
  };

  const handleNext = () => {
    // TODO: route to the actual next screen later
    // router.push("/someNextScreen");
    console.log("Next pressed");
  };

  const handleSeek = async (newProgress: number) => {
    if (!videoRef.current || !durationSec) return;
    const targetSeconds = newProgress * durationSec;
    try {
      videoRef.current.currentTime = targetSeconds;
    } catch {
      // ignore for now
    }
  };

  const onBarLayout = (e: LayoutChangeEvent) => {
    setBarWidth(e.nativeEvent.layout.width || 1);
  };

  return (
    <>
      <Stack.Screen options={{ headerShown: false }} />

      <SafeAreaView style={styles.safe}>
        {/* Top title strip */}
        <View style={styles.topBar}>
          {/* Swap this to your logo asset if desired */}
          <Image
            source={require("../assets/images/icon.png")}
            style={styles.topLogo}
          />
          <Text style={styles.topTitle}>{config.title}</Text>
        </View>

        {/* Video area */}
        <View style={styles.videoWrap}>
          {config.videoUrl ? (
            <video
              ref={videoRef}
              src={config.videoUrl}
              style={styles.video}
              autoPlay
              loop={false}
              controls={false}
              onTimeUpdate={onPlaybackStatusUpdate}
              onError={(e) => console.error("Video error:", e)}
            />
          ) : (
            <View style={[styles.video, styles.videoPlaceholder]}>
              <Text style={styles.placeholderText}>Video will load from backend</Text>
              <Text style={styles.placeholderSubText}>
                (Set config.videoUrl once your API is wired)
              </Text>
            </View>
          )}

          {/* Next arrow (glass) */}
          <Pressable
            onPress={handleNext}
            style={({ pressed }) => [
              styles.nextWrapper,
              pressed && { transform: [{ scale: 0.98 }] },
            ]}
          >
            <GlassCard style={styles.nextGlass}>
              <Text style={styles.nextArrow}>›</Text>
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
                <View style={[styles.progressFill, { width: `${progress * 100}%` }]} />
                <View style={[styles.progressDot, { left: `${progress * 100}%` }]} />
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
  safe: { flex: 1, backgroundColor: "#000" },

  // --- top strip ---
  topBar: {
    backgroundColor: "#c7f3f6",
    borderBottomWidth: 2,
    borderBottomColor: "rgba(0,0,0,0.6)",
    paddingHorizontal: 16,
    paddingVertical: 12,
    flexDirection: "row",
    alignItems: "center",
    gap: 14,
  },
  topLogo: {
    width: 42,
    height: 42,
    resizeMode: "contain",
  },
  topTitle: {
    fontSize: 44,
    fontWeight: "300",
    color: "rgba(0,0,0,0.85)",
  },

  // --- video ---
  videoWrap: {
    flex: 1,
    position: "relative",
  },
  video: {
    flex: 1,
  },
  videoPlaceholder: {
    backgroundColor: "#111",
    alignItems: "center",
    justifyContent: "center",
    gap: 8,
  },
  placeholderText: {
    color: "rgba(255,255,255,0.9)",
    fontSize: 18,
    fontWeight: "600",
  },
  placeholderSubText: {
    color: "rgba(255,255,255,0.7)",
    fontSize: 14,
  },

  // --- glass base ---
  glassBase: {
    borderRadius: 16,
    borderWidth: 1.5,
    borderColor: "rgba(0,0,0,0.25)",
    backgroundColor: "rgba(255,255,255,0.55)",
  },

  // --- next button (glass) ---
  nextWrapper: {
    position: "absolute",
    right: 26,
    top: "45%",
  },
  nextGlass: {
    width: 82,
    height: 82,
    borderRadius: 18,
    borderWidth: 2,
    borderColor: "rgba(0, 255, 240, 0.9)",
    alignItems: "center",
    justifyContent: "center",
  },
  nextArrow: {
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
    backgroundColor: "rgba(255,255,255,0.35)", // slightly “inner glass”
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