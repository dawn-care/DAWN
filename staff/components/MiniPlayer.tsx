/**
 * The bar at the bottom of every screen that keeps this phone's plays within
 * reach after staff leave the Playing screen (plays are followed by
 * lib/activePlays).
 *
 * A play shows here while it's playing or waiting for the tablet (with Stop),
 * while it can be rated (tap to rate on its Playing screen), and for two
 * minutes after it didn't play or didn't finish (with a close button). A play
 * replaced by a newer one on the same tablet leaves the bar. The play whose
 * Playing screen is open isn't shown twice.
 *
 * With several, the bar shows how many, and a tap lists them all.
 */

import { useState } from 'react';
import { ActivityIndicator, Pressable, ScrollView, Text, View } from 'react-native';
import { router, useGlobalSearchParams, usePathname } from 'expo-router';
import { Ionicons } from '@expo/vector-icons';
import { useSafeAreaInsets } from 'react-native-safe-area-context';

import {
  canRate,
  endedAt,
  formatDuration,
  phaseOf,
  positionOf,
  useActivePlays,
  useNow,
  type ActivePlay,
} from '../lib/activePlays';
import { useThemeColors } from '../lib/theme';

/** How long a play that didn't play or didn't finish stays in the bar. */
const PROBLEM_SHOWN_MS = 2 * 60 * 1000;

const ICON: Record<ActivePlay['clip']['kind'], string> = {
  video: 'videocam',
  audio: 'musical-notes',
  photo: 'image',
};

function isShown(p: ActivePlay, now: number): boolean {
  if (p.replaced || p.dismissed) return false;
  const phase = phaseOf(p, now);
  if (phase === 'playing' || phase === 'waiting') return true;
  if (phase === 'missed' || phase === 'failed') return now - endedAt(p) < PROBLEM_SHOWN_MS;
  return p.rated === null && canRate(p, now);
}

const isActive = (p: ActivePlay, now: number) => ['playing', 'waiting'].includes(phaseOf(p, now));

const open = (p: ActivePlay) =>
  router.push(`/resident/${p.resident.id}/playing?playbackId=${p.playbackId}&clipId=${p.clip.id}`);

export default function MiniPlayer() {
  const { plays } = useActivePlays();
  const pathname = usePathname();
  const { playbackId } = useGlobalSearchParams<{ playbackId?: string }>();
  const insets = useSafeAreaInsets();
  const colors = useThemeColors();
  const [expanded, setExpanded] = useState(false);

  // Tick only while a play could still be in the bar: not closed, replaced, or
  // rated after it ended, and still inside its rating window
  const now = useNow(
    1000,
    plays.some((p) => !p.replaced && !p.dismissed && !p.ratingClosed && !(p.rated !== null && p.state?.ended))
  );
  const viewing = pathname.endsWith('/playing') ? Number(playbackId) : null;
  const shown = plays.filter((p) => p.playbackId !== viewing && isShown(p, now));
  if (shown.length === 0) return null;

  const playing = shown.filter((p) => isActive(p, now)).length;
  const toRate = shown.filter((p) => !isActive(p, now) && canRate(p, now)).length;
  const summary = [
    playing > 0 && `${playing} playing`,
    toRate > 0 && `${toRate} to rate`,
    shown.length - playing - toRate > 0 && `${shown.length - playing - toRate} didn't play`,
  ]
    .filter(Boolean)
    .join(' · ');

  return (
    <View
      className="border-t border-border bg-surface px-3 pt-2"
      style={{ paddingBottom: Math.max(insets.bottom, 8) }}
    >
      {shown.length === 1 ? (
        <Row play={shown[0]} now={now} />
      ) : (
        <>
          {expanded && (
            <ScrollView style={{ maxHeight: 280 }}>
              {shown.map((p) => (
                <Row key={p.playbackId} play={p} now={now} />
              ))}
            </ScrollView>
          )}
          <Pressable
            onPress={() => setExpanded(!expanded)}
            accessibilityRole="button"
            accessibilityLabel={`${summary}. ${expanded ? 'Hide' : 'Show'} the list`}
            className="flex-row items-center py-2 active:opacity-70"
          >
            <View className="mr-3 h-10 w-10 items-center justify-center rounded-full bg-page">
              <Ionicons name="albums" size={20} color={colors.accent} />
            </View>
            <View className="flex-1">
              <Text className="text-base font-semibold text-ink">{summary}</Text>
              <Text numberOfLines={1} className="text-sm text-ink-muted">
                {[...new Set(shown.map((p) => p.resident.name.split(' ')[0]))].join(', ')}
              </Text>
            </View>
            <Ionicons name={expanded ? 'chevron-down' : 'chevron-up'} size={22} color={colors.inkMuted} />
          </Pressable>
        </>
      )}
    </View>
  );
}

function Row({ play, now }: { play: ActivePlay; now: number }) {
  const { stop, dismiss } = useActivePlays();
  const colors = useThemeColors();
  const phase = phaseOf(play, now);
  const active = phase === 'playing' || phase === 'waiting';
  const problem = phase === 'missed' || phase === 'failed';
  const firstName = play.resident.name.split(' ')[0];
  const duration = play.clip.duration;
  const startedAt = play.state?.startedAt ?? null;
  const position = positionOf(play, now);

  let detail: string;
  if (play.stopNote && active) detail = play.stopNote;
  else if (phase === 'waiting') detail = 'Waiting for the tablet to connect';
  else if (phase === 'playing')
    detail = startedAt === null ? 'Starting…' : duration ? `${formatDuration(duration - position)} left` : 'Playing';
  else if (phase === 'missed') detail = "Didn't reach the tablet";
  else if (phase === 'failed') detail = "Didn't finish playing";
  else detail = 'How did it go? Tap to rate';

  return (
    <View className="mb-1">
      <View className="flex-row items-center py-2">
        <Pressable
          onPress={() => open(play)}
          accessibilityRole="button"
          accessibilityLabel={`${firstName}, ${play.clip.title}. ${detail}. Open`}
          className="flex-1 flex-row items-center active:opacity-70"
        >
          <View className="mr-3 h-10 w-10 items-center justify-center rounded-full bg-page">
            <Ionicons
              name={ICON[play.clip.kind] as never}
              size={20}
              color={phase === 'playing' ? colors.accent : colors.inkMuted}
            />
          </View>
          <View className="flex-1">
            <Text numberOfLines={1} className="text-base font-semibold text-ink">
              {firstName} · {play.clip.title}
            </Text>
            <Text
              numberOfLines={2}
              className={`text-sm ${!active && !problem ? 'font-semibold text-link' : 'text-ink-muted'}`}
            >
              {detail}
            </Text>
          </View>
          {!active && !problem && (
            <Ionicons name="chevron-forward" size={22} color={colors.inkMuted} style={{ marginLeft: 12 }} />
          )}
        </Pressable>

        {active ? (
          <Pressable
            onPress={() => stop(play.playbackId)}
            disabled={play.stopping}
            hitSlop={8}
            accessibilityRole="button"
            accessibilityLabel={`${phase === 'waiting' ? 'Cancel' : 'Stop'} ${play.clip.title}`}
            className="ml-3 h-11 w-11 items-center justify-center rounded-full border border-border active:opacity-70"
          >
            {play.stopping ? (
              <ActivityIndicator color={colors.ink} />
            ) : (
              <Ionicons name={phase === 'waiting' ? 'close' : 'square'} size={phase === 'waiting' ? 20 : 14} color={colors.ink} />
            )}
          </Pressable>
        ) : problem && (
          <Pressable
            onPress={() => dismiss(play.playbackId)}
            hitSlop={8}
            accessibilityRole="button"
            accessibilityLabel={`Dismiss ${play.clip.title}`}
            className="ml-3 h-11 w-11 items-center justify-center active:opacity-70"
          >
            <Ionicons name="close" size={20} color={colors.inkMuted} />
          </Pressable>
        )}
      </View>

      {/* How far along it is, once the tablet has it on screen */}
      {phase === 'playing' && duration && startedAt !== null ? (
        <View className="ml-[52px] h-1 overflow-hidden rounded-full bg-border">
          <View className="h-full rounded-full bg-accent" style={{ width: `${(position / duration) * 100}%` }} />
        </View>
      ) : null}
    </View>
  );
}
