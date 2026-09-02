/**
 * photo.tsx
 * 
 * User: Carepartner
 * 
 * Displays photo-based content within the agenda flow.
 * 
 * Responsibilities:
 * - Render photos from backend media endpoints
 * - Support both photo and message modules
 * - Convert metadata into usable image URLs
 * - Display fallback UI when no image is available
 * 
 * Navigation:
 * - Controlled by AgendaView
 * 
 * Notes:
 * - Uses backend `/photo/file?id=` endpoint for photos
 * - Message modules may include image URIs
 * - Only displays one image at a time (future: carousel support)
 */

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

/* ============================
   Types
============================ */

/**
 * Normalized configuration used internally by this component.
 */
type PhotoViewerConfig = {
  photoUrls?: string[];
};

/**
 * Props passed from AgendaView.
 */
interface PhotoViewerProps {
  configuration: ConfigurationWithMetadata;
}

/* ============================
   Helpers
============================ */

/**
 * Reusable glass-style container.
 */
function GlassCard({
  children,
  style,
}: {
  children: React.ReactNode;
  style?: any;
}) {
  return <View style={[styles.glassBase, style]}>{children}</View>;
}

/* ============================
   Component
============================ */

/**
 * PhotoViewerScreen Component
 * 
 * Handles:
 * - Parsing configuration metadata
 * - Building backend image URLs
 * - Rendering image or fallback UI
 */
export default function PhotoViewerScreen({
  configuration: configuration,
  }: PhotoViewerProps){

  /**
   * Converts backend configuration into a usable photo URL list.
   * 
   * Handles:
   * - Photo modules → fetch from `/photo/file`
   * - Message modules → use `image_uri` if present
   */
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

  /* ============================
     State
  ============================ */
  const [config, setConfig] = useState<PhotoViewerConfig>(
    buildConfig(configuration)
  );
  const hasPhotos = !!config.photoUrls?.length;
  const currentUrl = hasPhotos ? config.photoUrls![0] : undefined;

  /* ============================
     Effects
  ============================ */

  /**
   * Rebuild config whenever the active agenda item changes.
   */
  useEffect(() => {
    const next = buildConfig(configuration);
    setConfig(next);
  }, [configuration]);

  /* ============================
     Render
  ============================ */
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
    objectFit: "contain",
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