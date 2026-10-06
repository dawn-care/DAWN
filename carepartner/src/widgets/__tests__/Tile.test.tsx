import React from "react";
import { render, screen } from "@testing-library/react-native";
import Tile from "../renderers/Tile";
import { FALLBACK_ICON, TILE_ICONS, tileIcon } from "../renderers/icons";

describe("Tile widget", () => {
  it("shows the icon, label and word", async () => {
    await render(<Tile data={{ icon: "baseball", word: "Won", label: "Yankees" }} />);

    expect(screen.getByTestId("tile-icon")).toBeTruthy();
    expect(screen.getByText("Yankees")).toBeTruthy();
    expect(screen.getByText("Won")).toBeTruthy();
  });

  it("leaves out the label when there is none", async () => {
    await render(<Tile data={{ icon: "sunny", word: "Warm" }} />);

    expect(screen.getByText("Warm")).toBeTruthy();
    expect(screen.queryByText("Yankees")).toBeNull();
  });
});

describe("tile icons", () => {
  it("maps known keys and falls back for unknown ones", () => {
    expect(tileIcon("rain")).toBe(TILE_ICONS.rain);
    expect(tileIcon("not-a-real-icon")).toBe(FALLBACK_ICON);
  });
});
