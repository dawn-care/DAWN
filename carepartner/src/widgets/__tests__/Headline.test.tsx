import React from "react";
import { render, screen } from "@testing-library/react-native";
import Headline from "../renderers/Headline";

describe("Headline widget", () => {
  it("shows the headline", async () => {
    await render(<Headline data={{ title: "Zoo welcomes a baby panda" }} />);

    expect(screen.getByText("Zoo welcomes a baby panda")).toBeTruthy();
  });
});
