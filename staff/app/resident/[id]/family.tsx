/**
 * S3 · Asking for whom
 *
 * Family members are folders for grouping clips, not accounts. A folder with a
 * caregiver-written guidance note is highlighted — that note is the whole point
 * of the screen in cases like a spouse who has died, where saying the wrong
 * thing causes real harm. Staff can read it; only caregivers can edit it.
 */

import { useEffect, useState } from 'react';
import { ActivityIndicator, Pressable, ScrollView, Text, View } from 'react-native';
import { router, useLocalSearchParams } from 'expo-router';
import { Ionicons } from '@expo/vector-icons';

import { getFamilyFolders } from '../../../data/mock';
import type { FamilyFolder } from '../../../data/types';
import { logEvent } from '../../../lib/log';

export default function FamilyScreen() {
  const { id, situation } = useLocalSearchParams<{
    id: string;
    situation?: string;
  }>();
  const residentId = Number(id);

  const [folders, setFolders] = useState<FamilyFolder[] | null>(null);

  useEffect(() => {
    getFamilyFolders(residentId).then(setFolders);
  }, [residentId]);

  const openFolder = (folder: FamilyFolder) => {
    logEvent({
      type: 'family_folder_opened',
      residentId,
      folderId: folder.id,
    });
    router.push(
      `/resident/${residentId}/clips?folderId=${folder.id}&situation=${situation ?? ''}`
    );
  };

  return (
    <ScrollView className="flex-1 bg-page" contentContainerClassName="p-4 pb-10">
      <Text className="mb-5 text-3xl font-bold text-ink">
        Who are they asking for?
      </Text>

      {folders === null ? (
        <ActivityIndicator className="mt-8" />
      ) : (
        <>
          {folders.map((folder) =>
            folder.guidanceNote ? (
              // Guidance note: read-only for staff, written by the caregiver.
              <View
                key={folder.id}
                className="mb-3 rounded-2xl border-2 border-status-yellow-text bg-status-yellow-bg p-4"
              >
                <View className="flex-row items-center">
                  <View className="mr-4 h-12 w-12 items-center justify-center rounded-full bg-status-yellow-bg">
                    <Ionicons name="person" size={24} color="rgb(61 52 25)" />
                  </View>
                  <View className="flex-1">
                    <Text className="text-lg font-semibold text-status-yellow-text">
                      {folder.name}
                    </Text>
                    <Text className="text-base text-status-yellow-text">
                      {folder.relationship} ·{' '}
                      {folder.clipCount === 0
                        ? 'no videos'
                        : `${folder.clipCount} video${folder.clipCount === 1 ? '' : 's'}`}
                    </Text>
                  </View>
                </View>

                <Text className="mt-3 text-base leading-6 text-status-yellow-text">
                  <Text className="font-bold">Family&apos;s guidance: </Text>
                  {folder.guidanceNote}
                </Text>
              </View>
            ) : (
              <Pressable
                key={folder.id}
                onPress={() => openFolder(folder)}
                className="mb-3 flex-row items-center rounded-2xl border border-border bg-surface p-4 active:opacity-70"
              >
                <View className="mr-4 h-12 w-12 items-center justify-center rounded-full bg-page">
                  <Ionicons name="person" size={24} color="rgb(91 96 145)" />
                </View>
                <View className="flex-1">
                  <Text className="text-lg font-semibold text-ink">
                    {folder.name}
                  </Text>
                  <Text className="text-base text-ink-muted">
                    {folder.relationship} ·{' '}
                    {folder.clipCount === 1
                      ? '1 video'
                      : `${folder.clipCount} videos`}
                  </Text>
                </View>
                <Ionicons
                  name="chevron-forward"
                  size={22}
                  color="rgb(91 96 145)"
                />
              </Pressable>
            )
          )}

          <Pressable
            onPress={() => router.push(`/resident/${residentId}/request`)}
            className="mt-2 items-center rounded-2xl border border-dashed border-border py-4 active:opacity-70"
          >
            <Text className="text-lg font-semibold text-ink">Someone else</Text>
          </Pressable>
        </>
      )}
    </ScrollView>
  );
}
