/**
 * S4 · Clip list  (and the top half of S6 when nothing family-made fits)
 *
 * Reached either from a family folder ("asking for someone") or straight from a
 * situation. Best-match ranking is out of the MVP, so clips are listed rather
 * than ranked — the most-used clip is simply shown first.
 *
 * Playing is blocked during the resident's sleeping hours. Staff can still read
 * the list; only the play action is disabled.
 */

import { useEffect, useState } from 'react';
import { ActivityIndicator, Pressable, ScrollView, Text, View } from 'react-native';
import { router, useLocalSearchParams } from 'expo-router';
import { Ionicons } from '@expo/vector-icons';

import {
  getClipsForFamilyFolder,
  getClipsForSituation,
  getFamilyFolder,
  getResident,
} from '../../../data/mock';
import type { Clip, Resident, SituationId } from '../../../data/types';
import { describeWindow, isAsleep } from '../../../lib/sleepingHours';
import { logEvent } from '../../../lib/log';

const formatDuration = (seconds: number) =>
  `${Math.floor(seconds / 60)}:${String(seconds % 60).padStart(2, '0')}`;

export default function ClipsScreen() {
  const { id, folderId, situation } = useLocalSearchParams<{
    id: string;
    folderId?: string;
    situation?: string;
  }>();
  const residentId = Number(id);

  const [resident, setResident] = useState<Resident | null>(null);
  const [clips, setClips] = useState<Clip[] | null>(null);
  const [fromName, setFromName] = useState<string | null>(null);

  useEffect(() => {
    getResident(residentId).then((r) => setResident(r ?? null));

    if (folderId) {
      getFamilyFolder(Number(folderId)).then((f) =>
        setFromName(f?.name ?? null)
      );
      getClipsForFamilyFolder(Number(folderId)).then(setClips);
    } else if (situation) {
      getClipsForSituation(residentId, situation as SituationId).then(setClips);
    } else {
      setClips([]);
    }
  }, [residentId, folderId, situation]);

  if (!resident || clips === null) return <ActivityIndicator className="mt-10" />;

  const asleep = isAsleep(resident.sleepingHours);
  const familyClips = clips.filter((c) => !c.isFacilityContent);
  const facilityClips = clips.filter((c) => c.isFacilityContent);
  // Most-played first, standing in for ranking until best-match exists.
  const ordered = [...familyClips].sort((a, b) => b.helpedCount - a.helpedCount);

  const play = (clip: Clip) => {
    if (asleep) {
      logEvent({
        type: 'play_blocked_sleeping_hours',
        residentId,
        clipId: clip.id,
      });
      return;
    }
    logEvent({ type: 'play_requested', residentId, clipId: clip.id });
    router.push(`/resident/${residentId}/playing?clipId=${clip.id}`);
  };

  return (
    <ScrollView className="flex-1 bg-page" contentContainerClassName="p-4 pb-10">
      {fromName && (
        <Text className="text-base text-ink-muted">Asking for {fromName}</Text>
      )}

      {ordered.length === 0 ? (
        <Text className="mb-5 text-3xl font-bold text-ink">
          No family video for this yet
        </Text>
      ) : (
        <Text className="mb-5 text-3xl font-bold text-ink">Their videos</Text>
      )}

      {asleep && (
        <View className="mb-5 rounded-2xl bg-status-purple-bg p-4">
          <Text className="text-base text-status-purple-text">
            <Text className="font-semibold">Sleeping hours. </Text>
            Playing is paused until {describeWindow(resident.sleepingHours)}.
          </Text>
        </View>
      )}

      {ordered.map((clip) => (
        <View
          key={clip.id}
          className="mb-4 overflow-hidden rounded-2xl border border-border bg-surface"
        >
          <View className="h-40 items-center justify-center bg-nav">
            <Ionicons name="play-circle" size={56} color="rgb(255 248 239)" />
            <Text className="absolute bottom-3 right-4 text-base font-semibold text-on-nav">
              {formatDuration(clip.duration)}
            </Text>
          </View>

          <View className="p-4">
            <Text className="text-xl font-bold text-ink">
              &quot;{clip.title}&quot;
            </Text>
            <Text className="mt-1 text-base text-ink-muted">
              From {clip.fromName} · {clip.recordedOn}
            </Text>

            <View className="mt-3 flex-row">
              <View className="rounded-full bg-status-gray-bg px-3 py-1.5">
                <Text className="text-sm font-semibold text-status-gray-text">
                  Played {clip.playCount}× · helped {clip.helpedCount}
                </Text>
              </View>
            </View>

            {clip.familyNote && (
              <Text className="mt-3 text-base leading-6 text-ink">
                <Text className="font-bold">Family note: </Text>
                {clip.familyNote}
              </Text>
            )}

            <Pressable
              onPress={() => play(clip)}
              disabled={asleep}
              className={`mt-4 flex-row items-center justify-center rounded-xl py-4 ${
                asleep ? 'bg-status-gray-bg' : 'bg-accent active:opacity-80'
              }`}
            >
              <Ionicons
                name="play"
                size={20}
                color={asleep ? 'rgb(91 96 145)' : 'rgb(30 33 64)'}
              />
              <Text
                className={`ml-2 text-lg font-semibold ${
                  asleep ? 'text-status-gray-text' : 'text-on-accent'
                }`}
              >
                {asleep
                  ? 'Playing paused'
                  : `Play on Room ${resident.room} tablet`}
              </Text>
            </Pressable>
          </View>
        </View>
      ))}

      {/* S6 top half — a facility clip to fall back on */}
      {ordered.length === 0 && facilityClips.length > 0 && (
        <>
          <Text className="mb-2 text-lg font-semibold text-ink-muted">
            Try instead
          </Text>
          {facilityClips.map((clip) => (
            <Pressable
              key={clip.id}
              onPress={() => play(clip)}
              disabled={asleep}
              className={`mb-4 flex-row items-center rounded-2xl border border-border bg-surface p-3 ${
                asleep ? 'opacity-60' : 'active:opacity-70'
              }`}
            >
              <View className="mr-3 h-14 w-20 items-center justify-center rounded-xl bg-nav">
                <Ionicons name="play" size={20} color="rgb(255 248 239)" />
              </View>
              <View className="flex-1">
                <Text className="text-lg font-semibold text-ink">
                  &quot;{clip.title}&quot;
                </Text>
                <Text className="text-base text-ink-muted">
                  {clip.fromName} · {formatDuration(clip.duration)}
                </Text>
              </View>
            </Pressable>
          ))}
        </>
      )}

      <Pressable
        onPress={() =>
          router.push(
            `/resident/${residentId}/request?situation=${situation ?? ''}`
          )
        }
        className="mt-2 items-center py-3 active:opacity-70"
      >
        <Text className="text-lg font-semibold text-link underline">
          None fit? Ask the caregiver for a video
        </Text>
      </Pressable>
    </ScrollView>
  );
}
