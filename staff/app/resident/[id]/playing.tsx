/**
 * S5 · Playing + rate it
 *
 * Play and Stop are real: the play command was sent from the clip picker, and
 * Stop asks the server to stop it on the tablet.
 *
 * The play itself is followed by lib/activePlays for the whole app, so leaving
 * this screen doesn't lose it: the mini player (components/MiniPlayer) keeps
 * Stop and the rating within reach. This screen shows the same play in full.
 *
 * What it shows comes from the server's answers about the play:
 *  - Progress is estimated: the time since the tablet put the clip on screen,
 *    up to the clip's length. When the tablet reports the end, the screen says
 *    so then (Finished, Stopped, even if another phone stopped it, or Didn't
 *    finish), whatever the estimate said.
 *  - Rating ("How did it go?") appears only while the server would accept it:
 *    once the clip has been on the tablet's screen for 15 seconds, or has ended
 *    or been stopped if that's sooner, until five minutes after it ends
 *    (rating_closes_at). Then it disappears, since later ratings aren't
 *    accurate. Nothing is offered that would be refused. 15 seconds is enough to
 *    see a first reaction without making staff wait out a long clip.
 *
 * Ratings are saved on the playback and rank the media list for every staff
 * member (best first). Nobody is shown the scores. One tap replaces the buttons
 * with a thanks, so nothing is overwritten by accident; its Change link brings
 * them back until the window closes, and the new rating replaces the old one.
 */

import { useEffect, useState } from 'react';
import { ActivityIndicator, Pressable, ScrollView, Text, View } from 'react-native';
import { router, useLocalSearchParams } from 'expo-router';
import { Ionicons } from '@expo/vector-icons';

import LoadError from '../../../components/LoadError';
import type { Rating } from '../../../data/types';
import {
  canRate,
  formatDuration,
  phaseOf,
  positionOf,
  useActivePlays,
  useNow,
} from '../../../lib/activePlays';
import { useThemeColors, type ThemeColors } from '../../../lib/theme';

/** Best to worst. The server weighs the ends more, and worse more than helped. */
const RATING_OPTIONS: { rating: Rating; label: string; icon: string; color: keyof ThemeColors }[] = [
  { rating: 'helped_a_lot', label: 'Helped a lot', icon: 'checkmark-done', color: 'statusGreenText' },
  { rating: 'helped', label: 'Helped a little', icon: 'checkmark', color: 'statusGreenText' },
  { rating: 'no_change', label: 'No change', icon: 'remove', color: 'statusGrayText' },
  { rating: 'worse', label: 'Made it a little worse', icon: 'arrow-down', color: 'statusRedText' },
  { rating: 'much_worse', label: 'Made it much worse', icon: 'arrow-down-circle', color: 'statusRedText' },
];

export default function PlayingScreen() {
  const { id, playbackId, clipId } = useLocalSearchParams<{
    id: string;
    playbackId: string;
    clipId: string;
  }>();
  const residentId = Number(id);
  const colors = useThemeColors();
  const { plays, stop, rate, changeRating, adopt } = useActivePlays();
  const play = plays.find((p) => p.playbackId === Number(playbackId));
  const now = useNow();
  const [lost, setLost] = useState(false);

  // A play this phone doesn't remember (the page was reloaded): follow it again
  useEffect(() => {
    if (play || lost) return;
    adopt(residentId, Number(playbackId), Number(clipId)).catch(() => setLost(true));
  }, [play, lost, adopt, residentId, playbackId, clipId]);

  if (lost) {
    return (
      <View className="flex-1 bg-page p-4">
        <LoadError onRetry={() => setLost(false)} />
      </View>
    );
  }
  if (!play) return <ActivityIndicator className="mt-10" />;

  const { clip, resident } = play;
  const firstName = resident.name.split(' ')[0];
  const phase = phaseOf(play, now);
  const position = positionOf(play, now);
  const duration = clip.duration;
  const progress = duration ? Math.min(100, (position / duration) * 100) : 0;
  const statusLabel = play.replaced
    ? 'Replaced by another clip'
    : {
        playing: `Sent to ${firstName}'s tablet`,
        waiting: `${firstName}'s tablet isn't connected`,
        ended: 'Finished',
        stopped: 'Stopped',
        missed: "Didn't play",
        failed: "Didn't finish",
      }[phase];

  return (
    <ScrollView className="flex-1 bg-page" contentContainerClassName="p-4 pb-10">
      <View className="rounded-2xl border border-border bg-surface p-4">
        <View className="mb-2 flex-row items-center">
          <View
            className={`mr-2 h-3 w-3 rounded-full ${
              phase === 'playing' && !play.replaced ? 'bg-accent' : 'bg-status-gray-text'
            }`}
          />
          <Text className="text-base font-semibold text-link">{statusLabel}</Text>
        </View>

        <Text className="text-xl font-bold text-ink">{clip.title}</Text>

        {phase === 'waiting' && (
          <Text className="mt-2 text-base text-ink-muted">
            It will start if the tablet reconnects within 30 seconds.
          </Text>
        )}
        {phase === 'missed' && (
          <Text className="mt-2 text-base text-ink-muted">
            It didn&apos;t reach the tablet, so nothing played.
          </Text>
        )}
        {phase === 'failed' && (
          <Text className="mt-2 text-base text-ink-muted">
            It started on the tablet but couldn&apos;t keep playing. You can try again.
          </Text>
        )}

        {/* Only shown once the tablet is known to have received the command */}
        {(phase === 'playing' || phase === 'ended' || phase === 'stopped') && duration ? (
          <>
            <View className="mt-4 h-2 overflow-hidden rounded-full bg-border">
              <View className="h-full rounded-full bg-accent" style={{ width: `${progress}%` }} />
            </View>
            <View className="mt-2 flex-row justify-between">
              <Text className="text-base text-ink-muted">{formatDuration(position)}</Text>
              <Text className="text-base text-ink-muted">{formatDuration(duration)}</Text>
            </View>
            <Text className="mt-1 text-sm text-ink-muted">
              {clip.kind === 'photo' ? 'Photos show for 30 seconds' : 'Estimated from the clip’s length'}
            </Text>
          </>
        ) : null}

        {(phase === 'playing' || phase === 'waiting') && !play.replaced && (
          <Pressable
            onPress={() => stop(play.playbackId)}
            disabled={play.stopping}
            className="mt-4 flex-row items-center justify-center rounded-xl border border-border py-4 active:opacity-70"
          >
            {play.stopping ? (
              <ActivityIndicator color={colors.ink} />
            ) : (
              <Ionicons name="square" size={16} color={colors.ink} />
            )}
            <Text className="ml-2 text-lg font-semibold text-ink">
              {phase === 'waiting' ? 'Cancel' : 'Stop'}
            </Text>
          </Pressable>
        )}

        {/* The estimate can end before the tablet does (a slow start, buffering), so
            stopping stays possible after the bar fills. */}
        {phase === 'ended' && !play.state?.ended && !play.stopNote && !play.replaced && (
          <Pressable
            onPress={() => stop(play.playbackId)}
            disabled={play.stopping}
            className="mt-3 self-start py-1 active:opacity-70"
          >
            <Text className="text-base font-semibold text-link">
              {play.stopping ? 'Stopping…' : 'Still playing? Stop it'}
            </Text>
          </Pressable>
        )}

        {play.stopNote && <Text className="mt-3 text-base text-ink-muted">{play.stopNote}</Text>}
      </View>

      {/* Rating: only offered while the server would accept it */}
      {play.rated ? (
        <View className="mt-6 flex-row items-center justify-between rounded-2xl bg-status-green-bg p-4">
          <Text className="flex-1 text-base font-semibold text-status-green-text">
            Thanks — recorded as {RATING_OPTIONS.find((o) => o.rating === play.rated)?.label}.
          </Text>
          {/* A mis-tap can be fixed while the window is open; the new rating replaces this one */}
          {canRate(play, now) && (
            <Pressable onPress={() => changeRating(play.playbackId)} hitSlop={12} className="ml-3 active:opacity-70">
              <Text className="text-base font-semibold text-link underline">Change</Text>
            </Pressable>
          )}
        </View>
      ) : (
        canRate(play, now) && (
          <>
            <Text className="mt-6 text-3xl font-bold text-ink">How did it go?</Text>
            <Text className="mb-4 text-base text-ink-muted">
              One tap. This is how the team learns what works.
            </Text>
            {RATING_OPTIONS.map((option) => (
              <Pressable
                key={option.rating}
                onPress={() => rate(play.playbackId, option.rating)}
                disabled={play.ratingPending !== null}
                className={`mb-3 flex-row items-center rounded-2xl border border-border bg-surface p-4 active:opacity-70 ${
                  play.ratingPending !== null && play.ratingPending !== option.rating ? 'opacity-60' : ''
                }`}
              >
                {play.ratingPending === option.rating ? (
                  <ActivityIndicator color={colors[option.color]} />
                ) : (
                  <Ionicons name={option.icon as never} size={24} color={colors[option.color]} />
                )}
                <Text className="ml-3 text-lg font-semibold text-ink">{option.label}</Text>
              </Pressable>
            ))}
            {play.ratingNote && <Text className="text-base text-ink-muted">{play.ratingNote}</Text>}
          </>
        ))}

      <Pressable
        onPress={() => router.dismissTo(`/resident/${residentId}`)}
        className="mt-3 items-center py-3 active:opacity-70"
      >
        <Text className="text-lg font-semibold text-link">Back to resident</Text>
      </Pressable>
    </ScrollView>
  );
}
