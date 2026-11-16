/**
 * files.ts
 *
 * User: Caregiver
 *
 * Checks the files a caregiver picked before anything is uploaded: is it a
 * photo, video or audio file the server accepts, is it small enough, and is
 * a video or audio file no longer than 10 minutes. Each file that passes
 * becomes a PendingFile; each that fails gets a friendly message naming it.
 * A photo or recording made on the page becomes one through
 * pendingFromRecording instead.
 *
 * These are early checks for a better message. The server checks again and
 * is the authority, so a file the browser can't measure is let through.
 */

import { probeMedia } from "./api";
import {
  ACCEPTED_TYPES,
  MAX_DURATION_SECONDS,
  MAX_FILE_BYTES,
  MIN_RECORDING_SECONDS,
  RECORDING_TOO_SHORT,
} from "./constants";
import type { MediaKind, RecordingSource } from "./constants";

/** A file that passed the checks and is waiting to be named and saved */
export interface PendingFile {
  id: number;
  file: File;
  kind: MediaKind;
  /**
   * Length in seconds of video or audio: for a picked file, what the browser
   * read from it (null if it couldn't); for a recording, how long the recorder
   * ran, since a browser recording may not store its length. Null for photos.
   */
  seconds: number | null;
  /** Set when it was recorded on this page rather than picked: the panel it came from, which Retake goes back to */
  recordedWith?: RecordingSource;
}

export interface CheckedFiles {
  accepted: PendingFile[];
  /** One message per rejected file */
  rejected: string[];
}

const KINDS: MediaKind[] = ["photo", "video", "audio"];

let nextId = 0;

function extensionOf(name: string): string {
  const dot = name.lastIndexOf(".");
  return dot === -1 ? "" : name.slice(dot).toLowerCase();
}

/**
 * Which kind of media a file is, or null if it isn't a type we accept. The
 * browser's MIME type decides when it gives a known one; HEIC, MOV and some
 * others arrive with no type or a generic one, so the extension decides then.
 * This is a first guess for files that could be either (.webm, .ogg): the
 * video list is tried first, and checkFiles moves a file with no picture to audio.
 */
export function kindOfFile(file: File): MediaKind | null {
  const byMime = KINDS.find((kind) => ACCEPTED_TYPES[kind].mimeTypes.includes(file.type));
  if (byMime) return byMime;

  const extension = extensionOf(file.name);
  return KINDS.find((kind) => ACCEPTED_TYPES[kind].extensions.includes(extension)) ?? null;
}

/**
 * Checks each file and splits them into accepted and rejected, keeping the
 * order picked. Files are read one at a time: many <video> elements loading
 * at once can make slow ones time out, which would skip their length check.
 */
export async function checkFiles(files: File[]): Promise<CheckedFiles> {
  const accepted: PendingFile[] = [];
  const rejected: string[] = [];
  for (const file of files) {
    let result: PendingFile | string;
    try {
      result = await checkFile(file);
    } catch (err) {
      // One unreadable file must not sink the others
      console.error(`Could not check "${file.name}"`, err);
      result = `"${file.name}" couldn't be checked. Try picking it again.`;
    }
    if (typeof result === "string") rejected.push(result);
    else accepted.push(result);
  }
  return { accepted, rejected };
}

async function checkFile(file: File): Promise<PendingFile | string> {
  let kind = kindOfFile(file);
  if (!kind) return `"${file.name}" isn't a supported file type.`;
  if (file.size > MAX_FILE_BYTES) return `"${file.name}" is too large. Try a shorter clip.`;

  let seconds: number | null = null;
  if (kind !== "photo") {
    const probe = await probeMedia(file);
    seconds = probe.seconds;
    // .webm, .ogg and .mp4 hold either; a "video" with no picture is audio
    if (kind === "video" && probe.audioOnly) kind = "audio";
    // The server rounds before comparing, so a file a fraction over still passes there
    if (seconds !== null && Math.round(seconds) > MAX_DURATION_SECONDS) {
      return `"${file.name}" is too long. Videos and audio can be at most 10 minutes long.`;
    }
  }

  nextId += 1;
  return { id: nextId, file, kind, seconds };
}

/**
 * A photo taken or a recording made on this page, ready to be named and
 * saved, or a message saying why it can't be. Its type needs no check (the
 * page chose it), but its size and length do: `seconds` is how long it was
 * recorded for (null for a photo), since a browser recording may not store
 * its length. Under a second is refused too, as it is almost always Start
 * and Stop pressed together.
 */
export function pendingFromRecording(
  file: File,
  kind: MediaKind,
  seconds: number | null,
  recordedWith: RecordingSource
): PendingFile | string {
  if (file.size > MAX_FILE_BYTES) return "This recording is too large. Try a shorter one.";
  if (seconds !== null && seconds < MIN_RECORDING_SECONDS) return RECORDING_TOO_SHORT;
  // Rounded like the server does
  if (seconds !== null && Math.round(seconds) > MAX_DURATION_SECONDS) {
    return "This recording is too long. Videos and audio can be at most 10 minutes long.";
  }
  nextId += 1;
  return { id: nextId, file, kind, seconds: kind === "photo" ? null : seconds, recordedWith };
}
