// app/photos.tsx (or whatever route you want)
import React, { useEffect, useState } from "react";
import {
  Image,
  StyleSheet,
  Text,
  View,
} from "react-native";
import { Stack } from "expo-router";
import { type ConfigurationWithMetadata, PhotoMetadata, MessageMetadata } from "./agendaView";
import { API_URL } from "./config";

type PhotoViewerConfig = {
  photoUrls?: string[];
};

interface PhotoViewerProps {
  configuration: ConfigurationWithMetadata;
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

export default function PhotoViewerScreen({
  configuration: configuration,
  }: PhotoViewerProps){

  const buildConfig = (configuration: ConfigurationWithMetadata): PhotoViewerConfig => {
    var metadata;
    if (configuration.module_type === "photo"){
      metadata = configuration.metadata as PhotoMetadata | null;
      if (!metadata) {
        return {
          photoUrls: [],
        };
      }

      return {
        photoUrls: [`${API_URL}/photo/file?id=${metadata.id}`],
      };
    } else {
      metadata = configuration.metadata as MessageMetadata | null;
      if (!metadata) {
        return {
          photoUrls: [],
        };
      }

      return {
        photoUrls: metadata.image_uri ? [`${API_URL}${metadata.image_uri}`] : [],
      };
    }
  };

  const [config, setConfig] = useState<PhotoViewerConfig>(
    buildConfig(configuration)
  );

  useEffect(() => {
    const next = buildConfig(configuration);
    setConfig(next);
  }, [configuration]);

  const hasPhotos = !!config.photoUrls?.length;
  const currentUrl = hasPhotos ? config.photoUrls![0] : undefined;
  console.log("PhotoViewer config:", config, "currentUrl:", currentUrl);

  return (
    <>
      <Stack.Screen options={{ headerShown: false }} />

      <View style={styles.safe}>
        <View style={styles.photoWrap}>
          {currentUrl ? (
            <img src={currentUrl} style={styles.photo} onError={(e) => console.log("Image failed:", e.nativeEvent)} />
          ) : (
            <View style={[styles.photo, styles.placeholder]}>
              <Text style={styles.placeholderText}>Photo will load from backend</Text>
              <Text style={styles.placeholderSubText}>
                (Set config.photoUrls once your API is wired)
              </Text>
            </View>
          )}
        </View>
      </View>
    </>
  );
}

const styles = StyleSheet.create({
  safe: { flex: 1, backgroundColor: "#000" },

  // --- photo ---
  photoWrap: {
    flex: 1,
    backgroundColor: "#000",
    justifyContent: "center", 
    alignItems: "center",
  },
  photo: {
    flex: 1,
    width: "100%",
    height: "100%",
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
});