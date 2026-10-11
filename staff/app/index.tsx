/**
 * S1 · Pick resident
 *
 * Staff see every resident (no floor scoping, no QR scan — both were cut from
 * the MVP). A "New video" badge marks residents whose family has added content.
 */

import { useEffect, useState } from 'react';
import {
  ActivityIndicator,
  Pressable,
  ScrollView,
  Text,
  View,
} from 'react-native';
import { router } from 'expo-router';
import { Ionicons } from '@expo/vector-icons';

import { getResidents } from '../data/mock';
import type { Resident } from '../data/types';
import { logEvent } from '../lib/log';

/** Placeholder until staff accounts exist. */
const STAFF_NAME = 'Sam';

export default function ResidentsScreen() {
  const [residents, setResidents] = useState<Resident[] | null>(null);

  useEffect(() => {
    getResidents().then(setResidents);
  }, []);

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

      {residents === null ? (
        <ActivityIndicator className="mt-8" />
      ) : (
        residents.map((resident) => (
          <Pressable
            key={resident.id}
            onPress={() => openResident(resident)}
            className="mb-3 flex-row items-center rounded-2xl border border-border bg-surface p-4 active:opacity-70"
          >
            <View className="mr-4 h-12 w-12 items-center justify-center rounded-full bg-page">
              <Ionicons name="person" size={24} color="rgb(91 96 145)" />
            </View>

            <View className="flex-1">
              <Text className="text-lg font-semibold text-ink">
                {resident.name}
              </Text>
              <Text className="text-base text-ink-muted">
                Room {resident.room}
              </Text>
            </View>

            {resident.hasNewContent && (
              <View className="rounded-full bg-status-blue-bg px-3 py-1.5">
                <Text className="text-sm font-semibold text-status-blue-text">
                  New video
                </Text>
              </View>
            )}
          </Pressable>
        ))
      )}
    </ScrollView>
  );
}
