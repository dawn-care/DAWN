import React from "react";
import { act, fireEvent, render, screen } from "@testing-library/react-native";
import Message from "../renderers/Message";

describe("Message widget", () => {
  it("shows the message, its role and photo", async () => {
    await render(<Message data={{ text: "I'll visit at 3", label: "Nurse", photoUri: "http://x/5.jpg" }} />);

    expect(screen.getByText("I'll visit at 3")).toBeTruthy();
    expect(screen.getByText("Nurse")).toBeTruthy();
    expect(screen.getByTestId("message-photo")).toBeTruthy();
  });

  it("keeps the text hidden until it has been measured, so it never shrinks in view", async () => {
    await render(<Message data={{ text: "Love you" }} />);
    expect(screen.getByText("Love you")).toHaveStyle({ opacity: 0 });

    const layout = { nativeEvent: { layout: { width: 343, height: 167 } } };
    await act(async () => fireEvent(screen.getByTestId("message-body"), "layout", layout));

    expect(screen.getByText("Love you")).toHaveStyle({ opacity: 1 });
  });

  it("shows just the message when it has no role or photo", async () => {
    await render(<Message data={{ text: "Love you" }} />);

    expect(screen.getByText("Love you")).toBeTruthy();
    expect(screen.queryByTestId("message-photo")).toBeNull();
  });
});
