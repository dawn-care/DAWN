import React from "react";
import { StyleSheet, Text, View } from "react-native";
import { MaterialCommunityIcons } from "@expo/vector-icons";
import { useNow } from "../../clock/useNow";
import { fonts } from "../../theme/fonts";
import { borders, colors, fontSizes, space } from "../../theme/theme";
import type { ReminderData, ReminderItem } from "../types";
import { reminderIcon } from "./icons";

/** How many reminders fit the widget at once. */
export const REMINDERS_SHOWN = 2;

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
        <View
          key={`${item.minute}-${item.text}`}
          testID={i === 0 ? "reminder-next" : undefined}
          style={[styles.row, i === 0 && styles.next]}
        >
          <MaterialCommunityIcons name={reminderIcon(item.icon)} size={fontSizes.name} color={colors.day.text} />
          <Text style={styles.time} numberOfLines={1}>
            {item.timeLabel}
          </Text>
          <Text style={styles.text} numberOfLines={1}>
            {item.text}
          </Text>
        </View>
      ))}
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
  text: { flex: 1, fontFamily: fonts.bold, fontSize: fontSizes.name, color: colors.day.text },
  done: { fontFamily: fonts.regular, fontSize: fontSizes.label, color: colors.day.textSecondary, textAlign: "center" },
});
