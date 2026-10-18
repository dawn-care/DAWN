import { API_URL } from "../../api/config";
import { placementsFor } from "../placements";
import type { Screen, ScreenWidget } from "../screen";

const widget = (overrides: Partial<ScreenWidget>): ScreenWidget => ({
  id: 1, type: "weather", col: 0, row: 0, width: 1, height: 1,
  config: {}, photo: null, data: null, error: null,
  ...overrides,
});

const screenWith = (...widgets: ScreenWidget[]): Screen => ({
  awakeHours: { wake: "07:00", bed: "21:00" }, chime: true,
  widgets,
  schedule: [],
});

const placed = (...widgets: ScreenWidget[]) => placementsFor(screenWith(...widgets));

describe("placementsFor", () => {
  beforeEach(() => {
    jest.spyOn(console, "warn").mockImplementation(() => {});
  });

  afterEach(() => {
    jest.mocked(console.warn).mockRestore();
  });

  it("places each widget where the caregiver put it", () => {
    const [placement] = placed(widget({ id: 7, col: 3, row: 2, data: { temp: 48, condition: "Clouds" } }));

    expect(placement).toMatchObject({ id: "7", col: 3, row: 2, width: 1, height: 1 });
  });

  it("shows the weather as its temperature and a plain word", () => {
    expect(placed(widget({ data: { temp: 47.6, condition: "Clouds" } }))[0]).toMatchObject({
      renderer: "tile",
      data: { icon: "cloudy", word: "48°", label: "Cloudy" },
    });
  });

  it("shows weather it has no word for with a plain icon and no label", () => {
    expect(placed(widget({ data: { temp: 60, condition: "Tornado" } }))[0].data).toEqual({
      icon: "", word: "60°", label: null,
    });
  });

  it("shows a sports result with the score and the opponent", () => {
    const sports = widget({
      type: "sports", width: 2, height: 1,
      data: {
        league: "mlb", team_name: "New York Yankees", result: "lost",
        opponent_team_name: "Tampa Bay Rays", team_score: 3, opponent_score: 4,
      },
    });

    expect(placed(sports)[0]).toMatchObject({
      renderer: "wideTile",
      data: { icon: "baseball", label: "New York Yankees", word: "Lost 3–4", detail: "vs Tampa Bay Rays" },
    });
  });

  it("says when a team hasn't played yet this season", () => {
    const sports = widget({
      type: "sports", width: 2, height: 1,
      data: { league: "nba", team_name: "Boston Celtics", result: "upcoming", opponent_team_name: "", team_score: 0, opponent_score: 0 },
    });

    expect(placed(sports)[0].data).toEqual({ icon: "basketball", label: "Boston Celtics", word: "No games yet" });
  });

  it("shows a photo with its caption, or else its name", () => {
    const photo = { url: "/media/4/file", displayName: "sarah_beach" };
    const [captioned, named] = placed(
      widget({ type: "photo", width: 2, height: 2, photo, data: { caption: "Sarah" } }),
      widget({ id: 2, type: "photo", col: 2, width: 2, height: 2, photo, data: null }),
    );

    expect(captioned).toMatchObject({ renderer: "photo", data: { uri: `${API_URL}/media/4/file`, name: "Sarah" } });
    expect(named.data).toMatchObject({ name: "sarah_beach" });
  });

  it("shows a message with its role and photo", () => {
    const message = widget({
      type: "message", width: 2, height: 1,
      config: { text: "I'll visit at 3", label: "Nurse" },
      photo: { url: "/media/5/file", displayName: "Ana" },
    });

    expect(placed(message)[0]).toMatchObject({
      renderer: "message",
      data: { text: "I'll visit at 3", label: "Nurse", photoUri: `${API_URL}/media/5/file` },
    });
  });

  it("shows reminders in time order, with readable times", () => {
    const reminder = widget({
      type: "reminder", width: 2, height: 1,
      config: {
        items: [
          { time: "12:00", text: "Lunch", icon: "meal" },
          { time: "08:30", text: "Pills", icon: "medication" },
          { time: "noon", text: "Broken" },
        ],
      },
    });

    expect(placed(reminder)[0]).toMatchObject({
      renderer: "reminder",
      data: {
        items: [
          { minute: 8 * 60 + 30, timeLabel: "8:30 AM", text: "Pills", icon: "medication" },
          { minute: 12 * 60, timeLabel: "12 PM", text: "Lunch", icon: "meal" },
        ],
      },
    });
  });

  it("shows the first news headline", () => {
    const news = widget({ type: "news", width: 2, height: 1, data: { articles: [{ title: "Zoo welcomes a baby panda" }] } });

    expect(placed(news)[0]).toMatchObject({ renderer: "headline", data: { title: "Zoo welcomes a baby panda" } });
  });

  it("shows a day and time widget, with the date unless switched off", () => {
    const [withDate, withoutDate] = placed(
      widget({ type: "day_time", width: 2, height: 2 }),
      widget({ id: 2, type: "day_time", col: 2, width: 2, height: 2, config: { show_date: false } }),
    );

    expect(withDate).toMatchObject({ renderer: "orientation", data: { showDate: true } });
    expect(withoutDate.data).toEqual({ showDate: false });
  });

  it("leaves out a widget whose outside service is down, quietly", () => {
    expect(placed(widget({ error: "UPSTREAM_UNAVAILABLE" }))).toEqual([]);
    expect(console.warn).not.toHaveBeenCalled();
  });

  it.each([
    ["isn't its type's size", widget({ width: 2, height: 2, data: { temp: 48 } })],
    ["has an unknown type", widget({ type: "slideshow" })],
    ["runs off the grid", widget({ col: 0, row: 3, data: { temp: 48 } })],
    ["has no data", widget({ data: null })],
    ["has a photo widget with no photo", widget({ type: "photo", width: 2, height: 2 })],
    ["has a message with no text", widget({ type: "message", width: 2, height: 1, config: { text: " " } })],
    ["has no reminders it can read", widget({ type: "reminder", width: 2, height: 1, config: { items: [{ time: "9" }] } })],
    ["has no headlines", widget({ type: "news", width: 2, height: 1, data: { articles: [] } })],
    ["has a sports result it doesn't know", widget({ type: "sports", width: 2, height: 1, data: { team_name: "Yankees", result: "rained out", team_score: 1, opponent_score: 0 } })],
    ["has a sports result with no score", widget({ type: "sports", width: 2, height: 1, data: { team_name: "Yankees", result: "won" } })],
  ])("leaves out a widget that %s, and says how many", (_name, bad) => {
    const good = widget({ id: 9, col: 3, data: { temp: 48 } });

    expect(placed(bad, good).map((p) => p.id)).toEqual(["9"]);
    expect(console.warn).toHaveBeenCalledWith("Left out 1 widgets the tablet can't draw");
  });

  it("keeps the first of two widgets that overlap, and says how many it left out", () => {
    const first = widget({ id: 1, type: "photo", width: 2, height: 2, photo: { url: "/media/3/file", displayName: "Ana" } });
    const second = widget({ id: 2, type: "message", col: 1, row: 1, width: 2, height: 1, config: { text: "Hello" } });

    expect(placed(first, second).map((p) => p.id)).toEqual(["1"]);
    expect(console.warn).toHaveBeenCalledWith("Left out 1 widgets the tablet can't draw");
  });
});
