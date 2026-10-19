/**
 * S1 · Pick resident
 *
 * Staff see every resident on the server: for the demo, every resident is
 * linked to every staff member, with no floor scoping and no QR scan.
 */

import { useCallback, useEffect, useState } from 'react';
import {
  ActivityIndicator,
  Pressable,
  ScrollView,
  Text,
  View,
} from 'react-native';
import { router } from 'expo-router';
import { Ionicons } from '@expo/vector-icons';

import LoadError from '../components/LoadError';
import { getResidents } from '../data/server';
import type { Resident } from '../data/types';
import { logEvent } from '../lib/log';
import { useThemeColors } from '../lib/theme';

/** Placeholder until staff accounts exist. */
const STAFF_NAME = 'Sam';

export default function ResidentsScreen() {
  const [residents, setResidents] = useState<Resident[] | null>(null);
  const [failed, setFailed] = useState(false);
  const colors = useThemeColors();

  const load = useCallback(() => {
    setFailed(false);
    setResidents(null);
    getResidents()
      .then(setResidents)
      .catch(() => setFailed(true));
  }, []);

  useEffect(load, [load]);

  const openResident = (resident: Resident) => {
    logEvent({ type: 'resident_opened', residentId: resident.id });
    router.push(`/resident/${resident.id}`);
  };

  return (
    <ScrollView
      className="flex-1 bg-page"
      contentContainerClassName="p-4 pb-10"
    >
      <Text className="text-base text-ink-muted">Hi, {STAFF_NAME}</Text>
      <Text className="mb-6 text-3xl font-bold text-ink">Residents</Text>

      {failed ? (
        <LoadError onRetry={load} />
      ) : residents === null ? (
        <ActivityIndicator className="mt-8" />
      ) : residents.length === 0 ? (
        <Text className="text-base text-ink-muted">No residents yet.</Text>
      ) : (
        residents.map((resident) => (
          <Pressable
            key={resident.id}
            onPress={() => openResident(resident)}
            className="mb-3 flex-row items-center rounded-2xl border border-border bg-surface p-4 active:opacity-70"
          >
            <View className="mr-4 h-12 w-12 items-center justify-center rounded-full bg-page">
              <Ionicons name="person" size={24} color={colors.inkMuted} />
            </View>
            <Text className="flex-1 text-lg font-semibold text-ink">
              {resident.name}
            </Text>
            <Ionicons name="chevron-forward" size={22} color={colors.inkMuted} />
          </Pressable>
        ))
      )}
    </ScrollView>
  );
}
