import React, { useEffect, useState } from "react";
import {
  Image,
  ScrollView,
  StyleSheet,
  Text,
  View,
} from "react-native";
import { Stack } from "expo-router";
import { API_URL } from "../src/api/config";
import {
  type ConfigurationWithMetadata,
  EventMetadata,
  MessageMetadata,
} from "./agendaView";

type TextViewerConfig = {
  content: string;
  imageUrl: string | null;
  hasImage: boolean;
};

interface TextViewerProps {
  configuration: ConfigurationWithMetadata;
}

async function buildConfig(
  configuration: ConfigurationWithMetadata
): Promise<TextViewerConfig> {
  if (configuration.module_type === "messages") {
    const metadata = configuration.metadata as MessageMetadata | null;
    const imagePath = metadata?.image_uri?.trim();
    const imageUrl = imagePath
      ? imagePath.startsWith("http://") || imagePath.startsWith("https://")
        ? imagePath
        : `${API_URL}${imagePath}`
      : null;

    return {
      content: metadata?.message ?? "",
      imageUrl,
      hasImage: !!imageUrl,
    };
  }

  const metadata = configuration.metadata as EventMetadata | null;

  return {
    content: metadata?.text ?? "",
    imageUrl: null,
    hasImage: false,
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
    imageUrl: null,
    hasImage: false,
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
            {config.hasImage && config.imageUrl ? (
              <View style={styles.splitLayout}>
                <ScrollView
                  style={styles.textColumn}
                  contentContainerStyle={styles.textColumnContent}
                  showsVerticalScrollIndicator={false}
                >
                  {loading && !config.content ? (
                    <Text style={[styles.contentText, styles.leftAlignedText]}>Loading…</Text>
                  ) : (
                    <Text style={[styles.contentText, styles.leftAlignedText]}>{config.content}</Text>
                  )}
                </ScrollView>

                <View style={styles.imageColumn}>
                  <Image
                    source={{ uri: config.imageUrl }}
                    style={styles.image}
                    resizeMode="cover"
                  />
                </View>
              </View>
            ) : (
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
            )}
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
  splitLayout: {
    flex: 1,
    flexDirection: "row",
    alignItems: "stretch",
  },
  textColumn: {
    flex: 2,
    marginRight: 16,
  },
  textColumnContent: {
    flexGrow: 1,
    justifyContent: "center",
    alignItems: "center",
  },
  imageColumn: {
    flex: 3,
    alignItems: "center",
    justifyContent: "center",
    borderLeftWidth: 1,
    borderLeftColor: "rgba(0,0,0,0.12)",
    paddingLeft: 16,
    overflow: "hidden",
  },
  image: {
    width: "100%",
    height: "100%",
    borderRadius: 12,
  },
  scrollContent: {
    flexGrow: 1,
    justifyContent: "center",
    alignItems: "center",
    paddingBottom: 20,
  },
  contentText: {
    fontSize: 26,
    lineHeight: 36,
    color: "rgba(0,0,0,0.82)",
    fontWeight: "500",
    textAlign: "center",
  },
  leftAlignedText: {
    textAlign: "left",
  },
});