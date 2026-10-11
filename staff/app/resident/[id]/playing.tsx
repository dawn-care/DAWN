/**
 * S5 · Playing + rate it
 *
 * Playback is simulated with a local timer. Once the WebSocket channel exists,
 * the progress shown here comes from the tablet reporting its real position and
 * `stop` sends a command instead of clearing an interval.
 *
 * Rating window: staff can rate for five minutes after playback finishes, then
 * the controls disappear. Ratings given later are not trustworthy, and a stale
 * rating is worse than none for the data this feeds.
 */

import { useEffect, useRef, useState } from 'react';
import { ActivityIndicator, Pressable, ScrollView, Text, View } from 'react-native';
import { router, useLocalSearchParams } from 'expo-router';
import { Ionicons } from '@expo/vector-icons';

import { getClip, rateClip } from '../../../data/mock';
import type { Clip, Rating } from '../../../data/types';
import { logEvent } from '../../../lib/log';

const RATING_WINDOW_MS = 5 * 60 * 1000;

const formatDuration = (seconds: number) =>
  `${Math.floor(seconds / 60)}:${String(Math.floor(seconds % 60)).padStart(2, '0')}`;

const OPTIONS: { rating: Rating; label: string; icon: string; color: string }[] = [
  { rating: 'helped', label: 'Helped', icon: 'checkmark', color: 'rgb(31 61 44)' },
  { rating: 'no_change', label: 'No change', icon: 'remove', color: 'rgb(91 96 145)' },
  { rating: 'worse', label: 'Made it worse', icon: 'arrow-down', color: 'rgb(74 34 38)' },
];

export default function PlayingScreen() {
  const { id, clipId } = useLocalSearchParams<{ id: string; clipId: string }>();
  const residentId = Number(id);

  const [clip, setClip] = useState<Clip | null>(null);
  const [position, setPosition] = useState(0);
  const [playing, setPlaying] = useState(true);
  const [submitted, setSubmitted] = useState<Rating | null>(null);
  const [windowOpen, setWindowOpen] = useState(true);

  const windowTimer = useRef<ReturnType<typeof setTimeout> | null>(null);

  useEffect(() => {
    getClip(Number(clipId)).then((c) => setClip(c ?? null));
  }, [clipId]);

  /** Simulated playback clock — replaced by tablet-reported position later. */
  useEffect(() => {
    if (!clip || !playing) return;

    const timer = setInterval(() => {
      setPosition((p) => {
        if (p + 1 >= clip.duration) {
          setPlaying(false);
          return clip.duration;
        }
        return p + 1;
      });
    }, 1000);

    return () => clearInterval(timer);
  }, [clip, playing]);

  /** Start the five-minute rating window once playback stops. */
  useEffect(() => {
    if (playing || !clip) return;

    windowTimer.current = setTimeout(() => {
      setWindowOpen(false);
      logEvent({ type: 'rating_window_expired', residentId, clipId: clip.id });
    }, RATING_WINDOW_MS);

    return () => {
      if (windowTimer.current) clearTimeout(windowTimer.current);
    };
  }, [playing, clip, residentId]);

  if (!clip) return <ActivityIndicator className="mt-10" />;

  const stop = () => {
    setPlaying(false);
    logEvent({
      type: 'play_stopped',
      residentId,
      clipId: clip.id,
      atSecond: position,
    });
  };

  const submitRating = async (rating: Rating) => {
    await rateClip(clip.id, rating);
    logEvent({ type: 'clip_rated', residentId, clipId: clip.id, rating });
    setSubmitted(rating);
  };

  const progress = Math.min(100, (position / clip.duration) * 100);

  return (
    <ScrollView className="flex-1 bg-page" contentContainerClassName="p-4 pb-10">
      {/* Playback card */}
      <View className="rounded-2xl border border-border bg-surface p-4">
        <View className="mb-2 flex-row items-center">
          <View
            className={`mr-2 h-3 w-3 rounded-full ${
              playing ? 'bg-accent' : 'bg-status-gray-text'
            }`}
          />
          <Text className="text-base font-semibold text-link">
            {playing ? 'Playing on the tablet now' : 'Finished'}
          </Text>
        </View>

        <Text className="text-xl font-bold text-ink">
          &quot;{clip.title}&quot;
        </Text>

        <View className="mt-4 h-2 overflow-hidden rounded-full bg-border">
          <View
            className="h-full rounded-full bg-accent"
            style={{ width: `${progress}%` }}
          />
        </View>

        <View className="mt-2 flex-row justify-between">
          <Text className="text-base text-ink-muted">
            {formatDuration(position)}
          </Text>
          <Text className="text-base text-ink-muted">
            {formatDuration(clip.duration)}
          </Text>
        </View>

        {playing && (
          <Pressable
            onPress={stop}
            className="mt-4 flex-row items-center justify-center rounded-xl border border-border py-4 active:opacity-70"
          >
            <Ionicons name="square" size={16} color="rgb(30 33 64)" />
            <Text className="ml-2 text-lg font-semibold text-ink">
              Stop video
            </Text>
          </Pressable>
        )}
      </View>

      {/* Rating */}
      {submitted ? (
        <View className="mt-6 rounded-2xl bg-status-green-bg p-4">
          <Text className="text-base font-semibold text-status-green-text">
            Thanks — recorded as{' '}
            {OPTIONS.find((o) => o.rating === submitted)?.label}.
          </Text>
        </View>
      ) : !windowOpen ? (
        <View className="mt-6 rounded-2xl bg-surface-raised p-4">
          <Text className="text-base text-ink-muted">
            The rating window has closed. Ratings are only collected for five
            minutes after a video plays, so they stay accurate.
          </Text>
        </View>
      ) : (
        <>
          <Text className="mt-6 text-3xl font-bold text-ink">
            How did it go?
          </Text>
          <Text className="mb-4 text-base text-ink-muted">
            One tap. This is how the team learns what works.
          </Text>

          {OPTIONS.map((option) => (
            <Pressable
              key={option.rating}
              onPress={() => submitRating(option.rating)}
              className="mb-3 flex-row items-center rounded-2xl border border-border bg-surface p-4 active:opacity-70"
            >
              <Ionicons name={option.icon as never} size={24} color={option.color} />
              <Text className="ml-3 text-lg font-semibold text-ink">
                {option.label}
              </Text>
            </Pressable>
          ))}
        </>
      )}

      <Pressable
        onPress={() => router.dismissTo(`/resident/${residentId}`)}
        className="mt-3 items-center py-3 active:opacity-70"
      >
        <Text className="text-lg font-semibold text-link">
          Back to resident
        </Text>
      </Pressable>
    </ScrollView>
  );
}
