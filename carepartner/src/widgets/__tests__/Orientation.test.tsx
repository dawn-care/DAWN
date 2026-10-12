import React from "react";
import { act, fireEvent, render, screen } from "@testing-library/react-native";
import { setClock } from "../../clock/clock";
import Orientation from "../renderers/Orientation";

describe("Orientation widget", () => {
  afterEach(() => {
    setClock(null);
    jest.useRealTimers();
  });

  it("shows the day and time of day", async () => {
    setClock(new Date("2026-10-06T09:00:00"));

    await render(<Orientation data={{ showDate: false }} />);

    expect(screen.getByText("Tuesday")).toBeTruthy();
    expect(screen.getByText("Morning")).toBeTruthy();
    expect(screen.queryByText("October 6")).toBeNull();
  });

  it("shrinks a long day name to fit, but not below 76px", async () => {
    setClock(new Date("2026-10-07T09:00:00"));
    await render(<Orientation data={{ showDate: false }} />);

    await act(async () =>
      fireEvent(screen.getByTestId("orientation"), "layout", { nativeEvent: { layout: { width: 576, height: 458 } } }),
    );

    const size = Object.assign({}, ...[screen.getByTestId("orientation-day").props.style].flat()).fontSize;
    expect(screen.getByText("Wednesday")).toBeTruthy();
    expect(size).toBeLessThan(152);
    expect(size).toBeGreaterThanOrEqual(76);
  });

  it("shows the date when switched on", async () => {
    setClock(new Date("2026-10-06T09:00:00"));

    await render(<Orientation data={{ showDate: true }} />);

    expect(screen.getByText("October 6")).toBeTruthy();
  });

  it("moves on to the afternoon by itself", async () => {
    jest.useFakeTimers();
    setClock(new Date("2026-10-06T11:59:00"));
    await render(<Orientation data={{ showDate: false }} />);
    expect(screen.getByText("Morning")).toBeTruthy();

    await act(async () => jest.advanceTimersByTime(90_000));

    expect(screen.getByText("Afternoon")).toBeTruthy();
  });

  it("moves on to the next day at midnight", async () => {
    jest.useFakeTimers();
    setClock(new Date("2026-10-06T23:59:30"));
    await render(<Orientation data={{ showDate: true }} />);
    expect(screen.getByText("Tuesday")).toBeTruthy();

    await act(async () => jest.advanceTimersByTime(60_000));

    expect(screen.getByText("Wednesday")).toBeTruthy();
    expect(screen.getByText("October 7")).toBeTruthy();
  });
});
