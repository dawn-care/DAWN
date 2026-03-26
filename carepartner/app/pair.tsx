import React, { useEffect, useState } from "react";
import {
  ActivityIndicator,
  Pressable,
  SafeAreaView,
  ScrollView,
  StyleSheet,
  Text,
  useWindowDimensions,
  View,
} from "react-native";
import { Stack, router, type Href } from "expo-router";
import { API_URL } from "./config";
import { getCarepartnerId, storeCarepartnerId } from "./carepartnerId";

type CodeResult =
  | { ok: true; pairingCode: string; expiresInSec?: number }
  | { ok: false; message: string };

type VerifyResult =
  | { ok: true; deviceId: string }
  | { ok: false; message: string };

const API_BASE = `${API_URL}/users`;

// ---- API CALLS ----
async function createNewCarepartner(): Promise<number> {
  const res = await fetch(`${API_BASE}/carepartner/create-carepartner`, { 
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({Name: "PlaceHolder Name", AutoStart: "08:00:00"}), 
  });
  if (!res.ok) throw new Error("Failed to create carepartner");
  const data = await res.json();
  return data.carepartner_id;
}

async function fetchPairingCodeForThisDevice(carepartnerID: number): Promise<CodeResult> {
  try {
    const res = await fetch(`${API_BASE}/device-pairing/code`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ carepartner_id: carepartnerID }),
    });

    if (!res.ok) {
      const errorData = await res.json();
      return { ok: false, message: errorData.error || "Failed to generate code" };
    }

    const data = await res.json();
    return { ok: true, pairingCode: data.pairing_code, expiresInSec: data.expires_in_sec };
  } catch {
    return { ok: false, message: "Network error while generating code" };
  }
}

async function verifyThisDeviceIsPaired(carepartnerID: number): Promise<VerifyResult> {
  try {
    const res = await fetch(`${API_BASE}/device-pairing/status`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ carepartner_id: carepartnerID }),
    });

    if (!res.ok) {
      const errorData = await res.json();
      return { ok: false, message: errorData.error || "Failed to check pairing status" };
    }

    const data = await res.json();
    return { ok: true, deviceId: data.paired ? "device_123" : "" };
  } catch {
    return { ok: false, message: "Network error while checking device pairing" };
  }
}

async function carepartnerExists(id: number): Promise<boolean> {
  try {
    const res = await fetch(`${API_BASE}/carepartner/${id}`);

    if (res.status === 404) return false;
    if (!res.ok) return false;

    return true;
  } catch {
    return false;
  }
}

function formatCountdown(totalSeconds: number) {
  const sec = Math.max(0, Math.floor(totalSeconds));
  const m = Math.floor(sec / 60);
  const s = sec % 60;
  return `${m}:${String(s).padStart(2, "0")}`;
}

export default function PairScreen() {
  const { height, width } = useWindowDimensions();
  const [pairingCode, setPairingCode] = useState<string | null>(null);
  const [expiresInSec, setExpiresInSec] = useState<number | null>(null);
  const [carepartnerID, setCarepartnerID] = useState<number | null>(null);

  const [isLoadingCode, setIsLoadingCode] = useState(true);
  const [isVerifying, setIsVerifying] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const canVerify = !!pairingCode && !isLoadingCode && !isVerifying;
  const isCompactScreen = height < 760 || width < 520;
  const isShortScreen = height < 640;

  // ---- Load or create carepartnerID ----
  useEffect(() => {
    async function init() {
      setIsLoadingCode(true);
      setError(null);

      try {
        let id = await getCarepartnerId();

        if (id) {
          const valid = await carepartnerExists(id);
          if (!valid) {
            id = null;
          }
        }

        if (!id) {
          id = await createNewCarepartner();
          await storeCarepartnerId(id);
        }

        setCarepartnerID(id);

        const result = await fetchPairingCodeForThisDevice(id);
        if (!result.ok) {
          setError(result.message);
          setPairingCode(null);
          setExpiresInSec(null);
          return;
        }

        setPairingCode(result.pairingCode);
        setExpiresInSec(result.expiresInSec ?? null);

      } catch (err: any) {
        setError(err.message || "Failed to initialize carepartner ID");
      } finally {
        setIsLoadingCode(false);
      }
    }

    init();
  }, []);

  // Countdown for expiration
  useEffect(() => {
    if (expiresInSec == null || expiresInSec <= 0) return;
    const id = setInterval(() => {
      setExpiresInSec((t) => (t == null ? t : Math.max(0, t - 1)));
    }, 1000);
    return () => clearInterval(id);
  }, [expiresInSec]);

  // Expired handling
  useEffect(() => {
    if (expiresInSec == null || expiresInSec > 0) return;
    setPairingCode(null);
    setError("This code expired. Tap Refresh Code to generate a new one.");
  }, [expiresInSec]);

  const handleVerify = async () => {
    if (!carepartnerID) {
      setError("Carepartner ID not initialized");
      return;
    }

    setError(null);
    setIsVerifying(true);

    try {
      if (!pairingCode) {
        setError("No pairing code available");
        return;
      }

      const result = await verifyThisDeviceIsPaired(carepartnerID);
      if (!result.ok) {
        setError(result.message);
        return;
      }

      // TODO: store deviceId if needed
      // await SecureStore.setItemAsync("deviceId", result.deviceId)

      router.push("/terms" as Href);
    } catch {
      setError("Something went wrong. Please check your connection and try again.");
    } finally {
      setIsVerifying(false);
    }
  };

  const loadCode = async () => {
    if (!carepartnerID) return;
    setError(null);
    setIsLoadingCode(true);
    try {
      const result = await fetchPairingCodeForThisDevice(carepartnerID);
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

  return (
    <>
      <Stack.Screen options={{ headerShown: false }} />

      <SafeAreaView style={styles.safe}>
        <ScrollView
          contentContainerStyle={[
            styles.container,
            isCompactScreen && styles.containerCompact,
          ]}
          showsVerticalScrollIndicator={false}
        >
          <View style={[styles.card, isCompactScreen && styles.cardCompact]}>
            <Text style={[styles.title, isCompactScreen && styles.titleCompact]}>Pair Device</Text>

            <Text style={[styles.subtitle, isCompactScreen && styles.subtitleCompact]}>
              On the caregiver’s device, open Start-Up and enter this code to link this
              screen/device. Then tap <Text style={styles.bold}>Verify</Text> here.
            </Text>

            {/* Code display */}
            <View style={[styles.codeBlock, isCompactScreen && styles.codeBlockCompact]}>
              <Text style={styles.label}>Pairing Code</Text>

              <View style={[styles.codeBox, isCompactScreen && styles.codeBoxCompact]}>
                {isLoadingCode ? (
                  <View style={styles.loadingRow}>
                    <ActivityIndicator />
                    <Text style={styles.codeLoadingText}>Generating…</Text>
                  </View>
                ) : pairingCode ? (
                  <Text style={[styles.codeText, isCompactScreen && styles.codeTextCompact]}>{pairingCode}</Text>
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

            <View style={[styles.infoBlock, isCompactScreen && styles.infoBlockCompact]}>
              <Text style={styles.infoLabel}>Carepartner ID</Text>
              <Text style={[styles.infoValue, isCompactScreen && styles.infoValueCompact]}>
                {carepartnerID != null ? String(carepartnerID) : "Loading..."}
              </Text>
              <Text style={styles.infoHelper}>
                Temporary debug display for caregiver pairing.
              </Text>
            </View>

            {/* Verify button */}
            <Pressable
              onPress={handleVerify}
              disabled={!canVerify}
              style={({ pressed }) => [
                styles.button,
                isCompactScreen && styles.buttonCompact,
                !canVerify && styles.buttonDisabled,
                pressed && canVerify && { transform: [{ scale: 0.99 }] },
              ]}
            >
              {isVerifying ? (
                <View style={styles.loadingRow}>
                  <ActivityIndicator />
                  <Text style={[styles.buttonText, isCompactScreen && styles.buttonTextCompact]}>Verifying…</Text>
                </View>
              ) : (
                <Text style={[styles.buttonText, isCompactScreen && styles.buttonTextCompact]}>Verify</Text>
              )}
            </Pressable>

            <Pressable
              onPress={() => router.replace("/" as Href)}
              style={({ pressed }) => [
                styles.secondaryButton,
                isCompactScreen && styles.secondaryButtonCompact,
                pressed && { opacity: 0.8 },
              ]}
            >
              <Text style={[styles.secondaryButtonText, isCompactScreen && styles.secondaryButtonTextCompact]}>Return Home</Text>
            </Pressable>

            <Text style={[styles.footerNote, isShortScreen && styles.footerNoteCompact]}>
              Keep the device plugged in and leave the Start-Up app open to ensure it plays
              automatically each morning.
            </Text>
          </View>
        </ScrollView>
      </SafeAreaView>
    </>
  );
}

const styles = StyleSheet.create({
  safe: { flex: 1, backgroundColor: "#fff" },
  container: {
    flexGrow: 1,
    paddingHorizontal: 18,
    paddingVertical: 20,
    alignItems: "center",
    justifyContent: "center",
  },
  containerCompact: {
    justifyContent: "flex-start",
    paddingVertical: 6,
    paddingHorizontal: 10,
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
  cardCompact: {
    maxWidth: 460,
    padding: 10,
    gap: 4,
    borderRadius: 14,
  },

  title: {
    fontSize: 34,
    fontWeight: "800",
    color: "rgba(0,0,0,0.85)",
  },
  titleCompact: {
    fontSize: 22,
  },
  subtitle: {
    fontSize: 16,
    lineHeight: 22,
    color: "rgba(0,0,0,0.65)",
  },
  subtitleCompact: {
    fontSize: 12,
    lineHeight: 16,
  },
  bold: { fontWeight: "800", color: "rgba(0,0,0,0.75)" },

  codeBlock: { gap: 8 },
  codeBlockCompact: { gap: 3 },
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
  codeBoxCompact: {
    height: 44,
    borderRadius: 12,
  },
  codeText: {
    fontSize: 34,
    fontWeight: "900",
    letterSpacing: 4,
    color: "rgba(0,0,0,0.85)",
  },
  codeTextCompact: {
    fontSize: 22,
    letterSpacing: 2,
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

  infoBlock: {
    gap: 4,
    padding: 12,
    borderRadius: 14,
    borderWidth: 1,
    borderColor: "rgba(0,0,0,0.12)",
    backgroundColor: "rgba(0,0,0,0.03)",
  },
  infoBlockCompact: {
    gap: 2,
    padding: 8,
    borderRadius: 10,
  },
  infoLabel: {
    fontSize: 13,
    fontWeight: "700",
    color: "rgba(0,0,0,0.6)",
  },
  infoValue: {
    fontSize: 22,
    fontWeight: "900",
    color: "rgba(0,0,0,0.85)",
  },
  infoValueCompact: {
    fontSize: 15,
  },
  infoHelper: {
    fontSize: 12,
    color: "rgba(0,0,0,0.5)",
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
  buttonCompact: {
    height: 40,
    borderRadius: 12,
  },
  buttonDisabled: {
    backgroundColor: "rgba(0,0,0,0.08)",
  },
  buttonText: {
    fontSize: 18,
    fontWeight: "800",
    color: "rgba(0,0,0,0.8)",
  },
  buttonTextCompact: {
    fontSize: 15,
  },
  secondaryButton: {
    height: 50,
    borderRadius: 16,
    backgroundColor: "rgba(0,0,0,0.06)",
    borderWidth: 1.5,
    borderColor: "rgba(0,0,0,0.12)",
    alignItems: "center",
    justifyContent: "center",
  },
  secondaryButtonCompact: {
    height: 38,
    borderRadius: 12,
  },
  secondaryButtonText: {
    fontSize: 16,
    fontWeight: "800",
    color: "rgba(0,0,0,0.7)",
  },
  secondaryButtonTextCompact: {
    fontSize: 13,
  },
  loadingRow: {
    flexDirection: "row",
    alignItems: "center",
    gap: 6,
  },

  footerNote: {
    marginTop: 6,
    fontSize: 13,
    lineHeight: 18,
    color: "rgba(0,0,0,0.55)",
  },
  footerNoteCompact: {
    marginTop: 0,
    fontSize: 11,
    lineHeight: 14,
  },
});