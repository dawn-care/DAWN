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

import { useCallback, useEffect, useState } from 'react';
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

import LastPlayed from '../../../components/LastPlayed';
import LoadError from '../../../components/LoadError';
import { SITUATIONS, getSleepingHours, setSleepingHours } from '../../../data/mock';
import { getResident } from '../../../data/server';
import type { Resident, Situation, SleepingHours } from '../../../data/types';
import {
  describeWindow,
  formatTime,
  formatTimeInput,
  isAsleep,
  isValidTime,
} from '../../../lib/sleepingHours';
import { logEvent } from '../../../lib/log';
import { useThemeColors } from '../../../lib/theme';

export default function ResidentScreen() {
  const { id } = useLocalSearchParams<{ id: string }>();
  const residentId = Number(id);
  const colors = useThemeColors();

  const [resident, setResident] = useState<Resident | null>(null);
  // Sleeping hours are fake for now: edits last until the app reloads.
  const [hours, setHours] = useState<SleepingHours | null>(null);
  const [failed, setFailed] = useState(false);
  const [editingHours, setEditingHours] = useState(false);
  const [draftStart, setDraftStart] = useState('');
  const [draftEnd, setDraftEnd] = useState('');
  const [hoursError, setHoursError] = useState<string | null>(null);

  const load = useCallback(() => {
    setFailed(false);
    Promise.all([getResident(residentId), getSleepingHours(residentId)])
      .then(([r, h]) => {
        if (!r) throw new Error('resident not found');
        setResident(r);
        setHours(h);
      })
      .catch(() => setFailed(true));
  }, [residentId]);

  useEffect(load, [load]);

  if (failed) {
    return (
      <View className="flex-1 bg-page p-4">
        <LoadError onRetry={load} />
      </View>
    );
  }
  if (!resident || !hours) return <ActivityIndicator className="mt-10" />;

  const asleep = isAsleep(hours);

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
    setDraftStart(hours.start);
    setDraftEnd(hours.end);
    setHoursError(null);
    setEditingHours(true);
  };

  const saveHours = async () => {
    if (!isValidTime(draftStart) || !isValidTime(draftEnd)) {
      setHoursError('Enter both times as 4 digits, like 2030 for 8:30 pm.');
      return;
    }
    if (draftStart === draftEnd) {
      setHoursError("Sleep and wake times can't be the same.");
      return;
    }

    await setSleepingHours(resident.id, { start: draftStart, end: draftEnd });
    logEvent({
      type: 'sleeping_hours_changed',
      residentId: resident.id,
      start: draftStart,
      end: draftEnd,
    });
    setHours({ start: draftStart, end: draftEnd });
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
          <Ionicons name="person" size={28} color={colors.inkMuted} />
        </View>
        <View className="flex-1">
          <Text className="text-xl font-semibold text-ink">
            {resident.name}
          </Text>
        </View>
      </View>

      {/* Sleeping-hours banner, shown only while playing is blocked */}
      {asleep && (
        <View className="mb-6 rounded-2xl bg-status-purple-bg p-4">
          <View className="mb-1 flex-row items-center">
            <Ionicons name="moon" size={18} color={colors.statusPurpleText} />
            <Text className="ml-2 text-base font-semibold text-status-purple-text">
              Sleeping hours
            </Text>
          </View>
          <Text className="text-base text-status-purple-text">
            Playing is paused until {formatTime(hours.end)}. You can still
            look through their media.
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
              color={colors.accent}
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
              {describeWindow(hours)}
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
                  onChangeText={(text) => setDraftStart(formatTimeInput(text))}
                  keyboardType="number-pad"
                  maxLength={5}
                  placeholder="20:30"
                  placeholderTextColor={colors.placeholder}
                  className="rounded-xl border border-border bg-page p-3 text-base text-ink"
                />
              </View>
              <View className="flex-1">
                <Text className="mb-1 text-sm text-ink-muted">
                  Wakes at (24h)
                </Text>
                <TextInput
                  value={draftEnd}
                  onChangeText={(text) => setDraftEnd(formatTimeInput(text))}
                  keyboardType="number-pad"
                  maxLength={5}
                  placeholder="07:00"
                  placeholderTextColor={colors.placeholder}
                  className="rounded-xl border border-border bg-page p-3 text-base text-ink"
                />
              </View>
            </View>

            {hoursError && (
              <Text className="mt-3 text-base text-status-red-text">{hoursError}</Text>
            )}

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

      <LastPlayed residentId={resident.id} />
    </ScrollView>
  );
}
