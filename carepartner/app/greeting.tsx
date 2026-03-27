import React, { useEffect, useMemo, useRef, useState } from "react";
import {
  Image,
  Pressable,
  StyleSheet,
  Text,
  View,
  LayoutChangeEvent,
} from "react-native";
import { Stack } from "expo-router";
import { Video } from "expo-av";
import { type ConfigurationWithMetadata, type MediaPlaybackProps, VideoMetadata } from "./agendaView";
import { API_URL } from "./config";

type GreetingMediaConfig = {
  videoUrl?: string;
};

interface GreetingScreenProps {
  configuration: ConfigurationWithMetadata;
  onPlaybackComplete: MediaPlaybackProps["onPlaybackComplete"];
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

export default function GreetingScreen({
  configuration: configuration,
  onPlaybackComplete,
  }: GreetingScreenProps) {
  const videoRef = useRef<HTMLVideoElement>(null);
  const buildConfig = (configuration: ConfigurationWithMetadata): GreetingMediaConfig => {
      const metadata = configuration.metadata as VideoMetadata | null;
  
      if (!metadata) {
        return {
          videoUrl: "",
        };
      }
  
      return {
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

  // Programmatically start playback when the video element mounts or src changes
  useEffect(() => {
    const video = videoRef.current;
    if (!video || !config.videoUrl) return;

    const tryPlay = () => {
      video.play().catch((e) => {
        console.error('Autoplay failed:', e);
      });
    };

    // If already loaded enough, play immediately; otherwise wait for canplay
    if (video.readyState >= 3) {
      tryPlay();
    } else {
      video.addEventListener('canplay', tryPlay, { once: true });
      return () => video.removeEventListener('canplay', tryPlay);
    }
  }, [config.videoUrl]);

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

      <View style={styles.safe}>
        <View style={styles.videoWrap}>
          {config.videoUrl ? (
            <video
              ref={videoRef}
              src={config.videoUrl}
              style={{ width: '100%', height: '100%', objectFit: 'cover' }}
              autoPlay
              playsInline
              loop={false}
              controls={false}
              onTimeUpdate={onPlaybackStatusUpdate}
              onEnded={onPlaybackComplete}
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
      </View>
    </>
  );
}

const styles = StyleSheet.create({
  safe: { flex: 1, backgroundColor: "#000" },

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