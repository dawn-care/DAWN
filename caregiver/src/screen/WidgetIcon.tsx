/**
 * WidgetIcon.tsx
 *
 * User: Caregiver
 *
 * The icon for each widget type, used on placeholder tiles (and later the add
 * bar), so a type always has the same picture in the editor.
 */

import { Bell, CalendarClock, Cloud, Image, MessageSquare, Newspaper, Trophy } from "lucide-react";
import type { LucideIcon } from "lucide-react";
import type { ScreenWidgetType } from "@shared";

const ICONS: Record<ScreenWidgetType, LucideIcon> = {
  day_time: CalendarClock,
  photo: Image,
  message: MessageSquare,
  reminder: Bell,
  weather: Cloud,
  sports: Trophy,
  news: Newspaper,
};

interface WidgetIconProps {
  type: ScreenWidgetType;
  size?: number;
  className?: string;
}

export default function WidgetIcon({ type, size = 28, className }: WidgetIconProps) {
  const Icon = ICONS[type];
  return <Icon size={size} strokeWidth={2} aria-hidden="true" className={className} />;
}
