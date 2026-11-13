/**
 * recording.ts
 *
 * User: Caregiver
 *
 * Browser helpers for recording on the media page, with no React: whether
 * this browser can record, opening the camera or microphone (with caregiver
 * wording when it can't), which file type to record, and turning a recording
 * or a camera frame into a File that the save step uploads like a picked file.
 */

import { timeLabel } from "@shared";
import { RECORDING_MESSAGES, RECORDING_TYPES, RECORDING_UNSUPPORTED, SNAPSHOT_FAILED } from "./constants";
import type { RecordingSource } from "./constants";

/** Thrown with a message that can be shown to the caregiver as it is */
export class RecordingError extends Error {
  constructor(message: string) {
    super(message);
    this.name = "RecordingError";
  }
}

/**
 * Whether this browser can open the camera and microphone and record them.
 * Browsers only allow it on https or localhost (a secure context).
 */
export function mediaSupport(): boolean {
  return (
    window.isSecureContext &&
    typeof navigator.mediaDevices?.getUserMedia === "function" &&
    typeof MediaRecorder !== "undefined"
  );
}

// getUserMedia errors that mean no device could be used
const NOT_FOUND = ["NotFoundError", "OverconstrainedError"];

function messageFor(err: unknown, source: RecordingSource): string {
  const messages = RECORDING_MESSAGES[source];
  const name = err instanceof DOMException ? err.name : "";
  if (name === "NotAllowedError" || name === "SecurityError") return messages.denied;
  if (NOT_FOUND.includes(name)) return messages.notFound;
  if (name === "NotReadableError" || name === "AbortError") return messages.inUse;
  return messages.failed;
}

/**
 * Opens the camera (front-facing where there is a choice, with sound) or the
 * microphone. A camera without a microphone still opens, without sound, so
 * photos and silent videos still work. Throws a RecordingError worded for the
 * caregiver. The caller must stopStream the result when done with it.
 */
export async function openStream(source: RecordingSource): Promise<MediaStream> {
  if (!mediaSupport()) throw new RecordingError(RECORDING_UNSUPPORTED);

  const video: MediaTrackConstraints = { facingMode: "user", width: { ideal: 1280 }, height: { ideal: 720 } };
  try {
    if (source === "voice") return await navigator.mediaDevices.getUserMedia({ audio: true });
    try {
      return await navigator.mediaDevices.getUserMedia({ video, audio: true });
    } catch (err) {
      // Most likely no microphone; try the camera on its own
      if (err instanceof DOMException && NOT_FOUND.includes(err.name)) {
        return await navigator.mediaDevices.getUserMedia({ video });
      }
      throw err;
    }
  } catch (err) {
    console.error(`Could not open the ${source === "camera" ? "camera" : "microphone"}`, err);
    throw new RecordingError(messageFor(err, source));
  }
}

/** Turns the camera or microphone off (its light goes out). Safe to call with null. */
export function stopStream(stream: MediaStream | null) {
  stream?.getTracks().forEach((track) => track.stop());
}

/**
 * The first type in RECORDING_TYPES this browser can record, or null to let
 * the browser choose (read the type it chose from the recorder afterwards).
 */
export function pickRecordingType(kind: "video" | "audio"): string | null {
  return RECORDING_TYPES[kind].find((type) => MediaRecorder.isTypeSupported(type)) ?? null;
}

/**
 * The file extension for a recorded type. The server only takes an MP4 as
 * audio when it is named .m4a.
 */
function extensionFor(mimeType: string, kind: "video" | "audio"): string {
  const base = mimeType.split(";")[0].trim().toLowerCase();
  if (base.endsWith("/mp4")) return kind === "audio" ? ".m4a" : ".mp4";
  if (base.endsWith("/ogg")) return ".ogg";
  return ".webm";
}

// "Nov 13, 3:05 pm", in the app's wording for times
function whenLabel(date: Date): string {
  const day = date.toLocaleDateString("en-US", { month: "short", day: "numeric" });
  return `${day}, ${timeLabel(`${date.getHours()}:${date.getMinutes()}`)}`;
}

/**
 * A recording as a File, named for when it was made ("Video Nov 13, 3:05 pm.webm").
 * The save step suggests that name, without the extension, as the item's name.
 */
export function recordingFile(blob: Blob, kind: "video" | "audio", mimeType: string, now = new Date()): File {
  const type = mimeType.split(";")[0].trim() || blob.type.split(";")[0];
  const label = kind === "video" ? "Video" : "Voice recording";
  return new File([blob], `${label} ${whenLabel(now)}${extensionFor(type, kind)}`, { type });
}

/**
 * The frame the camera shows right now, as a JPEG File ("Photo Nov 13, 3:05 pm.jpg").
 * It is the camera's own picture, not mirrored like a selfie preview.
 */
export function snapshotFile(video: HTMLVideoElement, now = new Date()): Promise<File> {
  return new Promise((resolve, reject) => {
    const { videoWidth: width, videoHeight: height } = video;
    const canvas = document.createElement("canvas");
    const context = canvas.getContext("2d");
    // No size means the camera hasn't shown a frame yet
    if (width === 0 || height === 0 || !context) {
      reject(new RecordingError(SNAPSHOT_FAILED));
      return;
    }
    canvas.width = width;
    canvas.height = height;
    context.drawImage(video, 0, 0, width, height);
    canvas.toBlob(
      (blob) => {
        if (!blob) reject(new RecordingError(SNAPSHOT_FAILED));
        else resolve(new File([blob], `Photo ${whenLabel(now)}.jpg`, { type: "image/jpeg" }));
      },
      "image/jpeg",
      0.9
    );
  });
}
