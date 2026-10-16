/**
 * S6 (bottom half) · Ask the caregiver for a video
 *
 * The mockup sent this request to family members directly. Per the team's
 * decision it goes to the caregiver instead, who then reaches out to family —
 * so there is no recipient picker here, only the topic and the message.
 */

import { useEffect, useState } from 'react';
import { ActivityIndicator, Pressable, ScrollView, Text, View } from 'react-native';
import { router, useLocalSearchParams } from 'expo-router';
import { Ionicons } from '@expo/vector-icons';

import { getResident } from '../../../data/server';
import type { Resident } from '../../../data/types';
import { logEvent } from '../../../lib/log';

const TOPICS = [
  { id: 'going_home', label: 'Going home' },
  { id: 'why_here', label: "Why they're here" },
  { id: 'family_visit', label: 'When family will visit' },
  { id: 'something_else', label: 'Something else' },
];

/** Facility name is a placeholder until staff accounts carry one. */
const FACILITY = "St. Ann's";

export default function RequestScreen() {
  const { id } = useLocalSearchParams<{ id: string }>();
  const residentId = Number(id);

  const [resident, setResident] = useState<Resident | null>(null);
  const [topic, setTopic] = useState(TOPICS[0].id);
  const [sent, setSent] = useState(false);

  useEffect(() => {
    getResident(residentId).then((r) => setResident(r ?? null));
  }, [residentId]);

  if (!resident) return <ActivityIndicator className="mt-10" />;

  const topicLabel = TOPICS.find((t) => t.id === topic)?.label.toLowerCase();
  const firstName = resident.name.split(' ')[0];

  const send = () => {
    logEvent({ type: 'caregiver_request_sent', residentId, topic });
    setSent(true);
  };

  if (sent) {
    return (
      <ScrollView className="flex-1 bg-page" contentContainerClassName="p-4">
        <View className="mt-6 rounded-2xl bg-status-green-bg p-5">
          <Ionicons name="checkmark-circle" size={32} color="rgb(31 61 44)" />
          <Text className="mt-2 text-xl font-bold text-status-green-text">
            Request sent
          </Text>
          <Text className="mt-1 text-base text-status-green-text">
            The caregiver has been asked for a video about {topicLabel}. They
            will reach out to the family.
          </Text>
        </View>

        <Pressable
          onPress={() => router.dismissTo(`/resident/${residentId}`)}
          className="mt-5 items-center rounded-xl bg-accent py-4 active:opacity-80"
        >
          <Text className="text-lg font-semibold text-on-accent">Done</Text>
        </Pressable>
      </ScrollView>
    );
  }

  return (
    <ScrollView className="flex-1 bg-page" contentContainerClassName="p-4 pb-10">
      <Text className="mb-1 text-3xl font-bold text-ink">
        Ask for a video
      </Text>
      <Text className="mb-6 text-base text-ink-muted">
        This goes to {firstName}&apos;s caregiver, who will ask the family.
      </Text>

      <Text className="mb-3 text-lg font-semibold text-ink">
        Keeps asking about
      </Text>

      <View className="mb-6 flex-row flex-wrap">
        {TOPICS.map((t) => {
          const selected = t.id === topic;
          return (
            <Pressable
              key={t.id}
              onPress={() => setTopic(t.id)}
              className={`mb-3 mr-3 rounded-full border px-5 py-3 active:opacity-70 ${
                selected
                  ? 'border-accent bg-accent'
                  : 'border-border bg-surface'
              }`}
            >
              <Text
                className={`text-base font-semibold ${
                  selected ? 'text-on-accent' : 'text-ink'
                }`}
              >
                {t.label}
              </Text>
            </Pressable>
          );
        })}
      </View>

      <Text className="mb-3 text-lg font-semibold text-ink">
        They&apos;ll receive
      </Text>

      <View className="rounded-2xl bg-surface-raised p-4">
        <Text className="text-base leading-6 text-ink">
          &quot;{resident.name} has been asking about {topicLabel}. Staff at{' '}
          {FACILITY} asked for a short video.&quot;
        </Text>
      </View>

      <Pressable
        onPress={send}
        className="mt-6 items-center rounded-xl bg-accent py-4 active:opacity-80"
      >
        <Text className="text-lg font-semibold text-on-accent">
          Send request
        </Text>
      </Pressable>
    </ScrollView>
  );
}
