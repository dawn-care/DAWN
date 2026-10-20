/**
 * S5 · Playing + rate it
 *
 * Play and Stop are real: the play command was sent from the clip picker, and
 * Stop asks the server to stop it on the tablet.
 *
 * What is NOT real yet:
 *  - Progress. The server can't yet report whether the tablet started or
 *    finished, so the bar is an estimate from the clip's length, and it is only
 *    shown when the server confirmed the tablet received the command.
 *  - Rating. It isn't saved anywhere; the screen behaves as if it were.
 *
 * Rating window: five minutes after the clip ends or is stopped, then the
 * controls disappear, since later ratings aren't accurate.
 */

import { useEffect, useRef, useState } from 'react';
import { ActivityIndicator, Pressable, ScrollView, Text, View } from 'react-native';
import { router, useLocalSearchParams } from 'expo-router';
import { Ionicons } from '@expo/vector-icons';

import { rateClip } from '../../../data/mock';
import { getClip, getResident, stopOnTablet } from '../../../data/server';
import type { Clip, Rating, Resident } from '../../../data/types';
import { logEvent } from '../../../lib/log';
import { useThemeColors, type ThemeColors } from '../../../lib/theme';

const RATING_WINDOW_MS = 5 * 60 * 1000;
/** How long the server waits for an offline tablet before the command expires. */
const DELIVERY_WINDOW_MS = 30 * 1000;

const formatDuration = (seconds: number) =>
  `${Math.floor(seconds / 60)}:${String(Math.floor(seconds % 60)).padStart(2, '0')}`;

const OPTIONS: { rating: Rating; label: string; icon: string; color: keyof ThemeColors }[] = [
  { rating: 'helped', label: 'Helped', icon: 'checkmark', color: 'statusGreenText' },
  { rating: 'no_change', label: 'No change', icon: 'remove', color: 'statusGrayText' },
  { rating: 'worse', label: 'Made it worse', icon: 'arrow-down', color: 'statusRedText' },
];

/**
 * - playing: sent to the tablet, estimated to be playing
 * - waiting: the tablet wasn't connected; it may still start if it reconnects
 * - ended: estimated to have finished
 * - stopped: staff stopped it
 * - missed: the tablet never received it in time
 */
type Phase = 'playing' | 'waiting' | 'ended' | 'stopped' | 'missed';

export default function PlayingScreen() {
  const { id, playbackId, clipId, delivered } = useLocalSearchParams<{
    id: string;
    playbackId: string;
    clipId: string;
    delivered: string;
  }>();
  const residentId = Number(id);
  const colors = useThemeColors();

  const [clip, setClip] = useState<Clip | null>(null);
  const [resident, setResident] = useState<Resident | null>(null);
  const [phase, setPhase] = useState<Phase>(delivered === '1' ? 'playing' : 'waiting');
  const [position, setPosition] = useState(0);
  const [stopping, setStopping] = useState(false);
  const [stopNote, setStopNote] = useState<string | null>(null);
  const [submitted, setSubmitted] = useState<Rating | null>(null);
  const [windowOpen, setWindowOpen] = useState(true);
  const windowTimer = useRef<ReturnType<typeof setTimeout> | null>(null);

  useEffect(() => {
    getClip(Number(clipId)).then((c) => setClip(c ?? null));
    getResident(residentId).then((r) => setResident(r ?? null));
  }, [clipId, residentId]);

  /** Estimated playback clock, from the clip's length. Not reported by the tablet. */
  useEffect(() => {
    if (phase !== 'playing' || !clip?.duration) return;
    const duration = clip.duration;
    const timer = setInterval(() => {
      setPosition((p) => {
        if (p + 1 >= duration) {
          setPhase('ended');
          return duration;
        }
        return p + 1;
      });
    }, 1000);
    return () => clearInterval(timer);
  }, [phase, clip]);

  /** An offline tablet has 30 seconds to reconnect before the command expires. */
  useEffect(() => {
    if (phase !== 'waiting') return;
    const timer = setTimeout(() => setPhase('missed'), DELIVERY_WINDOW_MS);
    return () => clearTimeout(timer);
  }, [phase]);

  /** The rating window opens once the clip is over. */
  useEffect(() => {
    if ((phase !== 'ended' && phase !== 'stopped') || !clip) return;
    windowTimer.current = setTimeout(() => {
      setWindowOpen(false);
      logEvent({ type: 'rating_window_expired', residentId, clipId: clip.id });
    }, RATING_WINDOW_MS);
    return () => {
      if (windowTimer.current) clearTimeout(windowTimer.current);
    };
  }, [phase, clip, residentId]);

  if (!clip || !resident) return <ActivityIndicator className="mt-10" />;

  const firstName = resident.name.split(' ')[0];

  const stop = async () => {
    setStopping(true);
    setStopNote(null);
    try {
      const result = await stopOnTablet(residentId, Number(playbackId));
      if (result === 'unreachable') {
        setStopNote("Couldn't reach the tablet to stop it. It may still be playing.");
        return;
      }
      if (result === 'finished') {
        setStopNote('It had already finished on the tablet.');
        setPhase('ended');
        return;
      }
      logEvent({ type: 'play_stopped', residentId, clipId: clip.id, atSecond: position });
      setPhase(phase === 'waiting' ? 'missed' : 'stopped');
    } catch {
      setStopNote("That didn't send. Check the connection and try again.");
    } finally {
      setStopping(false);
    }
  };

  const submitRating = async (rating: Rating) => {
    await rateClip(Number(playbackId), rating);
    logEvent({ type: 'clip_rated', residentId, clipId: clip.id, rating });
    setSubmitted(rating);
  };

  const duration = clip.duration;
  const progress = duration ? Math.min(100, (position / duration) * 100) : 0;
  const statusLabel = {
    playing: `Sent to ${firstName}'s tablet`,
    waiting: `${firstName}'s tablet isn't connected`,
    ended: 'Finished',
    stopped: 'Stopped',
    missed: "Didn't play",
  }[phase];

  return (
    <ScrollView className="flex-1 bg-page" contentContainerClassName="p-4 pb-10">
      <View className="rounded-2xl border border-border bg-surface p-4">
        <View className="mb-2 flex-row items-center">
          <View
            className={`mr-2 h-3 w-3 rounded-full ${
              phase === 'playing' ? 'bg-accent' : 'bg-status-gray-text'
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

        {(phase === 'playing' || phase === 'waiting') && (
          <Pressable
            onPress={stop}
            disabled={stopping}
            className="mt-4 flex-row items-center justify-center rounded-xl border border-border py-4 active:opacity-70"
          >
            {stopping ? (
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
        {phase === 'ended' && !stopNote && (
          <Pressable onPress={stop} disabled={stopping} className="mt-3 self-start py-1 active:opacity-70">
            <Text className="text-base font-semibold text-link">
              {stopping ? 'Stopping…' : 'Still playing? Stop it'}
            </Text>
          </Pressable>
        )}

        {stopNote && <Text className="mt-3 text-base text-ink-muted">{stopNote}</Text>}
      </View>

      {/* Rating. Not saved yet; only offered for a clip that actually reached the tablet. */}
      {phase !== 'waiting' && phase !== 'missed' &&
        (submitted ? (
          <View className="mt-6 rounded-2xl bg-status-green-bg p-4">
            <Text className="text-base font-semibold text-status-green-text">
              Thanks — recorded as {OPTIONS.find((o) => o.rating === submitted)?.label}.
            </Text>
          </View>
        ) : !windowOpen ? (
          <View className="mt-6 rounded-2xl bg-surface-raised p-4">
            <Text className="text-base text-ink-muted">
              The rating window has closed. Ratings are only collected for five minutes after a
              video plays, so they stay accurate.
            </Text>
          </View>
        ) : (
          <>
            <Text className="mt-6 text-3xl font-bold text-ink">How did it go?</Text>
            <Text className="mb-4 text-base text-ink-muted">
              One tap. This is how the team learns what works.
            </Text>
            {OPTIONS.map((option) => (
              <Pressable
                key={option.rating}
                onPress={() => submitRating(option.rating)}
                className="mb-3 flex-row items-center rounded-2xl border border-border bg-surface p-4 active:opacity-70"
              >
                <Ionicons name={option.icon as never} size={24} color={colors[option.color]} />
                <Text className="ml-3 text-lg font-semibold text-ink">{option.label}</Text>
              </Pressable>
            ))}
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
