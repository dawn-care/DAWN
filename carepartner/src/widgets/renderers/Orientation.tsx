import React, { useState } from "react";
import { LayoutChangeEvent, StyleSheet, Text, View } from "react-native";
import { MaterialCommunityIcons } from "@expo/vector-icons";
import { useNow } from "../../clock/useNow";
import { fonts } from "../../theme/fonts";
import { colors, fontSizes, legibility, space } from "../../theme/theme";
import { fitFontSize } from "../fitText";
import { dateLabel, dayName, timeOfDay } from "../timeOfDay";
import type { OrientationData } from "../types";

const LABELS = { morning: "Morning", afternoon: "Afternoon", evening: "Evening" } as const;
const ICONS = { morning: "weather-sunny", afternoon: "weather-sunny", evening: "weather-night" } as const;

const MIN = legibility.minOrientationText;

type ViewProps = OrientationData & { now: Date };

/** Draws the day and time of day for the given time, shrinking long words to fit (never below 76px). */
export function OrientationView({ now, showDate }: ViewProps) {
  const [width, setWidth] = useState(0);
  const onLayout = (e: LayoutChangeEvent) => setWidth(e.nativeEvent.layout.width);

  const period = timeOfDay(now);
  const day = dayName(now);
  const daySize = fitFontSize(day, width, fontSizes.day, MIN);
  const periodSize = fitFontSize(LABELS[period], width - fontSizes.timeOfDay - space.sm, fontSizes.timeOfDay, MIN);

  return (
    <View testID="orientation" style={styles.root} onLayout={onLayout}>
      <Text testID="orientation-day" style={[styles.day, { fontSize: daySize, lineHeight: daySize * 1.05 }]} numberOfLines={1}>
        {day}
      </Text>
      <View style={styles.periodRow}>
        <MaterialCommunityIcons name={ICONS[period]} size={periodSize} color={colors.day.text} />
        <Text style={[styles.period, { fontSize: periodSize }]} numberOfLines={1}>
          {LABELS[period]}
        </Text>
      </View>
      {showDate && <Text style={styles.date}>{dateLabel(now)}</Text>}
    </View>
  );
}

/** The day and time of day, kept current from the app clock. */
export default function Orientation({ data }: { data: OrientationData }) {
  const now = useNow();
  return <OrientationView now={now} showDate={data.showDate} />;
}

const styles = StyleSheet.create({
  root: { flex: 1, justifyContent: "center", gap: space.xs },
  day: {
    fontFamily: fonts.heavy,
    color: colors.day.text,
  },
  periodRow: { flexDirection: "row", alignItems: "center", gap: space.sm },
  period: {
    fontFamily: fonts.bold,
    color: colors.day.text,
  },
  date: {
    fontFamily: fonts.regular,
    fontSize: fontSizes.name,
    color: colors.day.textSecondary,
  },
});
