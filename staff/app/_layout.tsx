/**
 * Root layout.
 *
 * Header styling uses the `nav` / `on-nav` semantic tokens so it follows the
 * palette rather than hard-coded colors.
 */

import '../global.css';

import { Stack } from 'expo-router';
import { StatusBar } from 'expo-status-bar';
import { SafeAreaProvider } from 'react-native-safe-area-context';

export default function RootLayout() {
  return (
    <SafeAreaProvider>
      <StatusBar style="light" />
      <Stack
        screenOptions={{
          headerStyle: { backgroundColor: 'rgb(30 33 64)' },
          headerTintColor: 'rgb(255 248 239)',
          headerTitleStyle: { fontWeight: '600' },
          contentStyle: { backgroundColor: 'rgb(255 248 239)' },
        }}
      >
        <Stack.Screen name="index" options={{ title: 'Residents' }} />
        <Stack.Screen
          name="resident/[id]/index"
          options={{ title: 'Resident' }}
        />
        <Stack.Screen
          name="resident/[id]/family"
          options={{ title: 'Asking for whom' }}
        />
        <Stack.Screen name="resident/[id]/clips" options={{ title: 'Videos' }} />
        <Stack.Screen
          name="resident/[id]/playing"
          options={{ title: 'Playing' }}
        />
        <Stack.Screen
          name="resident/[id]/request"
          options={{ title: 'Ask for a video' }}
        />
      </Stack>
    </SafeAreaProvider>
  );
}
