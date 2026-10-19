/**
 * What last played on a resident's tablet: staff plays, scheduled plays and
 * taps alike. Reloads whenever the screen comes back into view, so it is
 * current after staff play something and come back.
 */

import { useCallback, useState } from 'react';
import { Text, View } from 'react-native';
import { useFocusEffect } from 'expo-router';
import { Ionicons } from '@expo/vector-icons';

import { getLastPlayed } from '../data/server';
import type { PlayedItem } from '../data/types';
import { formatTime } from '../lib/sleepingHours';
import { useThemeColors } from '../lib/theme';

const STATUS: Record<PlayedItem['status'], string> = {
  // The tablet hasn't reported an end, so don't claim it's still playing
  started: 'Started',
  finished: 'Played to the end',
  stopped: 'Stopped',
  failed: "Didn't finish",
};

const SOURCE: Record<PlayedItem['source'], string> = {
  staff: 'Sent by staff',
  scheduled: 'On the schedule',
  tap: 'Tapped on the tablet',
};

const ICON: Record<PlayedItem['kind'], string> = {
  video: 'videocam',
  audio: 'musical-notes',
  photo: 'image',
};

const startOfDay = (d: Date) => new Date(d.getFullYear(), d.getMonth(), d.getDate()).getTime();

/** "Today, 2:15 pm", "Yesterday, 9:00 am" or "Oct 14, 6:30 pm". */
export const formatPlayedAt = (iso: string, now: Date = new Date()): string => {
  const at = new Date(iso);
  const time = formatTime(`${at.getHours()}:${String(at.getMinutes()).padStart(2, '0')}`);
  const daysAgo = Math.round((startOfDay(now) - startOfDay(at)) / 86_400_000);
  const day =
    daysAgo === 0
      ? 'Today'
      : daysAgo === 1
        ? 'Yesterday'
        : at.toLocaleDateString(undefined, { month: 'short', day: 'numeric' });
  return `${day}, ${time}`;
};

type State = { phase: 'loading' } | { phase: 'failed' } | { phase: 'loaded'; played: PlayedItem | null };

export default function LastPlayed({ residentId }: { residentId: number }) {
  const colors = useThemeColors();
  const [state, setState] = useState<State>({ phase: 'loading' });

  useFocusEffect(
    useCallback(() => {
      let active = true;
      getLastPlayed(residentId)
        .then((played) => active && setState({ phase: 'loaded', played }))
        .catch(() => active && setState({ phase: 'failed' }));
      return () => {
        active = false;
      };
    }, [residentId])
  );

  const played = state.phase === 'loaded' ? state.played : null;

  return (
    <View className="mt-3 rounded-2xl bg-surface-raised p-4">
      <Text className="text-base font-semibold text-ink">Last played</Text>
      {state.phase === 'loading' && <Text className="text-base text-ink-muted">Checking…</Text>}
      {state.phase === 'failed' && (
        <Text className="text-base text-ink-muted">Couldn&apos;t check what last played.</Text>
      )}
      {state.phase === 'loaded' && !played && (
        <Text className="text-base text-ink-muted">Nothing has played yet.</Text>
      )}
      {played && (
        <View className="mt-1 flex-row items-start">
          <Ionicons
            name={ICON[played.kind] as never}
            size={20}
            color={colors.inkMuted}
            style={{ marginTop: 2 }}
          />
          <View className="ml-2 flex-1">
            <Text className="text-base text-ink">{played.title}</Text>
            <Text className="text-base text-ink-muted">
              {formatPlayedAt(played.startedAt)} · {STATUS[played.status]}
            </Text>
            <Text className="text-sm text-ink-muted">{SOURCE[played.source]}</Text>
          </View>
        </View>
      )}
    </View>
  );
}
