/**
 * S2 · What's happening
 *
 * The situation picker for one resident. Also the place staff edit sleeping
 * hours, since the rule is "wherever staff can play content, they can edit
 * sleeping hours".
 *
 * While the resident is within their sleeping hours, playing is blocked. Staff
 * can still browse content — only the play action is disabled (on S4).
 */

import { useEffect, useState } from 'react';
import {
  ActivityIndicator,
  Pressable,
  ScrollView,
  Text,
  TextInput,
  View,
} from 'react-native';
import { router, useLocalSearchParams } from 'expo-router';
import { Ionicons } from '@expo/vector-icons';

import { SITUATIONS, getResident, setSleepingHours } from '../../../data/mock';
import type { Resident, Situation } from '../../../data/types';
import { describeWindow, isAsleep } from '../../../lib/sleepingHours';
import { logEvent } from '../../../lib/log';

const RATING_LABEL: Record<string, string> = {
  helped: 'Helped',
  no_change: 'No change',
  worse: 'Made it worse',
};

export default function ResidentScreen() {
  const { id } = useLocalSearchParams<{ id: string }>();
  const residentId = Number(id);

  const [resident, setResident] = useState<Resident | null>(null);
  const [editingHours, setEditingHours] = useState(false);
  const [draftStart, setDraftStart] = useState('');
  const [draftEnd, setDraftEnd] = useState('');

  useEffect(() => {
    getResident(residentId).then((r) => setResident(r ?? null));
  }, [residentId]);

  if (!resident) return <ActivityIndicator className="mt-10" />;

  const asleep = isAsleep(resident.sleepingHours);

  const chooseSituation = (situation: Situation) => {
    logEvent({
      type: 'situation_selected',
      residentId: resident.id,
      situation: situation.id,
    });
    const next = situation.routesThroughFamily ? 'family' : 'clips';
    router.push(`/resident/${resident.id}/${next}?situation=${situation.id}`);
  };

  const startEditing = () => {
    setDraftStart(resident.sleepingHours.start);
    setDraftEnd(resident.sleepingHours.end);
    setEditingHours(true);
  };

  const saveHours = async () => {
    const valid = /^\d{2}:\d{2}$/;
    if (!valid.test(draftStart) || !valid.test(draftEnd)) return;

    await setSleepingHours(resident.id, { start: draftStart, end: draftEnd });
    logEvent({
      type: 'sleeping_hours_changed',
      residentId: resident.id,
      start: draftStart,
      end: draftEnd,
    });
    setResident({
      ...resident,
      sleepingHours: { start: draftStart, end: draftEnd },
    });
    setEditingHours(false);
  };

  return (
    <ScrollView
      className="flex-1 bg-page"
      contentContainerClassName="p-4 pb-10"
    >
      {/* Resident header */}
      <View className="mb-6 flex-row items-center rounded-2xl border border-border bg-surface p-4">
        <View className="mr-4 h-14 w-14 items-center justify-center rounded-full bg-page">
          <Ionicons name="person" size={28} color="rgb(91 96 145)" />
        </View>
        <View className="flex-1">
          <Text className="text-xl font-semibold text-ink">
            {resident.name}
          </Text>
          <Text className="text-base text-ink-muted">Room {resident.room}</Text>
        </View>
      </View>

      {/* Sleeping-hours banner, shown only while playing is blocked */}
      {asleep && (
        <View className="mb-6 rounded-2xl bg-status-purple-bg p-4">
          <View className="mb-1 flex-row items-center">
            <Ionicons name="moon" size={18} color="rgb(52 41 90)" />
            <Text className="ml-2 text-base font-semibold text-status-purple-text">
              Sleeping hours
            </Text>
          </View>
          <Text className="text-base text-status-purple-text">
            Playing is paused until {describeWindow(resident.sleepingHours)}
            {'. '}
            You can still look through their videos.
          </Text>
        </View>
      )}

      <Text className="mb-4 text-3xl font-bold text-ink">
        What&apos;s happening?
      </Text>

      <View className="flex-row flex-wrap justify-between">
        {SITUATIONS.map((situation) => (
          <Pressable
            key={situation.id}
            onPress={() => chooseSituation(situation)}
            className="mb-3 w-[48.5%] rounded-2xl border border-border bg-surface p-4 active:opacity-70"
          >
            <Ionicons
              name={situation.icon as never}
              size={26}
              color="rgb(224 122 95)"
            />
            <Text className="mt-3 text-lg font-semibold leading-6 text-ink">
              {situation.label}
            </Text>
          </Pressable>
        ))}
      </View>

      {/* Sleeping hours — editable by staff and caregivers */}
      <View className="mt-3 rounded-2xl border border-border bg-surface p-4">
        <View className="flex-row items-center justify-between">
          <View className="flex-1">
            <Text className="text-base font-semibold text-ink">
              Sleeping hours
            </Text>
            <Text className="text-base text-ink-muted">
              {describeWindow(resident.sleepingHours)}
            </Text>
          </View>
          {!editingHours && (
            <Pressable onPress={startEditing} className="px-2 py-1">
              <Text className="text-base font-semibold text-link">Edit</Text>
            </Pressable>
          )}
        </View>

        {editingHours && (
          <View className="mt-4">
            <View className="flex-row">
              <View className="mr-3 flex-1">
                <Text className="mb-1 text-sm text-ink-muted">
                  Sleeps at (24h)
                </Text>
                <TextInput
                  value={draftStart}
                  onChangeText={setDraftStart}
                  placeholder="20:30"
                  placeholderTextColor="rgb(169 173 208)"
                  className="rounded-xl border border-border bg-page p-3 text-base text-ink"
                />
              </View>
              <View className="flex-1">
                <Text className="mb-1 text-sm text-ink-muted">
                  Wakes at (24h)
                </Text>
                <TextInput
                  value={draftEnd}
                  onChangeText={setDraftEnd}
                  placeholder="07:00"
                  placeholderTextColor="rgb(169 173 208)"
                  className="rounded-xl border border-border bg-page p-3 text-base text-ink"
                />
              </View>
            </View>

            <View className="mt-4 flex-row">
              <Pressable
                onPress={saveHours}
                className="mr-3 flex-1 items-center rounded-xl bg-accent py-3 active:opacity-80"
              >
                <Text className="text-base font-semibold text-on-accent">
                  Save
                </Text>
              </Pressable>
              <Pressable
                onPress={() => setEditingHours(false)}
                className="flex-1 items-center rounded-xl border border-border py-3 active:opacity-70"
              >
                <Text className="text-base font-semibold text-ink">Cancel</Text>
              </Pressable>
            </View>
          </View>
        )}
      </View>

      {/* Last played */}
      {resident.lastPlayed && (
        <View className="mt-3 rounded-2xl bg-surface-raised p-4">
          <Text className="text-base text-ink-muted">
            <Text className="font-semibold text-ink">Last played: </Text>
            {resident.lastPlayed.title} from {resident.lastPlayed.fromName} ·{' '}
            {resident.lastPlayed.at}
            {resident.lastPlayed.rating
              ? ` · rated ${RATING_LABEL[resident.lastPlayed.rating]}`
              : ''}
          </Text>
        </View>
      )}
    </ScrollView>
  );
}
