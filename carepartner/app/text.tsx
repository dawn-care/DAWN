import React, { useEffect, useState } from "react";
import {
  ScrollView,
  StyleSheet,
  Text,
  View,
} from "react-native";
import { Stack } from "expo-router";
import { type ConfigurationWithMetadata, EventMetadata } from "./agendaView";

type TextViewerConfig = {
  content: string;
};

interface TextViewerProps {
  configuration: ConfigurationWithMetadata;
}

async function buildConfig(
  configuration: ConfigurationWithMetadata
): Promise<TextViewerConfig> {
  const metadata = configuration.metadata as EventMetadata | null;

  return {
    content: metadata?.text ?? "",
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

export default function TextViewerScreen({
  configuration: configuration
  }: TextViewerProps) {
  const [config, setConfig] = useState<TextViewerConfig>({
    content: "",
  });

  const [loading, setLoading] = useState(true);

  useEffect(() => {
    let mounted = true;

    const load = async () => {
      setLoading(true);

      const next = await buildConfig(configuration);

      if (!mounted) return;

      setConfig(next);
      setLoading(false);
    };

    load();

    return () => {
      mounted = false;
    };
  }, [configuration]);


  return (
    <>
      <Stack.Screen options={{ headerShown: false }} />

      <View style={styles.safe}>
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
      </View>
    </>
  );
}

const styles = StyleSheet.create({
  safe: { flex: 1, backgroundColor: "#000" },

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
    flexGrow: 1,              // KEY: lets content take full height
    justifyContent: "center", // vertical centering
    alignItems: "center",     // horizontal centering
    paddingBottom: 20,
  },
  contentText: {
    fontSize: 26,
    lineHeight: 36,
    color: "rgba(0,0,0,0.82)",
    fontWeight: "500",
    textAlign: "center", // centers multiline text
  },
});