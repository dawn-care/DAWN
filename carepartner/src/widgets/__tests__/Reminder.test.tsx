import React from "react";
import { act, fireEvent, render, screen, within } from "@testing-library/react-native";
import { TEXT_LIMITS } from "@shared";
import { fontSizes, legibility } from "../../theme/theme";
import { REMINDER_ICONS, reminderIcon } from "../renderers/icons";
import { ReminderView } from "../renderers/Reminder";
import type { ReminderItem } from "../types";

const item = (minute: number, timeLabel: string, text: string): ReminderItem => ({ minute, timeLabel, text, icon: "meal" });
const items = [item(8 * 60, "8 AM", "Breakfast"), item(12 * 60, "12 PM", "Lunch"), item(15 * 60, "3 PM", "Visit")];
const at = (time: string) => new Date(`2026-10-12T${time}:00`);

describe("Reminder widget", () => {
  it("shows the next two reminders, the soonest marked", async () => {
    await render(<ReminderView items={items} now={at("09:00")} />);

    expect(within(screen.getByTestId("reminder-next")).getByText("Lunch")).toBeTruthy();
    expect(screen.getByText("Visit")).toBeTruthy();
    expect(screen.queryByText("Breakfast")).toBeNull();
  });

  it("still shows a reminder during its minute", async () => {
    await render(<ReminderView items={items} now={at("12:00")} />);

    expect(within(screen.getByTestId("reminder-next")).getByText("Lunch")).toBeTruthy();
  });

  it("shows a short reminder at full size on one line", async () => {
    await render(<ReminderView items={items} now={at("09:00")} />);
    await act(async () => fireEvent(screen.getByText("Lunch"), "layout", { nativeEvent: { layout: { width: 300 } } }));

    expect(screen.getByText("Lunch")).toHaveStyle({ fontSize: fontSizes.name });
    expect(screen.getByText("Lunch").props.numberOfLines).toBe(1);
  });

  it("fits a reminder as long as the caregiver can write on two smaller lines", async () => {
    // About the room beside the icon and time on the 1280x800 design.
    const long = "Physical therapy in the gym room";
    expect(long).toHaveLength(TEXT_LIMITS.reminder);
    await render(<ReminderView items={[item(10 * 60, "10 AM", long)]} now={at("09:00")} />);
    await act(async () => fireEvent(screen.getByText(long), "layout", { nativeEvent: { layout: { width: 300 } } }));

    expect(screen.getByText(long)).toHaveStyle({ fontSize: legibility.minText });
    expect(screen.getByText(long).props.numberOfLines).toBe(2);
  });

  it("says so once the day's reminders are over", async () => {
    await render(<ReminderView items={items} now={at("16:00")} />);

    expect(screen.getByText("Nothing more today")).toBeTruthy();
  });
});

describe("reminder icons", () => {
  it("maps the server's keys and falls back for unknown ones", () => {
    expect(reminderIcon("medication")).toBe(REMINDER_ICONS.medication);
    expect(reminderIcon("constructor")).toBe(REMINDER_ICONS.other);
  });
});
