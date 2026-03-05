// app/text.tsx (or whatever route you want)
import React, { useEffect, useState } from "react";
import {
  SafeAreaView,
  ScrollView,
  StyleSheet,
  Text,
  View,
} from "react-native";
import { Stack } from "expo-router";

// ---- STUB TYPES (replace later) ----
type TextViewerConfig = {
  title: string; // e.g. "Today's Message"
  content: string; // long text
};

// ---- STUB BACKEND CALL (replace later) ----
async function fetchTextViewer(): Promise<TextViewerConfig> {
  // TODO: replace with real API call
  // const res = await fetch("https://your-api/text");
  // return await res.json();

  return {
    title: "Today's Message",
    content:
      "Good morning, Grandma.\n\nWe love you so much. Today is going to be a beautiful day.\n\n• Remember to drink some water\n• Your appointment is at 2:30 PM\n\nLove,\nYour family ❤️",
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

export default function TextViewerScreen() {
  const [config, setConfig] = useState<TextViewerConfig>({
    title: "Loading…",
    content: "",
  });

  const [loading, setLoading] = useState(true);

  useEffect(() => {
    let mounted = true;
    (async () => {
      try {
        const next = await fetchTextViewer();
        if (!mounted) return;
        setConfig(next);
      } catch {
        // TODO: log/report
        if (mounted) {
          setConfig({
            title: "Unable to load",
            content: "Please check your connection and try again later.",
          });
        }
      } finally {
        if (mounted) setLoading(false);
      }
    })();
    return () => {
      mounted = false;
    };
  }, []);

  return (
    <>
      <Stack.Screen options={{ headerShown: false }} />

      <SafeAreaView style={styles.safe}>
        {/* Top title strip */}
        <View style={styles.topBar}>
          <Text style={styles.topTitle}>{config.title}</Text>
        </View>

        {/* Content area */}
        <View style={styles.body}>
          <GlassCard style={styles.card}>
            <ScrollView
              contentContainerStyle={styles.scrollContent}
              showsVerticalScrollIndicator={false}
            >
              {loading && !config.content ? (
                <Text style={styles.contentText}>Loading…</Text>
              ) : (
                <Text style={styles.contentText}>{config.content}</Text>
              )}
            </ScrollView>
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
    paddingVertical: 14,
    alignItems: "center",
    justifyContent: "center",
  },
  topTitle: {
    fontSize: 54,
    fontWeight: "300",
    color: "rgba(0,0,0,0.85)",
    textAlign: "center",
  },

  // --- body ---
  body: {
    flex: 1,
    padding: 18,
  },

  // --- glass base ---
  glassBase: {
    borderRadius: 18,
    borderWidth: 1.5,
    borderColor: "rgba(0,0,0,0.25)",
    backgroundColor: "rgba(255,255,255,0.55)",
  },

  // --- content card ---
  card: {
    flex: 1,
    paddingHorizontal: 18,
    paddingVertical: 18,
  },
  scrollContent: {
    paddingBottom: 20,
  },
  contentText: {
    fontSize: 26,
    lineHeight: 36,
    color: "rgba(0,0,0,0.82)",
    fontWeight: "500",
  },
});