/**
 * Shown when a screen can't load from the server, with a way to try again.
 */

import { Pressable, Text, View } from 'react-native';

export default function LoadError({ onRetry }: { onRetry: () => void }) {
  return (
    <View className="mt-6 rounded-2xl bg-status-red-bg p-4">
      <Text className="text-base font-semibold text-status-red-text">
        Couldn&apos;t reach the server
      </Text>
      <Text className="mt-1 text-base text-status-red-text">
        Check that it&apos;s running and that this device is on the same network.
      </Text>
      <Pressable onPress={onRetry} className="mt-3 self-start py-1 active:opacity-70">
        <Text className="text-base font-semibold text-link">Try again</Text>
      </Pressable>
    </View>
  );
}
