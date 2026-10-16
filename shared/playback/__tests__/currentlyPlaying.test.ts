import { currentlyPlaying, overlappingItems } from "../currentlyPlaying";
import type { PlayableItem } from "../currentlyPlaying";

const minute = (hours: number, minutes = 0) => hours * 60 + minutes;
const at = (time: string) => new Date(`2026-10-12T${time}`);

// Awake from 8 am until 10 pm, the default.
const video: PlayableItem = {
  kind: "video",
  date: "2026-10-12",
  startMinute: minute(10),
  durationSeconds: 120,
};
const audio: PlayableItem = {
  kind: "audio",
  date: "2026-10-12",
  startMinute: minute(15),
  durationSeconds: 192,
};
const audioWithPhoto: PlayableItem = {
  kind: "audioWithPhoto",
  date: "2026-10-12",
  startMinute: minute(18),
  durationSeconds: 35,
};
const photo: PlayableItem = {
  kind: "photo",
  date: "2026-10-12",
  startMinute: minute(12),
  durationSeconds: 30,
};

describe("with nothing scheduled", () => {
  it.each([
    ["08:00:00", "static"],
    ["12:00:00", "static"],
    ["21:59:59", "static"],
    ["22:00:00", "night"],
    ["23:59:59", "night"],
    ["00:00:00", "night"],
    ["07:59:59", "night"],
  ])("at %s the screen is %s", (time, expected) => {
    expect(currentlyPlaying(at(time), []).kind).toBe(expected);
  });

  it("uses the awake hours it is given", () => {
    const early = { wakeMinute: minute(6), bedMinute: minute(19) };
    expect(currentlyPlaying(at("06:00:00"), [], early).kind).toBe("static");
    expect(currentlyPlaying(at("19:00:00"), [], early).kind).toBe("night");
    expect(currentlyPlaying(at("21:00:00"), [], early).kind).toBe("night");
  });
});

describe("a video", () => {
  const schedule = [video];

  it("is not playing the second before it starts", () => {
    expect(currentlyPlaying(at("09:59:59"), schedule).kind).toBe("static");
  });

  it("starts exactly at its start time", () => {
    const result = currentlyPlaying(at("10:00:00"), schedule);
    expect(result).toEqual({ kind: "video", item: video, elapsedSeconds: 0, remainingSeconds: 120 });
  });

  it("reports how far in it is", () => {
    const result = currentlyPlaying(at("10:01:30"), schedule);
    expect(result).toMatchObject({ kind: "video", elapsedSeconds: 90, remainingSeconds: 30 });
  });

  it("plays through its last second", () => {
    const result = currentlyPlaying(at("10:01:59"), schedule);
    expect(result).toMatchObject({ kind: "video", elapsedSeconds: 119, remainingSeconds: 1 });
  });

  it("is over at exactly its length, and her usual screen comes back", () => {
    expect(currentlyPlaying(at("10:02:00"), schedule).kind).toBe("static");
  });

  it("returns the item it was given", () => {
    const result = currentlyPlaying(at("10:00:30"), schedule);
    expect(result.kind === "video" && result.item).toBe(video);
  });
});

describe("the other kinds", () => {
  it("plays audio over her usual screen", () => {
    expect(currentlyPlaying(at("15:00:10"), [audio])).toMatchObject({ kind: "audio", item: audio });
  });

  it("plays audio over a chosen photo", () => {
    expect(currentlyPlaying(at("18:00:01"), [audioWithPhoto])).toMatchObject({
      kind: "audioWithPhoto",
      item: audioWithPhoto,
    });
  });

  it("shows a timed photo for its set time", () => {
    expect(currentlyPlaying(at("12:00:29"), [photo])).toMatchObject({ kind: "photo", remainingSeconds: 1 });
    expect(currentlyPlaying(at("12:00:30"), [photo]).kind).toBe("static");
  });

  it("finds the right item among several", () => {
    const schedule = [video, audio, audioWithPhoto, photo];
    expect(currentlyPlaying(at("10:00:30"), schedule)).toMatchObject({ kind: "video" });
    expect(currentlyPlaying(at("12:00:05"), schedule)).toMatchObject({ kind: "photo" });
    expect(currentlyPlaying(at("15:01:00"), schedule)).toMatchObject({ kind: "audio" });
    expect(currentlyPlaying(at("18:00:34"), schedule)).toMatchObject({ kind: "audioWithPhoto" });
    expect(currentlyPlaying(at("11:00:00"), schedule).kind).toBe("static");
  });
});

describe("days", () => {
  it("does not play an item scheduled for tomorrow", () => {
    const tomorrow = { ...video, date: "2026-10-13" };
    expect(currentlyPlaying(at("10:00:30"), [tomorrow]).kind).toBe("static");
  });

  it("does not play an item scheduled for yesterday", () => {
    const yesterday = { ...video, date: "2026-10-11" };
    expect(currentlyPlaying(at("10:00:30"), [yesterday]).kind).toBe("static");
  });

  it("plays today's item from a schedule that covers today and tomorrow", () => {
    const tomorrow = { ...video, date: "2026-10-13" };
    expect(currentlyPlaying(at("10:00:30"), [tomorrow, video]).kind).toBe("video");
  });

  it("finds an item on a month boundary", () => {
    const item = { ...video, date: "2026-11-01" };
    expect(currentlyPlaying(new Date("2026-11-01T10:00:30"), [item]).kind).toBe("video");
    expect(currentlyPlaying(new Date("2026-10-31T10:00:30"), [item]).kind).toBe("static");
  });
});

describe("bedtime", () => {
  const lateVideo: PlayableItem = {
    kind: "video",
    date: "2026-10-12",
    startMinute: minute(21, 55),
    durationSeconds: 600,
  };

  it("lets an item that started awake finish after bed", () => {
    expect(currentlyPlaying(at("22:00:00"), [lateVideo])).toMatchObject({
      kind: "video",
      elapsedSeconds: 300,
      remainingSeconds: 300,
    });
    expect(currentlyPlaying(at("22:04:59"), [lateVideo]).kind).toBe("video");
  });

  it("shows the night screen once that item is over", () => {
    expect(currentlyPlaying(at("22:05:00"), [lateVideo]).kind).toBe("night");
  });

  it("keeps playing across midnight", () => {
    const hours = { wakeMinute: minute(8), bedMinute: minute(23, 59) };
    const nearMidnight: PlayableItem = {
      kind: "audio",
      date: "2026-10-12",
      startMinute: minute(23, 55),
      durationSeconds: 600,
    };
    const after = new Date("2026-10-13T00:03:00");
    expect(currentlyPlaying(after, [nearMidnight], hours)).toMatchObject({
      kind: "audio",
      elapsedSeconds: 480,
      remainingSeconds: 120,
    });
    expect(currentlyPlaying(new Date("2026-10-13T00:05:00"), [nearMidnight], hours).kind).toBe("night");
  });
});

describe("items that start outside the awake hours", () => {
  it("never plays one that starts before wake up, even when it runs into the awake hours", () => {
    const early: PlayableItem = { ...video, startMinute: minute(7, 55), durationSeconds: 600 };
    expect(currentlyPlaying(at("07:56:00"), [early]).kind).toBe("night");
    expect(currentlyPlaying(at("08:02:00"), [early]).kind).toBe("static");
  });

  it("never plays one that starts at bed time", () => {
    const atBed: PlayableItem = { ...video, startMinute: minute(22), durationSeconds: 120 };
    expect(currentlyPlaying(at("22:00:30"), [atBed]).kind).toBe("night");
  });

  it("plays one that starts at wake up", () => {
    const atWake: PlayableItem = { ...video, startMinute: minute(8), durationSeconds: 120 };
    expect(currentlyPlaying(at("08:00:30"), [atWake]).kind).toBe("video");
  });

  it("follows the awake hours it is given", () => {
    const hours = { wakeMinute: minute(6), bedMinute: minute(19) };
    const evening: PlayableItem = { ...video, startMinute: minute(20), durationSeconds: 120 };
    expect(currentlyPlaying(at("20:00:30"), [evening], hours).kind).toBe("night");
    expect(currentlyPlaying(at("20:00:30"), [evening]).kind).toBe("video");
  });
});

// The caregiver refuses to schedule overlapping items (see overlappingItems), so this is a
// fallback for a schedule that has one anyway. The tablet instead finishes the item it is on.
describe("more than one item under way", () => {
  it("falls back to the one that started most recently", () => {
    const first: PlayableItem = { ...video, startMinute: minute(10), durationSeconds: 600 };
    const second: PlayableItem = { ...audio, startMinute: minute(10, 5), durationSeconds: 600 };
    expect(currentlyPlaying(at("10:06:00"), [first, second])).toMatchObject({ kind: "audio" });
    expect(currentlyPlaying(at("10:06:00"), [second, first])).toMatchObject({ kind: "audio" });
  });

  it("shows the first listed when two start together", () => {
    const first: PlayableItem = { ...video, startMinute: minute(10) };
    const second: PlayableItem = { ...audio, startMinute: minute(10) };
    expect(currentlyPlaying(at("10:00:30"), [first, second])).toMatchObject({ kind: "video" });
  });
});

describe("items it cannot use", () => {
  it.each([
    ["a date that is not a date", { date: "tomorrow" }],
    ["a date that does not exist", { date: "2026-02-30" }],
    ["a date with no zero padding", { date: "2026-10-2" }],
    ["a start before midnight", { startMinute: -5 }],
    ["a start after the day", { startMinute: 24 * 60 }],
    ["a start that is not whole", { startMinute: 600.5 }],
    ["a length of zero", { durationSeconds: 0 }],
    ["a negative length", { durationSeconds: -30 }],
    ["a length that is not a number", { durationSeconds: Number.NaN }],
  ])("skips an item with %s", (_name, change) => {
    const bad = { ...video, ...change };
    expect(currentlyPlaying(at("10:00:30"), [bad]).kind).toBe("static");
  });

  it("still plays a good item next to a bad one", () => {
    const bad = { ...video, date: "nope" };
    expect(currentlyPlaying(at("10:00:30"), [bad, video]).kind).toBe("video");
  });
});

describe("time is only read from its arguments", () => {
  it("gives the same answer every time", () => {
    const schedule = [video];
    const first = currentlyPlaying(at("10:00:30"), schedule);
    expect(currentlyPlaying(at("10:00:30"), schedule)).toEqual(first);
  });

  it("does not change the schedule it is given", () => {
    const schedule = [{ ...video }, { ...audio }];
    const copy = schedule.map((item) => ({ ...item }));
    currentlyPlaying(at("10:00:30"), schedule);
    expect(schedule).toEqual(copy);
  });

  it("carries extra fields on the item through", () => {
    const rich = { ...video, id: 7, title: "Good morning, Mom", url: "/media/2/file" };
    const result = currentlyPlaying(at("10:00:30"), [rich]);
    expect(result.kind === "video" && result.item.title).toBe("Good morning, Mom");
  });
});

describe("overlappingItems", () => {
  // The video above runs 10:00:00 to 10:02:00.
  const candidate = (startMinute: number, durationSeconds: number): PlayableItem => ({
    ...audio,
    startMinute,
    durationSeconds,
  });

  it("finds an item that starts while another is playing", () => {
    expect(overlappingItems(candidate(minute(10, 1), 60), [video, photo])).toEqual([video]);
  });

  it("finds an item that starts before another and runs into it", () => {
    expect(overlappingItems(candidate(minute(9, 59), 180), [video])).toEqual([video]);
  });

  it("finds an item that sits entirely inside another", () => {
    expect(overlappingItems(candidate(minute(10, 1), 30), [video])).toEqual([video]);
  });

  it("does not count an item that starts the moment another ends", () => {
    expect(overlappingItems(candidate(minute(10, 2), 60), [video])).toEqual([]);
  });

  it("does not count an item that ends the moment another starts", () => {
    expect(overlappingItems(candidate(minute(9, 58), 120), [video])).toEqual([]);
  });

  it("does not count an item on another day", () => {
    expect(overlappingItems(candidate(minute(10, 1), 60), [{ ...video, date: "2026-10-13" }])).toEqual([]);
  });

  it("finds every item in the way", () => {
    const other: PlayableItem = { ...photo, startMinute: minute(10, 1), durationSeconds: 30 };
    expect(overlappingItems(candidate(minute(10, 1), 60), [video, other, photo])).toEqual([video, other]);
  });

  it("finds nothing for a candidate that would never play", () => {
    const early = candidate(minute(7), 600);
    expect(overlappingItems(early, [{ ...video, startMinute: minute(7), durationSeconds: 3600 }])).toEqual([]);
  });

  it("ignores scheduled items that would never play", () => {
    const unusable: PlayableItem = { ...video, durationSeconds: 0 };
    expect(overlappingItems(candidate(minute(10), 60), [unusable])).toEqual([]);
  });
});
