/**
 * screensaver.tsx
 * 
 * User: Carepartner
 * 
 * The main idle/home screen for the device.
 * 
 * Responsibilities:
 * - Display Start-Up branding and auto-start time
 * - Allow manual entry into the daily routine
 * - Provide access to syncing schedule/content
 * - Provide access to device pairing
 * 
 * Navigation:
 * - On "Enter Start-Up" → redirects to `/agendaView`
 * - On "Pair Device" → redirects to `/pair`
 * 
 * Notes:
 * - Auto-start time is formatted from backend (HH:MM:SS → 12-hour)
 * - Sync button refreshes schedule data from backend
 */

import React, { useEffect, useMemo, useState } from "react";
import { Stack } from "expo-router";
import {
  ImageBackground,
  Pressable,
  SafeAreaView,
  StyleSheet,
  Text,
  View,
  ActivityIndicator,
} from "react-native";
import { router } from "expo-router";
import { Image } from "react-native";
import { getCarepartnerId } from "../src/pairing/carepartnerId";
import { useSchedule } from "./ScheduleContext";

/* ============================
   Helpers
============================ */

/**
 * Converts backend time (HH:MM:SS) into user-friendly 12-hour format.
 * Example:
 *  "08:00:00" → "8:00 AM"
 *  "14:30:00" → "2:30 PM"
 */
function formatTime(timeString: string) {
  const [hourStr, minuteStr] = timeString.split(':');
  let hour = parseInt(hourStr, 10);
  const minute = parseInt(minuteStr, 10);

  const ampm = hour >= 12 ? 'PM' : 'AM';
  hour = hour % 12 || 12; // convert 0 -> 12

  return `${hour}:${minute.toString().padStart(2, '0')} ${ampm}`;
}

/* ============================
   Component
============================ */

/**
 * Screensaver Component
 * 
 * Handles:
 *  - Home screen UI
 *  - Navigation to schedule
 *  - Sync trigger
 *  - Pairing navigation
 */
export default function Screensaver() {

  /* ============================
     State
  ============================ */
  const { carepartner, refresh } = useSchedule();
  const [hasCarepartnerId, setHasCarepartnerId] = useState<boolean | null>(null);
  const [syncing, setSyncing] = useState(false);

  const autoStartTimeLabel = carepartner?.auto_start
    ? formatTime(carepartner.auto_start)
    : '...';

  /* ============================
     Effects
  ============================ */

  /**
   * On mount, check if carepartner ID exists
   */
  useEffect(() => {
    let mounted = true;
    (async () => {
      const carepartnerId = await getCarepartnerId();
      if (mounted) {
        setHasCarepartnerId(carepartnerId != null);
      }
    })();
    return () => { mounted = false; };
  }, []);

  /**
   * Background image for screensaver
   * (Memoized so it doesn't reload on every render)
   */
  const backgroundSource = useMemo(() => {
    return require("../assets/images/defaultBackground.png");
  }, []);

  /* ============================
     Handlers
  ============================ */

  /**
   * Handles navigation to manually start daily agenda
   */
  const handlePrimaryAction = () => {
    router.push({ pathname: "/agendaView", params: { trigger: 'manual' } } as any);
  };

  /**
   * Handles navigation to device pairing screen
   */
  const handlePairDevice = () => {
    router.push({ pathname: "/pair" } as any);
  };

  /**
   * Handles syncing start time and scheduled content with caregiver's account
   */
  const handleSync = async () => {
    setSyncing(true);
    try {
      await refresh();
    } finally {
      setSyncing(false);
    }
  };

  /* ============================
     Render
  ============================ */
  return (
    <>
    <Stack.Screen options={{ headerShown: false }} />
    <View style={styles.root}>
        <ImageBackground source={backgroundSource} style={styles.bg} resizeMode="cover">
        <SafeAreaView style={styles.safe}>
          <View style={styles.topRightActionRow}>
            {/* Sync Button */}
            <Pressable
            onPress={handleSync}
            style={({ pressed }) => [
              styles.topRightButton,
              pressed && { transform: [{ scale: 0.98 }] },
            ]}
            >
            {syncing ? (
              <ActivityIndicator size="small" color="rgba(0,0,0,0.8)" />
            ) : (
              <Text style={styles.topRightButtonText}>Sync</Text>
            )}
            </Pressable>
            {/* Pair Device Button */}
            <Pressable
            onPress={handlePairDevice}
            style={({ pressed }) => [
              styles.topRightButton,
              pressed && { transform: [{ scale: 0.98 }] },
            ]}
            >
            <Text style={styles.topRightButtonText}>Pair Device</Text>
            </Pressable>
          </View>
            <View style={styles.content}>
            {/* TOP */}
            <View style={styles.topArea}>
                {/* App Logo and Title */}
                <View style={styles.titleGlass}>
                <Image source={require("../assets/images/icon.png")} style={styles.logo} />
                <Text style={styles.titleText}>Start-Up</Text>
                </View>

                <View style={styles.autoGlass}>
                <Text style={styles.autoText}>
                    Auto-Starts at{" "}
                    <Text style={styles.autoTimeText}>{autoStartTimeLabel}</Text>
                </Text>
                </View>
            </View>

            {/* BOTTOM */}
            <View style={styles.bottomArea}>
              {/* Manual Start Button */}
              <Pressable
                onPress={handlePrimaryAction}
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
  topRightActionRow: {
    position: "absolute",
    top: 20,
    right: 20,
    zIndex: 2,
    flexDirection: "row",
    gap: 10,
  },
  topRightButton: {
    paddingVertical: 10,
    paddingHorizontal: 16,
    borderRadius: 14,
    borderWidth: 1.5,
    borderColor: "rgba(0,0,0,0.25)",
    backgroundColor: "rgba(255,255,255,0.5)",
  },
  topRightButtonText: {
    fontSize: 20,
    fontWeight: "700",
    color: "rgba(0,0,0,0.8)",
  },
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