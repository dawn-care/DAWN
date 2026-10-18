import React from "react";
import { Text } from "react-native";
import { act, render, screen } from "@testing-library/react-native";
import { setClock } from "../clock";
import { useNow } from "../useNow";

function ShowMinutes() {
  return <Text>{useNow().getMinutes()}</Text>;
}

describe("useNow", () => {
  beforeEach(() => {
    jest.useFakeTimers();
    setClock(new Date("2026-10-05T09:15:40"));
  });

  afterEach(() => {
    setClock(null);
    jest.useRealTimers();
  });

  it("starts at the app clock's time and changes as each minute starts", async () => {
    await render(<ShowMinutes />);
    expect(screen.getByText("15")).toBeTruthy();

    await act(async () => jest.advanceTimersByTime(19_999));
    expect(screen.getByText("15")).toBeTruthy();

    await act(async () => jest.advanceTimersByTime(1));
    expect(screen.getByText("16")).toBeTruthy();
  });
});
