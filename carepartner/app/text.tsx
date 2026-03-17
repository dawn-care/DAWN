// app/text.tsx (or whatever route you want)
import React, { useEffect, useState } from "react";
import {
  ScrollView,
  StyleSheet,
  Text,
  View,
} from "react-native";
import { Stack } from "expo-router";
import { type ConfigurationWithMetadata, WeatherMetadata, NewsMetadata, EventMetadata } from "./agendaView";
import { API_URL } from "./config";

type TextViewerConfig = {
  content: string;
};

interface TextViewerProps {
  configuration: ConfigurationWithMetadata;
}

async function fetchWeather(metadata: WeatherMetadata): Promise<string> {
  var response = "";
  try{
    const res = await fetch(`${API_URL}/weather/overview`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(metadata),
    });
    response = await res.text();
  } catch (err: any){
    console.error("Failed getting weather overview: " + err.message);
  }
  return response;
}

async function fetchSports(metadata: NewsMetadata): Promise<string>{
  var response = "";
  try{
    const res = await fetch(`${API_URL}/sports/team-game`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(metadata),
    });
    response = await res.text();
  } catch (err: any){
    console.error("Failed getting sports overview: " + err.message);
  }
  return response;
}

async function fetchNews(): Promise<string>{
  var response = "";
  try{
    const res = await fetch(`${API_URL}/news/good-news`);
    response = await res.text();
  } catch (err: any){
    console.error("Failed getting sports overview: " + err.message);
  }
  return response;
}

async function buildConfig(
  configuration: ConfigurationWithMetadata
): Promise<TextViewerConfig> {
  let metadata;

  if (configuration.module_type === "weather") {
    metadata = configuration.metadata as WeatherMetadata | null;

    return {
      content: metadata ? await fetchWeather(metadata) : "",
    };
  }

  if (configuration.module_type === "sports") {
    metadata = configuration.metadata as NewsMetadata | null;

    return {
      content: metadata ? await fetchSports(metadata) : "",
    };
  }

  if (configuration.module_type === "news") {
    return {
      content: await fetchNews(),
    };
  }

  metadata = configuration.metadata as EventMetadata | null;

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
    paddingBottom: 20,
  },
  contentText: {
    fontSize: 26,
    lineHeight: 36,
    color: "rgba(0,0,0,0.82)",
    fontWeight: "500",
  },
});