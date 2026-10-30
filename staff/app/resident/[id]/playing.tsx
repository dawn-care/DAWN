/**
 * S5 · Playing + rate it
 *
 * Play and Stop are real: the play command was sent from the clip picker, and
 * Stop asks the server to stop it on the tablet.
 *
 * The screen asks the server about the play every few seconds (getPlayState):
 *  - Progress is estimated: the time since the tablet put the clip on screen,
 *    up to the clip's length. It is only shown when the server confirmed the
 *    tablet received the command. When the tablet reports the end, the screen
 *    says so then (Finished, Stopped, even if another phone stopped it, or
 *    Didn't finish), whatever the estimate said.
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

import {
  getClip,
  getPlayState,
  getResident,
  rateClip,
  stopOnTablet,
  type PlayState,
} from '../../../data/server';
import type { Clip, Rating, Resident } from '../../../data/types';
import { logEvent } from '../../../lib/log';
import { useThemeColors, type ThemeColors } from '../../../lib/theme';

/** How long the server waits for an offline tablet before the command expires. */
const DELIVERY_WINDOW_MS = 30 * 1000;
/** How often to ask the server what the tablet has done with the play. */
const POLL_MS = 3 * 1000;
/** How long a clip is on screen before staff can rate it, if it hasn't ended by then. */
const RATE_AFTER_MS = 15 * 1000;

const formatDuration = (seconds: number) =>
  `${Math.floor(seconds / 60)}:${String(Math.floor(seconds % 60)).padStart(2, '0')}`;

/** Best to worst. The server weighs the ends more, and worse more than helped. */
const OPTIONS: { rating: Rating; label: string; icon: string; color: keyof ThemeColors }[] = [
  { rating: 'helped_a_lot', label: 'Helped a lot', icon: 'checkmark-done', color: 'statusGreenText' },
  { rating: 'helped', label: 'Helped a little', icon: 'checkmark', color: 'statusGreenText' },
  { rating: 'no_change', label: 'No change', icon: 'remove', color: 'statusGrayText' },
  { rating: 'worse', label: 'Made it a little worse', icon: 'arrow-down', color: 'statusRedText' },
  { rating: 'much_worse', label: 'Made it much worse', icon: 'arrow-down-circle', color: 'statusRedText' },
];

/**
 * - playing: sent to the tablet, estimated to be playing
 * - waiting: the tablet wasn't connected; it may still start if it reconnects
 * - ended: finished, as the tablet reported or as estimated
 * - stopped: staff stopped it
 * - missed: the tablet never received it in time
 * - failed: the tablet showed it but couldn't play it to the end
 */
type Phase = 'playing' | 'waiting' | 'ended' | 'stopped' | 'missed' | 'failed';

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
  const [rating, setRating] = useState<Rating | null>(null);
  const [ratingNote, setRatingNote] = useState<string | null>(null);
  const [play, setPlay] = useState<PlayState | null>(null);
  const [ratingClosed, setRatingClosed] = useState(false);
  const [onScreenLongEnough, setOnScreenLongEnough] = useState(false);

  useEffect(() => {
    getClip(Number(clipId)).then((c) => setClip(c ?? null));
    getResident(residentId).then((r) => setResident(r ?? null));
  }, [clipId, residentId]);

  /**
   * Estimated playback clock: the time since the tablet put the clip on screen
   * (as the server reports it), capped at the clip's length. It stays at 0:00
   * until the tablet has started, so a slow load doesn't eat into it.
   */
  const startedAt = play?.startedAt ?? null;
  useEffect(() => {
    if (phase !== 'playing' || !clip?.duration || startedAt === null) return;
    const duration = clip.duration;
    const tick = () => {
      const elapsed = Math.floor((Date.now() - startedAt) / 1000);
      if (elapsed >= duration) {
        setPosition(duration);
        setPhase('ended');
      } else {
        setPosition(Math.max(0, elapsed));
      }
    };
    tick();
    const timer = setInterval(tick, 1000);
    return () => clearInterval(timer);
  }, [phase, clip, startedAt]);

  /** An offline tablet has 30 seconds to reconnect before the command expires. */
  useEffect(() => {
    if (phase !== 'waiting') return;
    const timer = setTimeout(() => setPhase('missed'), DELIVERY_WINDOW_MS);
    return () => clearTimeout(timer);
  }, [phase]);

  /** What the tablet has done with the play, until there is nothing left to learn. */
  useEffect(() => {
    if (phase === 'missed' || ratingClosed || submitted) return;
    let active = true;
    const check = () =>
      getPlayState(residentId, Number(playbackId))
        .then((state) => active && setPlay(state))
        // A missed check is retried on the next one
        .catch(() => undefined);
    check();
    const timer = setInterval(check, POLL_MS);
    return () => {
      active = false;
      clearInterval(timer);
    };
  }, [residentId, playbackId, phase, ratingClosed, submitted]);

  /** The tablet's own end beats the estimate, including when it couldn't finish. */
  useEffect(() => {
    if (play?.failed && (phase === 'playing' || phase === 'ended')) {
      setPhase('failed');
    } else if (play?.stopped && (phase === 'playing' || phase === 'ended')) {
      // Stopped from another phone, or the tablet confirmed this phone's Stop
      setPhase('stopped');
    } else if (play?.ended && phase === 'playing') {
      setPhase('ended');
      if (clip?.duration) setPosition(clip.duration);
    }
  }, [play, phase, clip]);

  /** Rating appears once the clip has been on the tablet's screen for RATE_AFTER_MS. */
  useEffect(() => {
    if (startedAt === null) return;
    const timer = setTimeout(
      () => setOnScreenLongEnough(true),
      Math.max(0, startedAt + RATE_AFTER_MS - Date.now())
    );
    return () => clearTimeout(timer);
  }, [startedAt]);

  /** Rating disappears when the server would stop accepting it. */
  const closesAt = play?.ratingClosesAt ?? null;
  useEffect(() => {
    if (closesAt === null || !clip) return;
    const timer = setTimeout(() => {
      setRatingClosed(true);
      logEvent({ type: 'rating_window_expired', residentId, clipId: clip.id });
    }, Math.max(0, closesAt - Date.now()));
    return () => clearTimeout(timer);
  }, [closesAt, clip, residentId]);

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

  const submitRating = async (chosen: Rating) => {
    setRating(chosen);
    setRatingNote(null);
    try {
      // Closed only when tapped right as the window ends; then rating simply goes away
      if ((await rateClip(residentId, Number(playbackId), chosen)) === 'closed') {
        setRatingClosed(true);
        return;
      }
    } catch {
      setRatingNote("That didn't save. Check the connection and try again.");
      return;
    } finally {
      setRating(null);
    }
    logEvent({ type: 'clip_rated', residentId, clipId: clip.id, rating: chosen });
    setSubmitted(chosen);
  };

  const duration = clip.duration;
  const progress = duration ? Math.min(100, (position / duration) * 100) : 0;
  // It reached the screen (the server gave it a closing time), and it has been on
  // long enough or is already over. A play the tablet couldn't finish isn't rated:
  // that says something about the tablet, not the clip.
  const canRate =
    closesAt !== null &&
    !ratingClosed &&
    phase !== 'failed' &&
    !play?.failed &&
    (onScreenLongEnough || play?.ended || phase === 'ended' || phase === 'stopped');
  const statusLabel = {
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

      {/* Rating: only offered while the server would accept it */}
      {submitted ? (
        <View className="mt-6 flex-row items-center justify-between rounded-2xl bg-status-green-bg p-4">
          <Text className="flex-1 text-base font-semibold text-status-green-text">
            Thanks — recorded as {OPTIONS.find((o) => o.rating === submitted)?.label}.
          </Text>
          {/* A mis-tap can be fixed while the window is open; the new rating replaces this one */}
          {!ratingClosed && (
            <Pressable onPress={() => setSubmitted(null)} hitSlop={12} className="ml-3 active:opacity-70">
              <Text className="text-base font-semibold text-link underline">Change</Text>
            </Pressable>
          )}
        </View>
      ) : (
        canRate && (
          <>
            <Text className="mt-6 text-3xl font-bold text-ink">How did it go?</Text>
            <Text className="mb-4 text-base text-ink-muted">
              One tap. This is how the team learns what works.
            </Text>
            {OPTIONS.map((option) => (
              <Pressable
                key={option.rating}
                onPress={() => submitRating(option.rating)}
                disabled={rating !== null}
                className={`mb-3 flex-row items-center rounded-2xl border border-border bg-surface p-4 active:opacity-70 ${
                  rating !== null && rating !== option.rating ? 'opacity-60' : ''
                }`}
              >
                {rating === option.rating ? (
                  <ActivityIndicator color={colors[option.color]} />
                ) : (
                  <Ionicons name={option.icon as never} size={24} color={colors[option.color]} />
                )}
                <Text className="ml-3 text-lg font-semibold text-ink">{option.label}</Text>
              </Pressable>
            ))}
            {ratingNote && <Text className="text-base text-ink-muted">{ratingNote}</Text>}
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
