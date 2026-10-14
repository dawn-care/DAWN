import { doesntFitNote, hiddenWidgets, shownWidgets, widgetName } from "../screenWidgets";
import type { LayoutPreset } from "../layoutPresets";

const simple: LayoutPreset = { name: "simple", spots: 2 };
const balanced: LayoutPreset = { name: "balanced", spots: 4 };
const grid: LayoutPreset = { name: "grid", spots: 4 };
const busier: LayoutPreset = { name: "busier", spots: 6 };


describe("widgetName", () => {
  it.each([
    ["photo", "Photo"],
    ["message", "Message"],
    ["reminder", "Reminder"],
    ["weather", "Weather"],
    ["sports", "Sports"],
    ["news", "News"],
  ] as const)("calls a %s widget %s", (type, name) => {
    expect(widgetName({ type })).toBe(name);
  });

  it("uses the team's name for a sports widget with a team", () => {
    expect(widgetName({ type: "sports", teamName: "Buffalo Bills" })).toBe("Buffalo Bills");
  });

  it("ignores a team name on any other widget", () => {
    expect(widgetName({ type: "news", teamName: "Buffalo Bills" })).toBe("News");
  });

  it("falls back to Sports when the team name is empty", () => {
    expect(widgetName({ type: "sports", teamName: "" })).toBe("Sports");
  });
});

describe("shownWidgets and hiddenWidgets", () => {
  const widgets = [
    { id: "a", spot: 0 },
    { id: "b", spot: 1 },
    { id: "c", spot: 3 },
    { id: "d", spot: 5 },
  ];

  it("splits widgets by whether their spot is in the layout", () => {
    expect(shownWidgets(widgets, simple).map((w) => w.id)).toEqual(["a", "b"]);
    expect(hiddenWidgets(widgets, simple).map((w) => w.id)).toEqual(["c", "d"]);
  });

  it("accounts for every widget exactly once in any layout", () => {
    for (const preset of [simple, balanced, grid, busier]) {
      const total = shownWidgets(widgets, preset).length + hiddenWidgets(widgets, preset).length;
      expect(total).toBe(widgets.length);
    }
  });

  it("brings hidden widgets back when the layout grows again", () => {
    expect(hiddenWidgets(widgets, simple)).toHaveLength(2);
    expect(hiddenWidgets(widgets, busier)).toEqual([]);
    expect(shownWidgets(widgets, busier)).toEqual(widgets);
  });

  it("does not change the widgets it is given", () => {
    const copy = widgets.map((w) => ({ ...w }));
    shownWidgets(widgets, simple);
    hiddenWidgets(widgets, simple);
    expect(widgets).toEqual(copy);
  });

  it("treats a bad spot as hidden", () => {
    expect(hiddenWidgets([{ spot: -1 }, { spot: 0.5 }], busier)).toHaveLength(2);
  });
});

describe("doesntFitNote", () => {
  const widget = (spot: number, name: string) => ({ spot, name });

  it("is empty when everything fits", () => {
    expect(doesntFitNote([widget(0, "Photo"), widget(1, "Weather")], simple)).toBe("");
    expect(doesntFitNote([], simple)).toBe("");
  });

  it("names one widget", () => {
    expect(doesntFitNote([widget(0, "Photo"), widget(2, "Reminder")], simple)).toBe(
      "Reminder doesn't fit in Simple. It's kept, and comes back if you switch.",
    );
  });

  it("names two widgets", () => {
    expect(doesntFitNote([widget(2, "Reminder"), widget(3, "Buffalo Bills")], simple)).toBe(
      "Reminder and Buffalo Bills don't fit in Simple. They're kept, and come back if you switch.",
    );
  });

  it("names three or more with commas and an and", () => {
    const widgets = [widget(2, "Reminder"), widget(3, "Weather"), widget(4, "News")];
    expect(doesntFitNote(widgets, simple)).toBe(
      "Reminder, Weather and News don't fit in Simple. They're kept, and come back if you switch.",
    );
  });

  it("lists widgets in the order given, whatever their spots", () => {
    expect(doesntFitNote([widget(5, "News"), widget(2, "Weather")], simple)).toContain(
      "News and Weather don't fit",
    );
  });

  it("names the layout being chosen", () => {
    expect(doesntFitNote([widget(4, "News")], balanced)).toContain("fit in Balanced.");
    expect(doesntFitNote([widget(4, "News")], grid)).toContain("fit in Grid.");
  });

  it("has nothing to say once the layout is big enough", () => {
    const widgets = [widget(0, "Photo"), widget(4, "News"), widget(5, "Weather")];
    expect(doesntFitNote(widgets, busier)).toBe("");
  });
});
