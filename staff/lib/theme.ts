/**
 * Colors for the places a className can't reach: icons, spinners, placeholder
 * text and the navigation header. They mirror the tokens in global.css, so
 * keep the two in step when the palette changes.
 *
 * Light or dark follows the phone's setting until staff flip it with the
 * header toggle (see ThemeToggle).
 */

import { useColorScheme } from 'nativewind';

const PALETTE = {
  light: {
    page: 'rgb(255 248 239)',
    nav: 'rgb(30 33 64)',
    onNav: 'rgb(255 248 239)',
    ink: 'rgb(30 33 64)',
    inkMuted: 'rgb(91 96 145)',
    placeholder: 'rgb(169 173 208)',
    accent: 'rgb(224 122 95)',
    onAccent: 'rgb(30 33 64)',
    statusGrayText: 'rgb(91 96 145)',
    statusGreenText: 'rgb(31 61 44)',
    statusYellowText: 'rgb(61 52 25)',
    statusPurpleText: 'rgb(52 41 90)',
    statusRedText: 'rgb(74 34 38)',
  },
  dark: {
    page: 'rgb(20 22 43)',
    nav: 'rgb(15 17 36)',
    onNav: 'rgb(243 237 228)',
    ink: 'rgb(243 237 228)',
    inkMuted: 'rgb(169 173 208)',
    placeholder: 'rgb(91 96 145)',
    accent: 'rgb(224 122 95)',
    onAccent: 'rgb(30 33 64)',
    statusGrayText: 'rgb(212 214 232)',
    statusGreenText: 'rgb(174 221 188)',
    statusYellowText: 'rgb(242 217 140)',
    statusPurpleText: 'rgb(213 198 242)',
    statusRedText: 'rgb(244 180 180)',
  },
};

export type ThemeColors = (typeof PALETTE)['light'];

export function useThemeColors(): ThemeColors {
  const { colorScheme } = useColorScheme();
  return PALETTE[colorScheme === 'dark' ? 'dark' : 'light'];
}
