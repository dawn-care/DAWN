import { AppState, Platform, type AppStateStatus } from "react-native";
import { renderHook } from "@testing-library/react-native";
import * as NavigationBar from "expo-navigation-bar";
import { setStatusBarHidden } from "expo-status-bar";
import { useHiddenSystemBars } from "../useHiddenSystemBars";

jest.mock("expo-status-bar", () => ({ setStatusBarHidden: jest.fn() }));
jest.mock("expo-navigation-bar", () => ({ setVisibilityAsync: jest.fn(() => Promise.resolve()) }));

/** Starts the hook and returns a way to send app state changes to it. */
async function start() {
  let listener: (state: AppStateStatus) => void = () => {};
  const remove = jest.fn();
  jest.spyOn(AppState, "addEventListener").mockImplementation((_type, handler) => {
    listener = handler as (state: AppStateStatus) => void;
    return { remove };
  });
  const view = await renderHook(() => useHiddenSystemBars());
  return { changeTo: (state: AppStateStatus) => listener(state), view, remove };
}

describe("useHiddenSystemBars", () => {
  beforeEach(() => {
    jest.clearAllMocks();
    jest.replaceProperty(Platform, "OS", "android");
  });

  afterEach(() => {
    jest.restoreAllMocks();
  });

  it("hides the status and navigation bars from the start", async () => {
    await start();

    expect(setStatusBarHidden).toHaveBeenCalledWith(true, "none");
    expect(NavigationBar.setVisibilityAsync).toHaveBeenCalledWith("hidden");
  });

  it("hides them again when the app comes back from the background", async () => {
    const { changeTo } = await start();
    jest.clearAllMocks();

    changeTo("background");
    expect(NavigationBar.setVisibilityAsync).not.toHaveBeenCalled();
    changeTo("active");

    expect(setStatusBarHidden).toHaveBeenCalledWith(true, "none");
    expect(NavigationBar.setVisibilityAsync).toHaveBeenCalledWith("hidden");
  });

  it("stops listening when it unmounts", async () => {
    const { view, remove } = await start();

    await view.unmount();

    expect(remove).toHaveBeenCalled();
  });

  it("leaves the bars alone off Android", async () => {
    jest.replaceProperty(Platform, "OS", "web");
    await start();

    expect(setStatusBarHidden).not.toHaveBeenCalled();
    expect(NavigationBar.setVisibilityAsync).not.toHaveBeenCalled();
  });
});
