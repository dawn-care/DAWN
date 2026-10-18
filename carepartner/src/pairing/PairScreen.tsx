import React from "react";
import { Pressable, StyleSheet, Text, useWindowDimensions, View } from "react-native";
import { useNow } from "../clock/useNow";
import { fonts } from "../theme/fonts";
import { borders, colors, fontSizes, legibility, radii, space } from "../theme/theme";
import { fitFontSize, GLYPH_WIDTH } from "../widgets/fitText";
import type { PairingView } from "./usePairing";

/** Extra space between the code's characters, as a share of the font size. */
const CODE_SPACING = 0.12;

/** The most of the screen's height the code may take, leaving room for the rest. */
const CODE_HEIGHT_SHARE = 0.22;

/** "6:42" for the time left until `expiresAt`. */
export function timeLeft(expiresAt: Date, now: Date): string {
  const seconds = Math.max(0, Math.ceil((expiresAt.getTime() - now.getTime()) / 1000));
  return `${Math.floor(seconds / 60)}:${String(seconds % 60).padStart(2, "0")}`;
}

type Props = { view: PairingView; onNewCode: () => void };

/**
 * The code a caregiver enters in the family app to pair this tablet. The
 * caregiver is in the room while pairing, so unlike the resident's screens it
 * has a button; pairing itself needs no tap.
 */
export default function PairScreen({ view, onNewCode }: Props) {
  const current = useNow(1000);
  const { width, height } = useWindowDimensions();

  return (
    <View testID="pair-screen" style={styles.root}>
      <Text style={styles.title}>Set up this tablet</Text>
      {view.kind === "showing" ? (
        <Code code={view.code} expiresIn={timeLeft(view.expiresAt, current)} screen={{ width, height }} onNewCode={onNewCode} />
      ) : (
        <Text style={styles.status}>
          {view.kind === "offline" ? "Can't reach the server. Trying again…" : view.kind === "paired" ? "Paired" : "Getting a code…"}
        </Text>
      )}
    </View>
  );
}

type CodeProps = { code: string; expiresIn: string; screen: { width: number; height: number }; onNewCode: () => void };

function Code({ code, expiresIn, screen, onNewCode }: CodeProps) {
  // The letter spacing widens each character, so fit against a narrower width.
  const widthForCode = (screen.width - 2 * space.lg) / (1 + CODE_SPACING / GLYPH_WIDTH);
  const largest = Math.min(fontSizes.day, Math.floor(screen.height * CODE_HEIGHT_SHARE));
  const size = fitFontSize(code, widthForCode, largest, legibility.minOrientationText);

  return (
    <>
      <Text style={styles.instruction}>Enter this code in the family app</Text>
      <Text testID="pairing-code" style={[styles.code, { fontSize: size, letterSpacing: size * CODE_SPACING }]}>
        {code}
      </Text>
      <Text style={styles.expiry}>Expires in {expiresIn}</Text>
      <Pressable
        accessibilityRole="button"
        onPress={onNewCode}
        style={({ pressed }) => [styles.button, pressed && styles.buttonPressed]}
      >
        <Text style={styles.buttonText}>New code</Text>
      </Pressable>
    </>
  );
}

const styles = StyleSheet.create({
  root: {
    flex: 1,
    alignItems: "center",
    justifyContent: "center",
    gap: space.sm,
    padding: space.lg,
    backgroundColor: colors.day.background,
  },
  title: { fontFamily: fonts.bold, fontSize: fontSizes.tileWord, color: colors.day.text },
  instruction: { fontFamily: fonts.regular, fontSize: fontSizes.name, color: colors.day.textSecondary },
  code: { fontFamily: fonts.heavy, color: colors.day.text },
  expiry: { fontFamily: fonts.regular, fontSize: fontSizes.label, color: colors.day.textSecondary },
  status: { fontFamily: fonts.regular, fontSize: fontSizes.name, color: colors.day.textSecondary, textAlign: "center" },
  button: {
    marginTop: space.sm,
    minHeight: 80,
    minWidth: 240,
    paddingHorizontal: space.md,
    alignItems: "center",
    justifyContent: "center",
    borderRadius: radii.tile,
    borderWidth: borders.tile,
    borderColor: colors.day.border,
    backgroundColor: colors.day.surface,
  },
  buttonPressed: { backgroundColor: colors.day.placeholder },
  buttonText: { fontFamily: fonts.bold, fontSize: fontSizes.label, color: colors.day.text },
});
