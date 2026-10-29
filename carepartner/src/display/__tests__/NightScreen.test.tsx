import React from "react";
import { render, screen } from "@testing-library/react-native";
import { setClock } from "../../clock/clock";
import NightScreen, { NIGHT_LINE, NIGHT_WORD } from "../NightScreen";

describe("NightScreen", () => {
  afterEach(() => setClock(null));

  it("shows the day, Nighttime and one reassuring line", async () => {
    setClock(new Date("2026-10-12T23:00:00"));
    await render(<NightScreen />);

    expect(screen.getByText("Monday")).toBeTruthy();
    expect(screen.getByText("Nighttime")).toBeTruthy();
    expect(NIGHT_WORD).toBe("Nighttime");
    expect(screen.getByTestId("night-moon")).toBeTruthy();
    expect(screen.getByText(NIGHT_LINE)).toBeTruthy();
  });

  it("shows the new day after midnight", async () => {
    setClock(new Date("2026-10-13T01:00:00"));
    await render(<NightScreen />);

    expect(screen.getByText("Tuesday")).toBeTruthy();
  });
});
