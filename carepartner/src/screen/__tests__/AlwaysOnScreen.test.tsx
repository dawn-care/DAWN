import React from "react";
import { render, screen } from "@testing-library/react-native";
import { setClock } from "../../clock/clock";
import AlwaysOnScreen from "../AlwaysOnScreen";
import type { Screen } from "../screen";
import { useScreen } from "../ScreenProvider";

jest.mock("../ScreenProvider", () => ({ useScreen: jest.fn() }));
jest.mock("../../widgets/Grid", () => {
  const { Text } = jest.requireActual("react-native");
  return {
    __esModule: true,
    default: ({ widgets }: { widgets: { id: string }[] }) => <Text>{`grid:${widgets.map((w) => w.id).join(",")}`}</Text>,
  };
});

const showing = (value: Screen | null) => jest.mocked(useScreen).mockReturnValue({ screen: value, refresh: jest.fn(), fetchedAt: null, failed: false });

const screenWith = (widgets: Screen["widgets"]): Screen => ({ awakeHours: { wake: "07:00", bed: "21:00" }, chime: true, widgets, schedule: [] });

describe("AlwaysOnScreen", () => {
  beforeEach(() => {
    setClock(new Date("2026-10-12T09:00:00"));
    jest.spyOn(console, "warn").mockImplementation(() => {});
  });

  afterEach(() => {
    setClock(null);
    jest.restoreAllMocks();
  });

  it("draws the day and time and the screen's widgets on the grid", async () => {
    showing({
      awakeHours: { wake: "07:00", bed: "21:00" }, chime: true,
      widgets: [
        {
          id: 3, type: "weather", col: 3, row: 0, width: 1, height: 1,
          config: {}, photo: null, data: { temp: 48 }, error: null,
        },
      ],
      schedule: [],
    });
    await render(<AlwaysOnScreen />);

    expect(screen.getByText("grid:day_time,3")).toBeTruthy();
  });

  it.each([
    ["with no widgets", screenWith([])],
    [
      "when it can't draw any of the widgets",
      screenWith([{ id: 3, type: "weather", col: 2, row: 0, width: 2, height: 2, config: {}, photo: null, data: { temp: 48 }, error: null }]),
    ],
  ])("still draws the day and time on the grid %s", async (_name, value) => {
    showing(value);
    await render(<AlwaysOnScreen />);

    expect(screen.getByText("grid:day_time")).toBeTruthy();
    expect(screen.queryByTestId("default-screen")).toBeNull();
  });

  it("shows the day, time of day and date before the first screen arrives", async () => {
    showing(null);
    await render(<AlwaysOnScreen />);

    expect(screen.getByTestId("default-screen")).toBeTruthy();
    expect(screen.getByText("Monday")).toBeTruthy();
    expect(screen.getByText("Morning")).toBeTruthy();
    expect(screen.getByText("October 12")).toBeTruthy();
    expect(screen.queryByText(/^grid:/)).toBeNull();
  });
});
