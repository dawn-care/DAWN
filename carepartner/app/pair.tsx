import React, { useEffect, useState } from "react";
import {
  ActivityIndicator,
  KeyboardAvoidingView,
  Platform,
  Pressable,
  SafeAreaView,
  StyleSheet,
  Text,
  View,
} from "react-native";
import { Stack, router, type Href } from "expo-router";

type CodeResult =
  | { ok: true; pairingCode: string; expiresInSec?: number }
  | { ok: false; message: string };

type VerifyResult =
  | { ok: true; deviceId: string }
  | { ok: false; message: string };

// ---- STUB BACKEND CALLS (replace later) ----
async function fetchPairingCodeForThisDevice(): Promise<CodeResult> {
  // TODO: replace with real API call
  // const res = await fetch("https://your-api/device/pairing-code", { method: "POST" });
  // return await res.json();

  await new Promise((r) => setTimeout(r, 700));
  return { ok: true, pairingCode: "ABC123", expiresInSec: 600 };
}

async function verifyThisDeviceIsPaired(): Promise<VerifyResult> {
  // TODO: replace with real API call
  // const res = await fetch("https://your-api/device/verify-pairing", { method: "POST" });
  // if (!res.ok) return { ok: false, message: "Not paired yet." };
  // return await res.json();

  await new Promise((r) => setTimeout(r, 800));
  // Stub behavior: pretend the caregiver already used the code.
  return { ok: true, deviceId: "device_123" };
}

function formatCountdown(totalSeconds: number) {
  const sec = Math.max(0, Math.floor(totalSeconds));
  const m = Math.floor(sec / 60);
  const s = sec % 60;
  return `${m}:${String(s).padStart(2, "0")}`;
}

export default function PairScreen() {
  const [pairingCode, setPairingCode] = useState<string | null>(null);
  const [expiresInSec, setExpiresInSec] = useState<number | null>(null);

  const [isLoadingCode, setIsLoadingCode] = useState(true);
  const [isVerifying, setIsVerifying] = useState(false);

  const [error, setError] = useState<string | null>(null);

  const canVerify = !!pairingCode && !isLoadingCode && !isVerifying;

  const loadCode = async () => {
    setError(null);
    setIsLoadingCode(true);
    try {
      const result = await fetchPairingCodeForThisDevice();
      if (!result.ok) {
        setError(result.message);
        setPairingCode(null);
        setExpiresInSec(null);
        return;
      }

      setPairingCode(result.pairingCode);
      setExpiresInSec(result.expiresInSec ?? null);
    } catch {
      setError("Could not load pairing code. Please check your connection.");
      setPairingCode(null);
      setExpiresInSec(null);
    } finally {
      setIsLoadingCode(false);
    }
  };

  useEffect(() => {
    loadCode();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  // Simple countdown (optional)
  useEffect(() => {
    if (expiresInSec == null) return;
    if (expiresInSec <= 0) return;

    const id = setInterval(() => {
      setExpiresInSec((t) => (t == null ? t : Math.max(0, t - 1)));
    }, 1000);

    return () => clearInterval(id);
  }, [expiresInSec]);

  // If it expires, clear code + show message (optional)
  useEffect(() => {
    if (expiresInSec == null) return;
    if (expiresInSec > 0) return;

    // Expired
    setPairingCode(null);
    setError("This code expired. Tap Refresh Code to generate a new one.");
  }, [expiresInSec]);

  const handleVerify = async () => {
    setError(null);
    setIsVerifying(true);

    try {
      const result = await verifyThisDeviceIsPaired();
      if (!result.ok) {
        setError(result.message);
        return;
      }

      // TODO: store deviceId / auth token returned from backend
      // Example: await SecureStore.setItemAsync("deviceId", result.deviceId)

      // Next step: Terms & Conditions screen
      router.push("/terms" as Href);
    } catch {
      setError("Something went wrong. Please check your connection and try again.");
    } finally {
      setIsVerifying(false);
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
              On the caregiver’s device, open Start-Up and enter this code to link this
              screen/device. Then tap <Text style={styles.bold}>Verify</Text> here.
            </Text>

            {/* Code display */}
            <View style={styles.codeBlock}>
              <Text style={styles.label}>Pairing Code</Text>

              <View style={styles.codeBox}>
                {isLoadingCode ? (
                  <View style={styles.loadingRow}>
                    <ActivityIndicator />
                    <Text style={styles.codeLoadingText}>Generating…</Text>
                  </View>
                ) : pairingCode ? (
                  <Text style={styles.codeText}>{pairingCode}</Text>
                ) : (
                  <Text style={styles.codeMissingText}>—</Text>
                )}
              </View>

              <View style={styles.codeMetaRow}>
                {expiresInSec != null && pairingCode ? (
                  <Text style={styles.helper}>
                    Expires in {formatCountdown(expiresInSec)}
                  </Text>
                ) : (
                  <Text style={styles.helper}>Code updates are handled by backend.</Text>
                )}

                <Pressable
                  onPress={loadCode}
                  disabled={isLoadingCode || isVerifying}
                  style={({ pressed }) => [
                    styles.linkButton,
                    (isLoadingCode || isVerifying) && { opacity: 0.5 },
                    pressed && !(isLoadingCode || isVerifying) && { opacity: 0.75 },
                  ]}
                >
                  <Text style={styles.linkText}>Refresh Code</Text>
                </Pressable>
              </View>

              {!!error && <Text style={styles.error}>{error}</Text>}
            </View>

            {/* Verify button */}
            <Pressable
              onPress={handleVerify}
              disabled={!canVerify}
              style={({ pressed }) => [
                styles.button,
                !canVerify && styles.buttonDisabled,
                pressed && canVerify && { transform: [{ scale: 0.99 }] },
              ]}
            >
              {isVerifying ? (
                <View style={styles.loadingRow}>
                  <ActivityIndicator />
                  <Text style={styles.buttonText}>Verifying…</Text>
                </View>
              ) : (
                <Text style={styles.buttonText}>Verify</Text>
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
  bold: { fontWeight: "800", color: "rgba(0,0,0,0.75)" },

  codeBlock: { gap: 8 },
  label: {
    fontSize: 14,
    fontWeight: "700",
    color: "rgba(0,0,0,0.7)",
  },
  codeBox: {
    height: 64,
    borderRadius: 16,
    borderWidth: 1.5,
    borderColor: "rgba(0,0,0,0.18)",
    paddingHorizontal: 14,
    alignItems: "center",
    justifyContent: "center",
    backgroundColor: "rgba(255,255,255,0.95)",
  },
  codeText: {
    fontSize: 34,
    fontWeight: "900",
    letterSpacing: 4,
    color: "rgba(0,0,0,0.85)",
  },
  codeMissingText: {
    fontSize: 26,
    fontWeight: "800",
    color: "rgba(0,0,0,0.35)",
  },
  codeLoadingText: {
    fontSize: 16,
    fontWeight: "700",
    color: "rgba(0,0,0,0.55)",
  },

  codeMetaRow: {
    flexDirection: "row",
    alignItems: "center",
    justifyContent: "space-between",
    gap: 12,
  },
  helper: {
    fontSize: 13,
    color: "rgba(0,0,0,0.55)",
    flex: 1,
  },

  linkButton: {
    paddingVertical: 6,
    paddingHorizontal: 10,
    borderRadius: 10,
    backgroundColor: "rgba(0,0,0,0.06)",
    borderWidth: 1,
    borderColor: "rgba(0,0,0,0.08)",
  },
  linkText: {
    fontSize: 13,
    fontWeight: "800",
    color: "rgba(0,0,0,0.7)",
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