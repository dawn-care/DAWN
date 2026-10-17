import React from "react";
import { render, screen } from "@testing-library/react-native";
import Message from "../renderers/Message";

describe("Message widget", () => {
  it("shows the message, its role and photo", async () => {
    await render(<Message data={{ text: "I'll visit at 3", label: "Nurse", photoUri: "http://x/5.jpg" }} />);

    expect(screen.getByText("I'll visit at 3")).toBeTruthy();
    expect(screen.getByText("Nurse")).toBeTruthy();
    expect(screen.getByTestId("message-photo")).toBeTruthy();
  });

  it("shows just the message when it has no role or photo", async () => {
    await render(<Message data={{ text: "Love you" }} />);

    expect(screen.getByText("Love you")).toBeTruthy();
    expect(screen.queryByTestId("message-photo")).toBeNull();
  });
});
