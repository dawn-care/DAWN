// app/photos.tsx (or whatever route you want)
import React, { useEffect, useState } from "react";
import {
  Image,
  Pressable,
  SafeAreaView,
  StyleSheet,
  Text,
  View,
} from "react-native";
import { Stack } from "expo-router";

// ---- STUB TYPES (replace later) ----
type PhotoViewerConfig = {
  title: string; // e.g. "Your Grandkids"
  photoUrls?: string[]; // remote URLs from backend
  startIndex?: number; // optional
};

// ---- STUB BACKEND CALL (replace later) ----
async function fetchPhotoViewer(): Promise<PhotoViewerConfig> {
  // TODO: replace with real API call
  // const res = await fetch("https://your-api/photos");
  // return await res.json();

  return {
    title: "Your Grandkids",
    photoUrls: undefined,
    startIndex: 0,
    // For quick testing you can set:
    // photoUrls: [
    //   "https://images.unsplash.com/photo-1520975958225-2671f0a1cf2b?auto=format&fit=crop&w=1600&q=80",
    //   "https://images.unsplash.com/photo-1500530855697-b586d89ba3ee?auto=format&fit=crop&w=1600&q=80",
    // ],
  };
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

export default function PhotoViewerScreen() {
  const [config, setConfig] = useState<PhotoViewerConfig>({
    title: "Your Grandkids",
    photoUrls: undefined,
    startIndex: 0,
  });

  const [index, setIndex] = useState(0);

  useEffect(() => {
    let mounted = true;
    (async () => {
      try {
        const next = await fetchPhotoViewer();
        if (!mounted) return;
        setConfig(next);
        setIndex(next.startIndex ?? 0);
      } catch {
        // TODO: log/report
      }
    })();
    return () => {
      mounted = false;
    };
  }, []);

  const hasPhotos = !!config.photoUrls?.length;
  const safeIndex = hasPhotos
    ? ((index % config.photoUrls!.length) + config.photoUrls!.length) %
      config.photoUrls!.length
    : 0;

  const currentUrl = hasPhotos ? config.photoUrls![safeIndex] : undefined;

  const handlePrev = () => {
    if (!hasPhotos) return;
    setIndex((i) => i - 1);
  };

  const handleNext = () => {
    if (!hasPhotos) return;
    setIndex((i) => i + 1);
  };

  const arrowsDisabled = !hasPhotos || (config.photoUrls?.length ?? 0) <= 1;

  return (
    <>
      <Stack.Screen options={{ headerShown: false }} />

      <SafeAreaView style={styles.safe}>
        {/* Top title strip */}
        <View style={styles.topBar}>
          <Text style={styles.topTitle}>{config.title}</Text>
        </View>

        {/* Photo area */}
        <View style={styles.photoWrap}>
          {currentUrl ? (
            <Image source={{ uri: currentUrl }} style={styles.photo} />
          ) : (
            <View style={[styles.photo, styles.placeholder]}>
              <Text style={styles.placeholderText}>Photo will load from backend</Text>
              <Text style={styles.placeholderSubText}>
                (Set config.photoUrls once your API is wired)
              </Text>
            </View>
          )}

          {/* Left arrow (glass) */}
          <Pressable
            onPress={handlePrev}
            disabled={arrowsDisabled}
            style={({ pressed }) => [
              styles.leftWrapper,
              pressed && !arrowsDisabled && { transform: [{ scale: 0.98 }] },
              arrowsDisabled && { opacity: 0.45 },
            ]}
          >
            <GlassCard style={styles.arrowGlass}>
              <Text style={styles.arrowText}>‹</Text>
            </GlassCard>
          </Pressable>

          {/* Right arrow (glass) */}
          <Pressable
            onPress={handleNext}
            disabled={arrowsDisabled}
            style={({ pressed }) => [
              styles.rightWrapper,
              pressed && !arrowsDisabled && { transform: [{ scale: 0.98 }] },
              arrowsDisabled && { opacity: 0.45 },
            ]}
          >
            <GlassCard style={styles.arrowGlass}>
              <Text style={styles.arrowText}>›</Text>
            </GlassCard>
          </Pressable>
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
    paddingVertical: 14,
    alignItems: "center",
    justifyContent: "center",
  },
  topTitle: {
    fontSize: 54,
    fontWeight: "300",
    color: "rgba(0,0,0,0.85)",
  },

  // --- photo ---
  photoWrap: {
    flex: 1,
    position: "relative",
    backgroundColor: "#000",
  },
  photo: {
    flex: 1,
    width: "100%",
    height: "100%",
    resizeMode: "cover",
  },

  placeholder: {
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

  // --- arrows (glass) ---
  leftWrapper: {
    position: "absolute",
    left: 26,
    top: "50%",
    marginTop: -41, // centers the 82px button
  },
  rightWrapper: {
    position: "absolute",
    right: 26,
    top: "50%",
    marginTop: -41,
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
});