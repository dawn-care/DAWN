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
export const SECONDARY_BUTTON = `h-11 px-4 rounded-[10px] border-2 border-dawn-control bg-white text-base font-bold text-dawn-ink cursor-pointer disabled:opacity-60 disabled:cursor-default ${FOCUS_RING}`;

// Filled accent 44px button: Save, confirm
export const PRIMARY_BUTTON = `h-11 px-4 rounded-[10px] border-2 border-dawn-accent bg-dawn-accent text-base font-bold text-white cursor-pointer hover:bg-dawn-accent-hover hover:border-dawn-accent-hover disabled:opacity-60 disabled:cursor-default ${FOCUS_RING}`;

// Bold accent text link, e.g. "See on calendar"
export const ACTION_LINK = `shrink-0 font-bold text-dawn-accent hover:text-dawn-accent-hover ${FOCUS_RING}`;
