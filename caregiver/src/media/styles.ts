/**
 * styles.ts
 *
 * User: Caregiver
 *
 * Tailwind class strings for the buttons and links on the media page's
 * panels, so they look the same in every panel.
 */

import { FOCUS_RING } from "../components/ui/styles";

// Outlined 44px button: Rename, Remove, Cancel, Start over
export const SECONDARY_BUTTON = `h-11 px-4 rounded-[10px] border-2 border-ink-muted bg-surface text-base font-bold text-ink cursor-pointer disabled:opacity-60 disabled:cursor-default ${FOCUS_RING}`;

// Filled accent 44px button: Save, confirm
export const PRIMARY_BUTTON = `h-11 px-4 rounded-[10px] border-2 border-accent bg-accent text-base font-bold text-on-accent cursor-pointer hover:brightness-95 disabled:opacity-60 disabled:cursor-default ${FOCUS_RING}`;

// Bold accent text link, e.g. "See on calendar"
export const ACTION_LINK = `shrink-0 inline-flex items-center min-h-11 font-bold text-link hover:underline ${FOCUS_RING}`;
