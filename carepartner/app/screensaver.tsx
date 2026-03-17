import React, { useEffect, useMemo, useState } from "react";
import { Stack } from "expo-router";
import {
  ImageBackground,
  Pressable,
  SafeAreaView,
  StyleSheet,
  Text,
  View,
} from "react-native";
import { router } from "expo-router";
import { Image } from "react-native";

// ---- STUB TYPES (replace later with your real backend types) ----
type ScreensaverConfig = {
  title: string; // "StartUp"
  autoStartTimeLabel: string; // e.g. "8:00am"
  backgroundImageUrl?: string; // remote URL from backend
};

// ---- STUB BACKEND CALL (replace later) ----
async function fetchScreensaverConfig(): Promise<ScreensaverConfig> {
  // TODO: replace with real API call
  // const res = await fetch("https://your-api/screensaver-config");
  // return await res.json();

  return {
    title: "Start-Up",
    autoStartTimeLabel: "8:00am",
    backgroundImageUrl: undefined, // stub: set to a real URL later
  };
}

export default function Screensaver() {
  const [config, setConfig] = useState<ScreensaverConfig>({
    title: "StartUp",
    autoStartTimeLabel: "8:00am",
    backgroundImageUrl: undefined,
  });

  useEffect(() => {
    let mounted = true;

    (async () => {
      try {
        const next = await fetchScreensaverConfig();
        if (mounted) setConfig(next);
      } catch {
        // TODO: optionally log/report
      }
    })();

    return () => {
      mounted = false;
    };
  }, []);


  const backgroundSource = useMemo(() => {
    if (config.backgroundImageUrl) return { uri: config.backgroundImageUrl };

    return require("../assets/images/defaultBackground.png");
  }, [config.backgroundImageUrl]);

  const handleEnter = () => {
    router.push({ pathname: "/agendaView" } as any);
  };

  return (
    <>
    <Stack.Screen options={{ headerShown: false }} />
    <View style={styles.root}>
        <ImageBackground source={backgroundSource} style={styles.bg} resizeMode="cover">
        <SafeAreaView style={styles.safe}>
            <View style={styles.content}>
            {/* TOP */}
            <View style={styles.topArea}>
                <View style={styles.titleGlass}>
                <Image source={require("../assets/images/icon.png")} style={styles.logo} />
                <Text style={styles.titleText}>{config.title}</Text>
                </View>

                <View style={styles.autoGlass}>
                <Text style={styles.autoText}>
                    Auto-Starts at{" "}
                    <Text style={styles.autoTimeText}>{config.autoStartTimeLabel}</Text>
                </Text>
                </View>
            </View>

            {/* BOTTOM */}
            <View style={styles.bottomArea}>
                <Pressable
                onPress={handleEnter}
                style={({ pressed }) => [
                    styles.arrowButton,
                    pressed && { transform: [{ scale: 0.98 }] },
                ]}
                >
                <Text style={styles.arrowText}>›</Text>
                </Pressable>

                <View style={styles.enterPill}>
                <Text style={styles.enterText}>Enter Start-Up</Text>
                </View>
            </View>
            </View>
        </SafeAreaView>
        </ImageBackground>
    </View>
    </>
    );
}

const styles = StyleSheet.create({
  root: { flex: 1, backgroundColor: "#000" },
  bg: { flex: 1 },
  safe: { flex: 1 },
  content: {
    flex: 1,
    alignItems: "center",
    justifyContent: "space-between",
    paddingHorizontal: 18,
    paddingVertical: 30, // gives breathing room top/bottom
  },

  // Glass cards
  titleGlass: {
    width: "92%",
    maxWidth: 760,
    flexDirection: "row",
    alignItems: "center",
    justifyContent: "center",
    gap: 18,
    paddingVertical: 22,
    paddingHorizontal: 22,
    borderRadius: 18,
    borderWidth: 1.5,
    borderColor: "rgba(0,0,0,0.35)",
    backgroundColor: "rgba(255,255,255,0.35)",
    marginBottom: 18,
  },
  titleText: {
    fontSize: 64,
    fontWeight: "800",
    letterSpacing: 0.5,
    color: "rgba(0,0,0,0.85)",
  },

  autoGlass: {
    width: "70%",
    maxWidth: 560,
    paddingVertical: 14,
    paddingHorizontal: 18,
    borderRadius: 14,
    borderWidth: 1.5,
    borderColor: "rgba(0,0,0,0.25)",
    backgroundColor: "rgba(255,255,255,0.35)",
    alignItems: "center",
  },
  autoText: {
    fontSize: 34,
    color: "rgba(0,0,0,0.75)",
  },
  autoTimeText: {
    fontWeight: "900",
    color: "rgba(0,0,0,0.85)",
  },

  topArea: {
    width: "100%",
    alignItems: "center",
    paddingTop: 40, // move header group down a bit from the very top
    gap: 18,
  },

  bottomArea: {
    alignItems: "center",
    justifyContent: "center",
    gap: 12,
    paddingBottom: 20,
    },
  arrowButton: {
    width: 108,
    height: 108,
    borderRadius: 999,
    alignItems: "center",
    justifyContent: "center",
    backgroundColor: "rgba(0, 255, 240, 0.9)",
    borderWidth: 1.5,
    borderColor: "rgba(0,0,0,0.25)",
  },
  arrowText: {
    fontSize: 64,
    lineHeight: 70,
    color: "rgba(0,0,0,0.85)",
    fontWeight: "700",
  },
  enterPill: {
    paddingVertical: 8,
    paddingHorizontal: 18,
    borderRadius: 12,
    borderWidth: 1.5,
    borderColor: "rgba(0,0,0,0.25)",
    backgroundColor: "rgba(255,255,255,0.35)",
  },
  enterText: {
    fontSize: 28,
    color: "rgba(0,0,0,0.75)",
    fontWeight: "600",
  },
  logo: {
  width: 75,
  height: 75,
  resizeMode: "contain",
},
});