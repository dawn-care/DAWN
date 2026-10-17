/**
 * DevToolbar.tsx
 *
 * Developer builds only (`app/_layout.tsx` renders it when `__DEV__`). Two
 * buttons in the top right, above every screen: ☰ opens a panel with the
 * app's state (clock, awake hours, pairing, the screen fetch, today's
 * schedule and the carepartner's media, each playable as a staff play), and
 * ⏭ skips whatever is playing, like a staff stop. Never in a release build.
 */

import React, { useCallback, useEffect, useState } from "react";
import { Pressable, ScrollView, StatusBar, StyleSheet, Text, View } from "react-native";
import { isNight, nightHoursFromAwake } from "@shared";
import { dayOf, fakeStartTime } from "../clock/clock";
import { useNow } from "../clock/useNow";
import { usePlayback } from "../playback/PlaybackProvider";
import { loadPlayed } from "../schedule/playedStore";
import { useScreen } from "../screen/ScreenProvider";
import { fetchConnected, fetchMedia, playStatus, startStaffPlay, type DevMedia } from "./devInfo";

const time = (date: Date) => date.toLocaleTimeString([], { hour: "2-digit", minute: "2-digit", second: "2-digit" });

/** What the panel loads when opened or refreshed. */
type Loaded = { media: DevMedia[] | null; connected: boolean | null; played: Set<string> };

export default function DevToolbar() {
  const [open, setOpen] = useState(false);
  const { state, screen: mode, stop } = usePlayback();

  return (
    <View style={styles.overlay} pointerEvents="box-none">
      <View style={styles.buttons}>
        {state.playing && (
          <Pressable testID="dev-skip" accessibilityLabel="Skip" style={styles.button} onPress={() => stop()}>
            <Text style={styles.buttonText}>⏭</Text>
          </Pressable>
        )}
        <Pressable
          testID="dev-menu"
          accessibilityLabel="Developer menu"
          style={styles.button}
          onPress={() => setOpen((value) => !value)}
        >
          <Text style={styles.buttonText}>☰</Text>
        </Pressable>
      </View>
      {open && <DevPanel mode={mode} />}
    </View>
  );
}

function DevPanel({ mode }: { mode: string }) {
  const { state } = usePlayback();
  const { screen, refresh, fetchedAt } = useScreen();
  const current = useNow(1000);
  const id = state.carepartnerId;
  const [loaded, setLoaded] = useState<Loaded>({ media: null, connected: null, played: new Set() });
  const [message, setMessage] = useState<string | null>(null);

  const load = useCallback(async () => {
    const played = await loadPlayed();
    if (id == null) return setLoaded({ media: null, connected: null, played });
    const [media, connected] = await Promise.all([
      fetchMedia(id).catch(() => null),
      fetchConnected(id).catch(() => null),
    ]);
    setLoaded({ media, connected, played });
  }, [id]);

  useEffect(() => {
    load();
  }, [load]);

  const play = async (media: DevMedia) => {
    if (id == null) return;
    try {
      const playbackId = await startStaffPlay(id, media.id);
      setMessage(`Sent staff play of media ${media.id} (playback ${playbackId ?? "?"})`);
    } catch (error) {
      setMessage(`Couldn't play media ${media.id}: ${String(error)}`);
    }
  };

  const fake = fakeStartTime();
  const awake = screen?.awakeHours;
  const today = dayOf(current);
  const playing = state.playing;

  return (
    <ScrollView testID="dev-panel" style={styles.panel} contentContainerStyle={styles.panelContent}>
      <Row label="Clock" value={`${time(current)} ${today}${fake ? ` (fake, from ${dayOf(fake)} ${time(fake)})` : ""}`} />
      <Row
        label="Screen"
        value={playing ? `${mode}: ${playing.item.kind} ${playing.item.title ?? playing.item.url} (${playing.source})` : mode}
      />
      <Row
        label="Awake"
        value={
          awake
            ? `${awake.wake} to ${awake.bed}, ${isNight(current, nightHoursFromAwake(awake)) ? "night" : "day"} now`
            : "no screen yet (default night 20:00 to 07:00)"
        }
      />
      <Row
        label="Carepartner"
        value={`${id ?? "none"} (${state.pairing}), socket ${loaded.connected == null ? "?" : loaded.connected ? "connected" : "not connected"}`}
      />
      <Row
        label="Screen fetch"
        value={`${fetchedAt ? `last good at ${time(fetchedAt)}` : "none yet"}, ${screen?.widgets.length ?? 0} widgets`}
      />

      <View style={styles.actions}>
        <SmallButton label="Refetch screen" onPress={() => refresh()} />
        <SmallButton label="Reload panel" onPress={() => load()} />
      </View>

      <Text style={styles.heading}>Schedule</Text>
      {screen?.schedule.length ? (
        screen.schedule.map((entry) => (
          <Text key={`${entry.scheduledItemId}@${entry.date}`} style={styles.text}>
            {entry.date === today ? "today" : entry.date} {entry.time} · item {entry.scheduledItemId} · {entry.item.kind}{" "}
            {entry.item.title ?? ""} · {playStatus(entry, current, loaded.played)}
          </Text>
        ))
      ) : (
        <Text style={styles.text}>nothing scheduled today or tomorrow</Text>
      )}

      <Text style={styles.heading}>Media (tap to staff-play)</Text>
      {message && <Text style={styles.message}>{message}</Text>}
      {loaded.media == null ? (
        <Text style={styles.text}>{id == null ? "not paired" : "couldn't load"}</Text>
      ) : loaded.media.length ? (
        loaded.media.map((media) => (
          <Pressable key={media.id} testID={`dev-media-${media.id}`} onPress={() => play(media)} style={styles.mediaRow}>
            <Text style={styles.text}>
              #{media.id} · {media.kind} · {media.name}
            </Text>
          </Pressable>
        ))
      ) : (
        <Text style={styles.text}>no media</Text>
      )}
    </ScrollView>
  );
}

function Row({ label, value }: { label: string; value: string }) {
  return (
    <Text style={styles.text}>
      <Text style={styles.label}>{label}: </Text>
      {value}
    </Text>
  );
}

function SmallButton({ label, onPress }: { label: string; onPress: () => void }) {
  return (
    <Pressable onPress={onPress} style={styles.smallButton}>
      <Text style={styles.smallButtonText}>{label}</Text>
    </Pressable>
  );
}

// Plain developer styling, deliberately unlike the resident's screens.
const styles = StyleSheet.create({
  // Below the Android status bar, which the app draws under.
  overlay: { position: "absolute", top: (StatusBar.currentHeight ?? 0) + 8, right: 12, alignItems: "flex-end", maxWidth: "70%", maxHeight: "95%" },
  buttons: { flexDirection: "row", gap: 8 },
  button: {
    width: 44,
    height: 44,
    borderRadius: 22,
    alignItems: "center",
    justifyContent: "center",
    backgroundColor: "rgba(0,0,0,0.6)",
  },
  buttonText: { color: "#fff", fontSize: 22 },
  panel: { marginTop: 8, backgroundColor: "rgba(20,20,20,0.92)", borderRadius: 8 },
  panelContent: { padding: 12, gap: 4 },
  heading: { color: "#9cf", fontSize: 15, fontWeight: "bold", marginTop: 8 },
  label: { color: "#9cf", fontWeight: "bold" },
  text: { color: "#eee", fontSize: 14, fontFamily: "monospace" },
  message: { color: "#fc6", fontSize: 14 },
  actions: { flexDirection: "row", gap: 8, marginTop: 6 },
  smallButton: { paddingHorizontal: 10, paddingVertical: 6, borderRadius: 6, backgroundColor: "#345" },
  smallButtonText: { color: "#fff", fontSize: 14 },
  mediaRow: { paddingVertical: 4 },
});
