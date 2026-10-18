import React from "react";
import { render, screen } from "@testing-library/react-native";
import WideTile from "../renderers/WideTile";

describe("WideTile widget", () => {
  it("shows the label, the word and the detail", async () => {
    await render(<WideTile data={{ icon: "baseball", label: "New York Yankees", word: "Lost 3–4", detail: "vs Tampa Bay Rays" }} />);

    expect(screen.getByText("New York Yankees")).toBeTruthy();
    expect(screen.getByText("Lost 3–4")).toBeTruthy();
    expect(screen.getByText("vs Tampa Bay Rays")).toBeTruthy();
  });

  it("leaves out the detail line when there is none", async () => {
    await render(<WideTile data={{ icon: "basketball", label: "Boston Celtics", word: "No games yet" }} />);

    expect(screen.getByText("No games yet")).toBeTruthy();
    expect(screen.queryByText(/^vs /)).toBeNull();
  });
});
