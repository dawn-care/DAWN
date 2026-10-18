import { hideStaleWidgets, readScreen, STALE_AFTER_MS, type Screen } from "../screen";

/** The example from docs/api/screen-and-media.md, shortened. */
const response = () => ({
  carepartner_id: 1,
  generated_at: "2026-10-12T15:39:52-04:00",
  timezone: "America/New_York",
  today: "2026-10-12",
  awake_hours: { wake: "07:00", bed: "21:00", is_awake_now: true },
  widgets: [
    {
      id: 1, type: "photo", col: 0, row: 0, width: 2, height: 2, config: {},
      photo: { id: 1, display_name: "Sarah", url: "/media/1/file" },
      data: { caption: "Sarah" }, error: null,
    },
    {
      id: 3, type: "weather", col: 3, row: 0, width: 1, height: 1,
      config: { zip: "14617" }, photo: null, data: { temp: 48 }, error: null,
    },
    {
      id: 4, type: "news", col: 2, row: 1, width: 2, height: 1,
      config: {}, photo: null, data: null, error: "UPSTREAM_UNAVAILABLE",
    },
  ] as Record<string, unknown>[],
  schedule: [
    {
      scheduled_item_id: 1, date: "2026-10-12", time: "10:00", repeat: "daily", shape: "video",
      media: { id: 2, kind: "video", display_name: "Birthday", duration_seconds: 120, url: "/media/2/file" },
      background_photo: null, display_seconds: null, duration_seconds: 120, outside_awake_hours: false,
    },
  ] as Record<string, unknown>[],
});

describe("readScreen", () => {
  beforeEach(() => {
    jest.spyOn(console, "warn").mockImplementation(() => {});
  });

  afterEach(() => {
    jest.mocked(console.warn).mockRestore();
  });

  it("reads the awake hours, widgets with their places, and schedule", () => {
    expect(readScreen(response())).toEqual({
      awakeHours: { wake: "07:00", bed: "21:00" },
      widgets: [
        {
          id: 1, type: "photo", col: 0, row: 0, width: 2, height: 2, config: {},
          photo: { url: "/media/1/file", displayName: "Sarah" },
          data: { caption: "Sarah" }, error: null,
        },
        {
          id: 3, type: "weather", col: 3, row: 0, width: 1, height: 1,
          config: { zip: "14617" }, photo: null, data: { temp: 48 }, error: null,
        },
        {
          id: 4, type: "news", col: 2, row: 1, width: 2, height: 1,
          config: {}, photo: null, data: null, error: "UPSTREAM_UNAVAILABLE",
        },
      ],
      schedule: [
        {
          scheduledItemId: 1, date: "2026-10-12", time: "10:00", outsideAwakeHours: false,
          item: { kind: "video", url: "/media/2/file", title: "Birthday" },
        },
      ],
    });
  });

  it.each([
    ["no id", { type: "photo" }],
    ["no place", { id: 5, type: "photo", col: 0, row: 0, width: 2 }],
    ["a photo with no file", { id: 5, type: "photo", col: 0, row: 0, width: 2, height: 2, photo: { id: 1 } }],
  ])("drops a widget with %s and keeps the others", (_name, widget) => {
    const raw = response();
    raw.widgets[0] = widget;

    expect(readScreen(raw)!.widgets.map((w) => w.id)).toEqual([3, 4]);
    expect(console.warn).toHaveBeenCalledWith("Dropped 1 malformed widgets from the screen");
  });

  it("drops schedule entries the tablet can't play", () => {
    const raw = response();
    const entry = raw.schedule[0];
    raw.schedule.push(
      { ...entry, time: "10am" },
      { ...entry, media: { id: 3, kind: "slideshow", url: "/media/3/file" } },
      { ...entry, scheduled_item_id: undefined },
    );

    expect(readScreen(raw)!.schedule).toHaveLength(1);
    expect(console.warn).toHaveBeenCalledWith("Dropped 3 schedule entries the tablet can't play");
  });

  it("shows a scheduled audio item's background photo as its cover", () => {
    const raw = response();
    raw.schedule[0] = {
      ...raw.schedule[0], shape: "audio_with_background",
      media: { id: 9, kind: "audio", display_name: "Morning hymn", duration_seconds: 214, url: "/media/9/file" },
      background_photo: { id: 4, display_name: "Sarah", url: "/media/4/file" },
    };

    expect(readScreen(raw)!.schedule[0].item).toEqual({
      kind: "audio", url: "/media/9/file", title: "Morning hymn", coverUrl: "/media/4/file",
    });
  });

  it("shows a scheduled photo for its display time", () => {
    const raw = response();
    raw.schedule[0] = {
      ...raw.schedule[0], shape: "photo",
      media: { id: 4, kind: "photo", display_name: "Sarah", duration_seconds: null, url: "/media/4/file" },
      display_seconds: 45, duration_seconds: 45,
    };

    expect(readScreen(raw)!.schedule[0].item.durationSeconds).toBe(45);
  });

  it("marks plays outside the awake hours", () => {
    const raw = response();
    raw.schedule[0].outside_awake_hours = true;

    expect(readScreen(raw)!.schedule[0].outsideAwakeHours).toBe(true);
  });

  it.each([
    ["null", null],
    ["no widgets", { ...response(), widgets: undefined }],
    ["no schedule", { ...response(), schedule: null }],
    ["no awake hours", { ...response(), awake_hours: undefined }],
  ])("rejects a response with %s", (_name, raw) => {
    expect(readScreen(raw)).toBeNull();
  });
});

describe("hideStaleWidgets", () => {
  const screen = readScreen(response()) as Screen;
  const fetchedAt = new Date("2026-10-12T10:00:00");

  it("leaves a recently fetched screen alone", () => {
    expect(hideStaleWidgets(screen, fetchedAt, new Date(fetchedAt.getTime() + STALE_AFTER_MS - 1))).toBe(screen);
  });

  it("removes weather, sports and news widgets once the data is hours old", () => {
    const stale = hideStaleWidgets(screen, fetchedAt, new Date(fetchedAt.getTime() + STALE_AFTER_MS));

    expect(stale.widgets.map((widget) => widget.type)).toEqual(["photo"]);
    expect(stale.schedule).toBe(screen.schedule);
  });
});
