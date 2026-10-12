import React from "react";
import { render, screen } from "@testing-library/react-native";
import { setClock } from "../../clock/clock";
import Grid from "../Grid";
import { sampleScreen } from "../sampleScreen";
import type { WidgetPlacement } from "../types";

const SCREEN = { width: 1280, height: 800 };

describe("Grid", () => {
  beforeEach(() => setClock(new Date("2026-10-06T09:00:00")));
  afterEach(() => setClock(null));

  it("draws every widget in the sample screen", async () => {
    await render(<Grid widgets={sampleScreen} initialSize={SCREEN} />);

    expect(screen.getByText("Tuesday")).toBeTruthy();
    const photo = sampleScreen.find((w) => w.renderer === "photo")!;
    expect(screen.getByText((photo.data as { name: string }).name)).toBeTruthy();
    expect(screen.getByText("Chilly")).toBeTruthy();
    expect(screen.getByText("Yankees")).toBeTruthy();
  });

  it("positions each widget by its grid area", async () => {
    await render(<Grid widgets={sampleScreen} initialSize={SCREEN} />);

    const style = (id: string) => Object.assign({}, ...[screen.getByTestId(`widget-${id}`).props.style].flat(3));
    expect(style("weather")).toMatchObject({ left: 48, width: 272 });
    expect(style("sports")).toMatchObject({ left: 48 + 272 + 32, width: 272 });
    expect(style("photo")).toMatchObject({ left: 48 + 2 * (272 + 32), top: 48, width: 2 * 272 + 32 });
  });

  it("frames tiles but not the day and time or the photo", async () => {
    await render(<Grid widgets={sampleScreen} initialSize={SCREEN} />);

    const style = (id: string) => Object.assign({}, ...[screen.getByTestId(`widget-${id}`).props.style].flat(3));
    expect(style("weather").borderWidth).toBeGreaterThan(0);
    expect(style("orientation").borderWidth).toBeUndefined();
    expect(style("photo").borderWidth).toBeUndefined();
  });

  it("skips a widget whose renderer this app doesn't know", async () => {
    const widgets = [
      ...sampleScreen,
      { id: "future", renderer: "hologram", col: 0, row: 2, width: 1, height: 1, data: {} },
    ] as unknown as WidgetPlacement[];

    await render(<Grid widgets={widgets} initialSize={SCREEN} />);

    expect(screen.queryByTestId("widget-future")).toBeNull();
    expect(screen.getByText("Chilly")).toBeTruthy();
  });

  it("skips a widget that doesn't fit the grid", async () => {
    const widgets: WidgetPlacement[] = [
      { id: "offscreen", renderer: "tile", col: 3, row: 2, width: 2, height: 1, data: { icon: "sunny", word: "Warm" } },
    ];

    await render(<Grid widgets={widgets} initialSize={SCREEN} />);

    expect(screen.queryByText("Warm")).toBeNull();
  });

  it("draws nothing until it knows its size", async () => {
    await render(<Grid widgets={sampleScreen} />);

    expect(screen.queryByText("Chilly")).toBeNull();
  });
});
