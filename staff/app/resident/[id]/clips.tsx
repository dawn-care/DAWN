/**
 * S4 · Clip picker
 *
 * Lists the resident's real media (videos, audio and photos) from the server and plays
 * one on their tablet, newest first. A photo shows on the tablet for 30 seconds. Coming from a family folder, it shows the clips the
 * caregiver linked to that person; coming from a situation, the clips the caregiver
 * put under it. Clips the caregiver tagged with neither never show here.
 *
 * Playing is blocked during the resident's sleeping hours. Staff can still look
 * through the list; only the Play button is disabled.
 */

import { useCallback, useEffect, useState } from 'react';
import { ActivityIndicator, Image, Pressable, ScrollView, Text, View } from 'react-native';
import { router, useLocalSearchParams } from 'expo-router';
import { Ionicons } from '@expo/vector-icons';

import LoadError from '../../../components/LoadError';
import { SITUATIONS, getSleepingHours } from '../../../data/mock';
import { getClips, getFamilyFolders, getResident, playOnTablet } from '../../../data/server';
import type { Clip, Resident, SleepingHours } from '../../../data/types';
import { formatTime, isAsleep } from '../../../lib/sleepingHours';
import { logEvent } from '../../../lib/log';
import { useThemeColors } from '../../../lib/theme';

const formatDuration = (seconds: number | null) =>
  seconds === null ? '' : `${Math.floor(seconds / 60)}:${String(seconds % 60).padStart(2, '0')}`;

const KIND_LABELS: Record<Clip['kind'], string> = { video: 'Video', audio: 'Audio', photo: 'Photo' };

const formatAdded = (iso: string) =>
  new Date(iso).toLocaleDateString(undefined, { month: 'short', day: 'numeric' });

export default function ClipsScreen() {
  const { id, folderId, situation } = useLocalSearchParams<{
    id: string;
    folderId?: string;
    situation?: string;
  }>();
  const residentId = Number(id);
  const colors = useThemeColors();

  const [resident, setResident] = useState<Resident | null>(null);
  const [hours, setHours] = useState<SleepingHours | null>(null);
  const [clips, setClips] = useState<Clip[] | null>(null);
  const [heading, setHeading] = useState<string | null>(null);
  const [failed, setFailed] = useState(false);
  const [playingId, setPlayingId] = useState<number | null>(null);
  const [playError, setPlayError] = useState(false);

  const load = useCallback(() => {
    setFailed(false);
    setClips(null);
    const familyMemberId = folderId ? Number(folderId) : undefined;
    Promise.all([
      getResident(residentId),
      getSleepingHours(residentId),
      getClips(residentId, { familyMemberId, situation }),
    ])
      .then(([r, h, c]) => {
        if (!r) throw new Error('resident not found');
        setResident(r);
        setHours(h);
        setClips(c);
      })
      .catch(() => setFailed(true));

    // The heading says who they're asking for, or the situation
    if (familyMemberId) {
      getFamilyFolders(residentId)
        .then((folders) => {
          const folder = folders.find((f) => f.id === familyMemberId);
          setHeading(folder ? `Asking for ${folder.name}` : null);
        })
        .catch(() => setHeading(null));
    } else {
      setHeading(SITUATIONS.find((s) => s.id === situation)?.label ?? null);
    }
  }, [residentId, folderId, situation]);

  useEffect(load, [load]);

  if (failed) {
    return (
      <View className="flex-1 bg-page p-4">
        <LoadError onRetry={load} />
      </View>
    );
  }
  if (!resident || !hours || clips === null) return <ActivityIndicator className="mt-10" />;

  const asleep = isAsleep(hours);
  const firstName = resident.name.split(' ')[0];

  const play = async (clip: Clip) => {
    if (asleep) {
      logEvent({ type: 'play_blocked_sleeping_hours', residentId, clipId: clip.id });
      return;
    }
    setPlayError(false);
    setPlayingId(clip.id);
    logEvent({ type: 'play_requested', residentId, clipId: clip.id });
    try {
      const { playback, delivered } = await playOnTablet(residentId, clip.id);
      router.push(
        `/resident/${residentId}/playing?playbackId=${playback.id}&clipId=${clip.id}&delivered=${delivered ? 1 : 0}`
      );
    } catch {
      setPlayError(true);
    } finally {
      setPlayingId(null);
    }
  };

  return (
    <ScrollView className="flex-1 bg-page" contentContainerClassName="p-4 pb-10">
      {heading && <Text className="text-base text-ink-muted">{heading}</Text>}
      <Text className="mb-5 text-3xl font-bold text-ink">
        {clips.length === 0 ? (folderId || situation ? 'Nothing for this yet' : 'No media yet') : `${firstName}'s media`}
      </Text>

      {asleep && (
        <View className="mb-5 rounded-2xl bg-status-purple-bg p-4">
          <Text className="text-base text-status-purple-text">
            <Text className="font-semibold">Sleeping hours. </Text>
            Playing is paused until {formatTime(hours.end)}.
          </Text>
        </View>
      )}

      {playError && (
        <View className="mb-5 rounded-2xl bg-status-red-bg p-4">
          <Text className="text-base text-status-red-text">
            That didn&apos;t send. Check the connection and try again.
          </Text>
        </View>
      )}

      {clips.map((clip) => {
        const sending = playingId === clip.id;
        const disabled = asleep || playingId !== null;
        return (
          <View
            key={clip.id}
            className="mb-4 overflow-hidden rounded-2xl border border-border bg-surface"
          >
            <View className="h-32 items-center justify-center bg-nav">
              {clip.kind === 'photo' ? (
                <Image
                  source={{ uri: clip.fileUrl }}
                  accessibilityLabel={clip.title}
                  resizeMode="cover"
                  style={{ width: '100%', height: '100%' }}
                />
              ) : (
                <Ionicons
                  name={clip.kind === 'video' ? 'videocam' : 'musical-notes'}
                  size={44}
                  color={colors.onNav}
                />
              )}
              {clip.kind !== 'photo' && clip.duration !== null && (
                <Text className="absolute bottom-3 right-4 text-base font-semibold text-on-nav">
                  {formatDuration(clip.duration)}
                </Text>
              )}
            </View>

            <View className="p-4">
              <Text className="text-xl font-bold text-ink">{clip.title}</Text>
              <Text className="mt-1 text-base text-ink-muted">
                {KIND_LABELS[clip.kind]} · Added {formatAdded(clip.addedAt)}
              </Text>

              <Pressable
                onPress={() => play(clip)}
                disabled={disabled}
                className={`mt-4 flex-row items-center justify-center rounded-xl py-4 ${
                  asleep ? 'bg-status-gray-bg' : 'bg-accent active:opacity-80'
                } ${disabled && !asleep ? 'opacity-60' : ''}`}
              >
                {sending ? (
                  <ActivityIndicator color={colors.onAccent} />
                ) : (
                  <Ionicons name="play" size={20} color={asleep ? colors.statusGrayText : colors.onAccent} />
                )}
                <Text
                  className={`ml-2 text-lg font-semibold ${
                    asleep ? 'text-status-gray-text' : 'text-on-accent'
                  }`}
                >
                  {asleep
                    ? 'Playing paused'
                    : sending
                      ? 'Sending…'
                      : `${clip.kind === 'photo' ? 'Show' : 'Play'} on ${firstName}'s tablet`}
                </Text>
              </Pressable>
            </View>
          </View>
        );
      })}

      <Pressable
        onPress={() => router.push(`/resident/${residentId}/request?situation=${situation ?? ''}`)}
        className="mt-2 items-center py-3 active:opacity-70"
      >
        <Text className="text-lg font-semibold text-link underline">
          {clips.length === 0 ? 'Ask the caregiver for a video' : 'None fit? Ask the caregiver for a video'}
        </Text>
      </Pressable>
    </ScrollView>
  );
}
