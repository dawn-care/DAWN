/**
 * Header button that switches between light and dark. The choice lasts until
 * the app reloads; before that the app follows the phone's setting.
 */

import { Pressable } from 'react-native';
import { useColorScheme } from 'nativewind';
import { Ionicons } from '@expo/vector-icons';

import { useThemeColors } from '../lib/theme';

export default function ThemeToggle() {
  const { colorScheme, toggleColorScheme } = useColorScheme();
  const colors = useThemeColors();
  const dark = colorScheme === 'dark';

  return (
    <Pressable
      onPress={toggleColorScheme}
      accessibilityRole="button"
      accessibilityLabel={dark ? 'Switch to light mode' : 'Switch to dark mode'}
      hitSlop={12}
      className="px-1 active:opacity-70"
    >
      <Ionicons name={dark ? 'sunny' : 'moon'} size={22} color={colors.onNav} />
    </Pressable>
  );
}
