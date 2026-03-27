import React from "react";
import { Stack } from "expo-router";
import { ScheduleProvider } from "./ScheduleContext";

export default function RootLayout() {
  return (
    <ScheduleProvider>
      <Stack screenOptions={{ headerShown: false }} />
    </ScheduleProvider>
  );
}
