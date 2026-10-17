import type { ComponentProps } from "react";
import type { MaterialCommunityIcons } from "@expo/vector-icons";

type IconName = ComponentProps<typeof MaterialCommunityIcons>["name"];

/**
 * The tile icons the server can ask for, by our own key, so the server never
 * depends on an icon library's names. Add a key here to support a new icon.
 */
export const TILE_ICONS: Record<string, IconName> = {
  sunny: "weather-sunny",
  "partly-cloudy": "weather-partly-cloudy",
  cloudy: "weather-cloudy",
  rain: "weather-rainy",
  snow: "weather-snowy",
  storm: "weather-lightning",
  baseball: "baseball",
  football: "football",
  basketball: "basketball",
  hockey: "hockey-puck",
  soccer: "soccer",
  fog: "weather-fog",
};

/** Shown for an icon key the app doesn't know. */
export const FALLBACK_ICON: IconName = "circle-outline";

export function tileIcon(key: string): IconName {
  return TILE_ICONS[key] ?? FALLBACK_ICON;
}

/** Reminder icons, by the keys the server allows (`ReminderIcons` in the server's screen widget config). */
export const REMINDER_ICONS: Record<string, IconName> = {
  meal: "silverware-fork-knife",
  medication: "pill",
  walk: "walk",
  visit: "account-group",
  appointment: "calendar-clock",
  call: "phone",
  rest: "bed",
  other: "bell-outline",
};

export function reminderIcon(key: string): IconName {
  return Object.prototype.hasOwnProperty.call(REMINDER_ICONS, key) ? REMINDER_ICONS[key] : REMINDER_ICONS.other;
}
