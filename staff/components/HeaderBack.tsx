/**
 * The header's back arrow. Goes back when there is somewhere to go back to;
 * otherwise (the page was reloaded or opened from a link) it goes to the
 * screen above this one, so every screen but Residents has a way out.
 */

import { Pressable, Text } from 'react-native';
import { router, type Href } from 'expo-router';
import { Ionicons } from '@expo/vector-icons';

import { useThemeColors } from '../lib/theme';

export default function HeaderBack({ parent }: { parent: Href }) {
  const colors = useThemeColors();

  const goBack = () => {
    if (router.canGoBack()) router.back();
    else router.replace(parent);
  };

  return (
    <Pressable
      onPress={goBack}
      accessibilityRole="button"
      accessibilityLabel="Back"
      hitSlop={12}
      className="flex-row items-center pr-3 active:opacity-70"
    >
      <Ionicons name="chevron-back" size={26} color={colors.onNav} />
      <Text className="text-lg text-on-nav">Back</Text>
    </Pressable>
  );
}
