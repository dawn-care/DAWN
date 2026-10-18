import { gestureOf, LONG_PRESS_MS, SWIPE_DISTANCE } from "../gestures";

const at = (x: number, y: number, ms: number) => ({ x, y, at: ms });

describe("gestureOf", () => {
  it.each([
    ["a quick touch in place", at(100, 100, LONG_PRESS_MS - 1), "tap"],
    ["a small wobble", at(110, 105, 200), "tap"],
    ["a touch held in place", at(100, 100, LONG_PRESS_MS), "long_press"],
    ["a quick drag", at(100 + SWIPE_DISTANCE, 100, 150), "swipe"],
    ["a slow drag", at(100, 100 + SWIPE_DISTANCE, 2_000), "swipe"],
  ])("calls %s a %s", (_name, end, gesture) => {
    expect(gestureOf(at(100, 100, 0), end)).toBe(gesture);
  });
});
