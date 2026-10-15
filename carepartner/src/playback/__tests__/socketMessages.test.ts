import { readServerMessage, statusMessage } from "../socketMessages";

const read = (value: unknown) => readServerMessage(JSON.stringify(value));

describe("readServerMessage", () => {
  it("reads a play command", () => {
    expect(read({ type: "play", playback_id: 12, module_id: 3, expires_at: "2026-10-12T20:00:30Z" })).toEqual({
      type: "play",
      playbackId: 12,
      moduleId: 3,
      expiresAt: new Date("2026-10-12T20:00:30Z"),
    });
  });

  it.each([
    ["no playback id", { type: "play", module_id: 3, expires_at: "2026-10-12T20:00:30Z" }],
    ["no module id", { type: "play", playback_id: 12, expires_at: "2026-10-12T20:00:30Z" }],
    ["a text id", { type: "play", playback_id: "12", module_id: 3, expires_at: "2026-10-12T20:00:30Z" }],
    ["a zero id", { type: "play", playback_id: 0, module_id: 3, expires_at: "2026-10-12T20:00:30Z" }],
    ["no deadline", { type: "play", playback_id: 12, module_id: 3 }],
    ["an unreadable deadline", { type: "play", playback_id: 12, module_id: 3, expires_at: "soon" }],
  ])("drops a play command with %s", (_name, message) => {
    expect(read(message)).toBeNull();
  });

  it("reads a stop command, and drops one without a playback id", () => {
    expect(read({ type: "stop", playback_id: 12 })).toEqual({ type: "stop", playbackId: 12 });
    expect(read({ type: "stop" })).toBeNull();
  });

  it("reads a refresh nudge", () => {
    expect(read({ type: "refresh" })).toEqual({ type: "refresh" });
  });

  it("reads a status result with a status or an error", () => {
    expect(read({ type: "status_result", playback_id: 12, status: "expired" })).toEqual({
      type: "status_result",
      playbackId: 12,
      status: "expired",
      error: undefined,
    });
    expect(read({ type: "status_result", playback_id: 12, error: "playback not found" })).toMatchObject({
      error: "playback not found",
    });
    expect(read({ type: "status_result" })).toBeNull();
  });

  it("reads an error", () => {
    expect(read({ type: "error", error: "unknown message type" })).toEqual({
      type: "error",
      error: "unknown message type",
    });
  });

  it.each([
    ["an unknown type", JSON.stringify({ type: "dance" })],
    ["a list", "[]"],
    ["null", "null"],
    ["text that isn't JSON", "hello"],
  ])("drops %s", (_name, raw) => {
    expect(readServerMessage(raw)).toBeNull();
  });
});

describe("statusMessage", () => {
  it("reports a play's status in the server's format", () => {
    expect(JSON.parse(statusMessage(12, "started"))).toEqual({ type: "status", playback_id: 12, status: "started" });
  });
});
