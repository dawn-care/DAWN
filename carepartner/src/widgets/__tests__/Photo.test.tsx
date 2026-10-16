import React from "react";
import { act, fireEvent, render, screen } from "@testing-library/react-native";
import Photo from "../renderers/Photo";

describe("Photo widget", () => {
  it("shows the photo and the person's name", async () => {
    await render(<Photo data={{ uri: "http://server/photo/file?id=1", name: "Sarah" }} />);

    expect(screen.getByTestId("photo-image").props.source).toEqual({ uri: "http://server/photo/file?id=1" });
    expect(screen.getByText("Sarah")).toBeTruthy();
  });

  it("shows a placeholder when there is no photo", async () => {
    await render(<Photo data={{ uri: null, name: "Sarah" }} />);

    expect(screen.getByTestId("photo-placeholder")).toBeTruthy();
    expect(screen.queryByTestId("photo-image")).toBeNull();
  });

  it("falls back to the placeholder if the photo fails to load", async () => {
    await render(<Photo data={{ uri: "http://server/photo/file?id=404" }} />);

    await act(async () => fireEvent(screen.getByTestId("photo-image"), "error"));

    expect(screen.getByTestId("photo-placeholder")).toBeTruthy();
  });

  it("leaves out the caption when there is no name", async () => {
    await render(<Photo data={{ uri: "http://server/photo/file?id=1" }} />);

    expect(screen.queryByText(/./)).toBeNull();
  });
});
