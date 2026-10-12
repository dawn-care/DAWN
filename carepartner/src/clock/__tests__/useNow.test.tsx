import React from "react";
import { Text } from "react-native";
import { act, render, screen } from "@testing-library/react-native";
import { setClock } from "../clock";
import { useNow } from "../useNow";

function ShowMinutes() {
  return <Text>{useNow(30_000).getMinutes()}</Text>;
}

describe("useNow", () => {
  beforeEach(() => {
    jest.useFakeTimers();
    setClock(new Date("2026-10-05T09:15:00"));
  });

  afterEach(() => {
    setClock(null);
    jest.useRealTimers();
  });

  it("starts at the app clock's time and keeps up as time passes", async () => {
    await render(<ShowMinutes />);
    expect(screen.getByText("15")).toBeTruthy();

    await act(async () => jest.advanceTimersByTime(60_000));

    expect(screen.getByText("16")).toBeTruthy();
  });
});
