import React, { useState } from "react";
import { LayoutChangeEvent, StyleSheet, Text, View } from "react-native";
import { MaterialCommunityIcons } from "@expo/vector-icons";
import { useNow } from "../../clock/useNow";
import { fonts } from "../../theme/fonts";
import { borders, colors, fontSizes, legibility, space } from "../../theme/theme";
import { fitFontSizeToBox, LINE_HEIGHT, linesAt } from "../fitText";
import type { ReminderData, ReminderItem } from "../types";
import { reminderIcon } from "./icons";

/** How many reminders fit the widget at once. */
export const REMINDERS_SHOWN = 2;

/** Room for a reminder's text: one line at full size, or two at the smallest. */
const TEXT_HEIGHT = 2 * legibility.minText * LINE_HEIGHT;

/** The reminders still to come at `now`, soonest first, at most `REMINDERS_SHOWN`. */
export function upcomingReminders(items: ReminderItem[], now: Date): ReminderItem[] {
  const minute = now.getHours() * 60 + now.getMinutes();
  return items.filter((item) => item.minute >= minute).slice(0, REMINDERS_SHOWN);
}

type ViewProps = ReminderData & { now: Date };

/** The next reminders at `now`, the soonest marked; once the day's are over, says so. */
export function ReminderView({ items, now }: ViewProps) {
  const upcoming = upcomingReminders(items, now);

  if (!upcoming.length) {
    return (
      <View testID="reminder" style={styles.root}>
        <Text style={styles.done}>Nothing more today</Text>
      </View>
    );
  }

  return (
    <View testID="reminder" style={styles.root}>
      {upcoming.map((item, i) => (
        <ReminderRow key={`${item.minute}-${item.text}`} item={item} next={i === 0} />
      ))}
    </View>
  );
}

/** One reminder, its text as large as fits beside the icon and time. */
function ReminderRow({ item, next }: { item: ReminderItem; next: boolean }) {
  const [width, setWidth] = useState(0);
  const onLayout = (e: LayoutChangeEvent) => setWidth(e.nativeEvent.layout.width);
  const size = fitFontSizeToBox(item.text, width, TEXT_HEIGHT, fontSizes.name, legibility.minText);

  return (
    <View testID={next ? "reminder-next" : undefined} style={[styles.row, next && styles.next]}>
      <MaterialCommunityIcons name={reminderIcon(item.icon)} size={fontSizes.name} color={colors.day.text} />
      <Text style={styles.time} numberOfLines={1}>
        {item.timeLabel}
      </Text>
      {/* Hidden until measured, so it doesn't show at full size and then shrink. */}
      <Text
        style={[styles.text, { fontSize: size, lineHeight: size * LINE_HEIGHT, opacity: width ? 1 : 0 }]}
        numberOfLines={linesAt(size, TEXT_HEIGHT)}
        onLayout={onLayout}
      >
        {item.text}
      </Text>
    </View>
  );
}

/** The day's reminders, kept current from the app clock. */
export default function Reminder({ data }: { data: ReminderData }) {
  const now = useNow();
  return <ReminderView items={data.items} now={now} />;
}

const styles = StyleSheet.create({
  root: { flex: 1, justifyContent: "center", gap: space.xs },
  row: {
    flexDirection: "row",
    alignItems: "center",
    gap: space.xs,
    paddingLeft: space.xs,
    borderLeftWidth: borders.accent,
    borderLeftColor: "transparent",
  },
  next: { borderLeftColor: colors.day.accent },
  time: { fontFamily: fonts.bold, fontSize: fontSizes.label, color: colors.day.text },
  text: { flex: 1, fontFamily: fonts.bold, color: colors.day.text },
  done: { fontFamily: fonts.regular, fontSize: fontSizes.label, color: colors.day.textSecondary, textAlign: "center" },
});
