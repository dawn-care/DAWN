import React, { useMemo, useState } from "react";
import {
  ActivityIndicator,
  KeyboardAvoidingView,
  Platform,
  Pressable,
  SafeAreaView,
  StyleSheet,
  Text,
  TextInput,
  View,
} from "react-native";
import { Stack, router, type Href } from "expo-router";

type PairingResult =
  | { ok: true; deviceId: string }
  | { ok: false; message: string };

// ---- STUB BACKEND CALL (replace later) ----
async function pairDeviceWithCode(pairingCode: string): Promise<PairingResult> {
  // TODO: replace with real API call
  // const res = await fetch("https://your-api/pair", {
  //   method: "POST",
  //   headers: { "Content-Type": "application/json" },
  //   body: JSON.stringify({ pairingCode }),
  // });
  // if (!res.ok) return { ok: false, message: "Invalid pairing code." };
  // return await res.json();

  // stub behavior:
  await new Promise((r) => setTimeout(r, 800));

  // accept a specific code for dev testing:
  if (pairingCode.toUpperCase() === "ABC123") {
    return { ok: true, deviceId: "device_123" };
  }

  return { ok: false, message: "Invalid pairing code. Please try again." };
}

function normalizeCode(raw: string) {
  // uppercase, remove spaces/dashes, keep alphanumeric only
  return raw
    .toUpperCase()
    .replace(/[\s-]/g, "")
    .replace(/[^A-Z0-9]/g, "")
    .slice(0, 8); // allow up to 8 just in case your backend uses that later
}

export default function PairScreen() {
  const [code, setCode] = useState("");
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const cleanedCode = useMemo(() => normalizeCode(code), [code]);
  const canSubmit = cleanedCode.length >= 6 && !isSubmitting;

  const handleSubmit = async () => {
    setError(null);
    setIsSubmitting(true);

    try {
      const result = await pairDeviceWithCode(cleanedCode);
      if (!result.ok) {
        setError(result.message);
        return;
      }

      // TODO: store deviceId / auth token returned from backend
      // Example: await SecureStore.setItemAsync("deviceId", result.deviceId)

      // Next step per your manual: Terms & Conditions screen (stub route for now)
      router.push("/terms" as Href);
    } catch {
      setError("Something went wrong. Please check your connection and try again.");
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <>
      <Stack.Screen options={{ headerShown: false }} />

      <SafeAreaView style={styles.safe}>
        <KeyboardAvoidingView
          style={styles.container}
          behavior={Platform.select({ ios: "padding", android: undefined })}
        >
          <View style={styles.card}>
            <Text style={styles.title}>Pair Device</Text>

            <Text style={styles.subtitle}>
              Enter the pairing code from your caregiver’s account page to link this device.
            </Text>

            <View style={styles.inputGroup}>
              <Text style={styles.label}>Pairing Code</Text>

              <TextInput
                value={code}
                onChangeText={(t) => setCode(normalizeCode(t))}
                placeholder="e.g. ABC123"
                placeholderTextColor="rgba(0,0,0,0.35)"
                autoCapitalize="characters"
                autoCorrect={false}
                keyboardType="default"
                returnKeyType="done"
                onSubmitEditing={() => {
                  if (canSubmit) handleSubmit();
                }}
                style={styles.input}
                maxLength={8}
              />

              <Text style={styles.helper}>
                Tip: codes are usually 6–8 letters/numbers.
              </Text>

              {!!error && <Text style={styles.error}>{error}</Text>}
            </View>

            <Pressable
              onPress={handleSubmit}
              disabled={!canSubmit}
              style={({ pressed }) => [
                styles.button,
                !canSubmit && styles.buttonDisabled,
                pressed && canSubmit && { transform: [{ scale: 0.99 }] },
              ]}
            >
              {isSubmitting ? (
                <View style={styles.loadingRow}>
                  <ActivityIndicator />
                  <Text style={styles.buttonText}>Pairing…</Text>
                </View>
              ) : (
                <Text style={styles.buttonText}>Pair Device</Text>
              )}
            </Pressable>

            <Text style={styles.footerNote}>
              Keep the device plugged in and leave the Start-Up app open to ensure it plays
              automatically each morning.
            </Text>
          </View>
        </KeyboardAvoidingView>
      </SafeAreaView>
    </>
  );
}

const styles = StyleSheet.create({
  safe: { flex: 1, backgroundColor: "#fff" },
  container: {
    flex: 1,
    paddingHorizontal: 18,
    alignItems: "center",
    justifyContent: "center",
  },

  card: {
    width: "100%",
    maxWidth: 520,
    borderRadius: 18,
    borderWidth: 1.5,
    borderColor: "rgba(0,0,0,0.15)",
    backgroundColor: "rgba(255,255,255,0.9)",
    padding: 18,
    gap: 14,
  },

  title: {
    fontSize: 34,
    fontWeight: "800",
    color: "rgba(0,0,0,0.85)",
  },
  subtitle: {
    fontSize: 16,
    lineHeight: 22,
    color: "rgba(0,0,0,0.65)",
  },

  inputGroup: { gap: 8 },
  label: {
    fontSize: 14,
    fontWeight: "700",
    color: "rgba(0,0,0,0.7)",
  },
  input: {
    height: 54,
    borderRadius: 14,
    borderWidth: 1.5,
    borderColor: "rgba(0,0,0,0.18)",
    paddingHorizontal: 14,
    fontSize: 22,
    fontWeight: "800",
    letterSpacing: 2,
    color: "rgba(0,0,0,0.85)",
    backgroundColor: "rgba(255,255,255,0.95)",
  },
  helper: {
    fontSize: 13,
    color: "rgba(0,0,0,0.55)",
  },
  error: {
    fontSize: 14,
    color: "#B00020",
    fontWeight: "600",
  },

  button: {
    height: 54,
    borderRadius: 16,
    backgroundColor: "rgba(0, 255, 240, 0.9)",
    borderWidth: 1.5,
    borderColor: "rgba(0,0,0,0.18)",
    alignItems: "center",
    justifyContent: "center",
  },
  buttonDisabled: {
    backgroundColor: "rgba(0,0,0,0.08)",
  },
  buttonText: {
    fontSize: 18,
    fontWeight: "800",
    color: "rgba(0,0,0,0.8)",
  },
  loadingRow: {
    flexDirection: "row",
    alignItems: "center",
    gap: 10,
  },

  footerNote: {
    marginTop: 6,
    fontSize: 13,
    lineHeight: 18,
    color: "rgba(0,0,0,0.55)",
  },
});