/**
 * Root layout.
 *
 * Header colors come from lib/theme so they follow light and dark mode. Every
 * screen but Residents gets a back arrow, even after a reload (see HeaderBack).
 *
 * Plays started from this phone are followed for the whole app (lib/activePlays),
 * and the mini player under every screen keeps them within reach.
 */

import '../global.css';

import { useEffect } from 'react';
import { Appearance, Platform, View } from 'react-native';
import { Stack, type Href } from 'expo-router';
import { StatusBar } from 'expo-status-bar';
import { colorScheme } from 'nativewind';
import { SafeAreaProvider } from 'react-native-safe-area-context';

import HeaderBack from '../components/HeaderBack';
import MiniPlayer from '../components/MiniPlayer';
import ThemeToggle from '../components/ThemeToggle';
import { ActivePlaysProvider } from '../lib/activePlays';
import { useThemeColors } from '../lib/theme';

/** Where a screen's back arrow leads when there is no history to go back to. */
function parentOf(screen: string, residentId: string | undefined): Href | null {
  if (screen === 'index') return null;
  if (screen === 'resident/[id]/index') return '/';
  if (screen === 'resident/[id]/playing') return `/resident/${residentId}/clips`;
  return `/resident/${residentId}`;
}

export default function RootLayout() {
  const colors = useThemeColors();

  // On a phone the app follows the system setting by itself. On web, dark mode
  // is a class on the page that starts off, so match the system once here.
  useEffect(() => {
    if (Platform.OS === 'web' && Appearance.getColorScheme() === 'dark') colorScheme.set('dark');
  }, []);

  return (
    <SafeAreaProvider>
      <ActivePlaysProvider>
        {/* The header is dark in both themes */}
        <StatusBar style="light" />
        <View style={{ flex: 1, backgroundColor: colors.page }}>
          <Stack
            screenOptions={({ route }) => {
              const parent = parentOf(route.name, (route.params as { id?: string } | undefined)?.id);
              return {
                headerStyle: { backgroundColor: colors.nav },
                headerTintColor: colors.onNav,
                headerTitleStyle: { fontWeight: '600' },
                contentStyle: { backgroundColor: colors.page },
                headerLeft: parent ? () => <HeaderBack parent={parent} /> : undefined,
                headerRight: () => <ThemeToggle />,
              };
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
            <Stack.Screen name="resident/[id]/clips" options={{ title: 'Media' }} />
            <Stack.Screen
              name="resident/[id]/playing"
              options={{ title: 'Playing' }}
            />
            <Stack.Screen
              name="resident/[id]/request"
              options={{ title: 'Ask for a video' }}
            />
          </Stack>
          <MiniPlayer />
        </View>
      </ActivePlaysProvider>
    </SafeAreaProvider>
  );
}
