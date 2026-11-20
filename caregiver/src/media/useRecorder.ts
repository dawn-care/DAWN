/**
 * useRecorder.ts
 *
 * User: Caregiver
 *
 * Runs the camera or microphone for one recording panel on the media page:
 * open it (`open`), record (`start`, `stop`) and turn it off (`close`). The
 * open stream is returned so the camera panel can show it live and take a
 * photo from it. When a recording stops, `onRecorded` gets it as a File with
 * its length, measured here because a browser recording may not store one.
 * A recording stops by itself at the 10 minute limit (MAX_DURATION_SECONDS),
 * as if Stop was pressed, so one left running by mistake can still be saved.
 *
 * Status: idle → opening → ready ⇄ recording, or error (with a message for
 * the caregiver; `open` again to retry). The camera and microphone are always
 * turned off when the panel goes away, and a recording still running then is
 * thrown away. If the camera or microphone stops by itself (unplugged,
 * permission taken back), the status becomes error and nothing is kept.
 */

import { useCallback, useEffect, useRef, useState } from "react";
import { MAX_DURATION_SECONDS, RECORDING_MESSAGES } from "./constants";
import type { RecordingSource } from "./constants";
import { RecordingError, openStream, pickRecordingType, recordingFile, stopStream } from "./recording";

export type RecorderStatus = "idle" | "opening" | "ready" | "recording" | "error";

export interface Recording {
  file: File;
  /** How long the recording ran, in seconds */
  seconds: number;
}

export interface Recorder {
  status: RecorderStatus;
  /** The open camera or microphone, while ready or recording */
  stream: MediaStream | null;
  /** Seconds recorded so far; updated a few times a second while recording */
  seconds: number;
  /** Why it isn't working, in the caregiver's words, while the status is error */
  error: string | null;
  /** Opens the camera or microphone (the browser may ask first) */
  open: () => Promise<void>;
  /** Starts recording from the open stream */
  start: () => void;
  /** Stops recording; `onRecorded` gets the result */
  stop: () => void;
  /** Turns the camera or microphone off, throwing away a recording still running */
  close: () => void;
}

// How often the seconds shown while recording are updated
const TICK_MS = 250;
// How often the recorder hands over what it has, so a long recording isn't one buffer at the end
const TIMESLICE_MS = 1000;

export function useRecorder(source: RecordingSource, onRecorded: (recording: Recording) => void): Recorder {
  const [status, setStatus] = useState<RecorderStatus>("idle");
  const [stream, setStream] = useState<MediaStream | null>(null);
  const [seconds, setSeconds] = useState(0);
  const [error, setError] = useState<string | null>(null);

  const streamRef = useRef<MediaStream | null>(null);
  const recorderRef = useRef<MediaRecorder | null>(null);
  const tickRef = useRef<number | null>(null);
  // Stops the recording when it reaches the length limit
  const limitRef = useRef<number | null>(null);
  // Bumped whenever the stream is let go, so an open or a recording that finishes afterwards is ignored
  const generation = useRef(0);
  // The latest callback, so a recording that ends after a re-render reports to the current one
  const onRecordedRef = useRef(onRecorded);
  useEffect(() => {
    onRecordedRef.current = onRecorded;
  });

  const stopTicking = () => {
    if (tickRef.current !== null) window.clearInterval(tickRef.current);
    tickRef.current = null;
    if (limitRef.current !== null) window.clearTimeout(limitRef.current);
    limitRef.current = null;
  };

  // Stops any recording without keeping it and turns the camera or microphone off
  const release = useCallback(() => {
    generation.current += 1;
    stopTicking();
    const recorder = recorderRef.current;
    recorderRef.current = null;
    if (recorder && recorder.state !== "inactive") recorder.stop();
    stopStream(streamRef.current);
    streamRef.current = null;
  }, []);

  // Leaving the panel turns everything off
  useEffect(() => release, [release]);

  const fail = useCallback(
    (message: string) => {
      release();
      setStream(null);
      setError(message);
      setStatus("error");
    },
    [release]
  );

  const open = useCallback(async () => {
    release();
    const mine = generation.current;
    setStream(null);
    setError(null);
    setSeconds(0);
    setStatus("opening");
    try {
      const opened = await openStream(source);
      // Closed, reopened or left while the browser was asking
      if (generation.current !== mine) {
        stopStream(opened);
        return;
      }
      // A device that stops by itself (unplugged, permission taken back, another app took it) ends its
      // track; a recording cut short that way is thrown away with a message, never saved as if Stop
      // was pressed. Turning it off here (stopStream) doesn't fire this, and a later open or close
      // changes the generation, so only this stream's own loss counts.
      opened.getTracks().forEach((track) =>
        track.addEventListener("ended", () => {
          if (generation.current !== mine) return;
          console.warn(`The ${track.kind} track ended by itself`);
          fail(RECORDING_MESSAGES[source].disconnected);
        })
      );
      streamRef.current = opened;
      setStream(opened);
      setStatus("ready");
    } catch (err) {
      if (generation.current !== mine) return;
      setStatus("error");
      setError(err instanceof RecordingError ? err.message : RECORDING_MESSAGES[source].failed);
    }
  }, [source, release, fail]);

  const start = useCallback(() => {
    const current = streamRef.current;
    if (!current || recorderRef.current) return;

    const kind = source === "camera" ? "video" : "audio";
    const type = pickRecordingType(kind);
    let recorder: MediaRecorder;
    try {
      recorder = new MediaRecorder(current, type ? { mimeType: type } : undefined);
    } catch (err) {
      console.error("Could not start recording", err);
      fail(RECORDING_MESSAGES[source].failed);
      return;
    }

    const mine = generation.current;
    const chunks: Blob[] = [];
    let failed = false;
    let startedAt = 0;

    recorder.ondataavailable = (event) => {
      if (event.data.size > 0) chunks.push(event.data);
    };
    recorder.onerror = (event) => {
      console.error("Recording failed", event);
      failed = true;
    };
    recorder.onstop = () => {
      const took = (performance.now() - startedAt) / 1000;
      // Thrown away: the panel was closed or the camera reopened
      if (generation.current !== mine) return;
      stopTicking();
      recorderRef.current = null;
      if (failed || chunks.length === 0) {
        fail(RECORDING_MESSAGES[source].failed);
        return;
      }
      // The browser may add codecs to the type it was given, or choose one itself
      const mimeType = recorder.mimeType || type || chunks[0].type;
      setSeconds(took);
      setStatus("ready");
      onRecordedRef.current({ file: recordingFile(new Blob(chunks, { type: mimeType }), kind, mimeType), seconds: took });
    };

    try {
      recorder.start(TIMESLICE_MS);
    } catch (err) {
      // e.g. the stream's tracks have ended (the device was unplugged)
      console.error("Could not start recording", err);
      fail(RECORDING_MESSAGES[source].failed);
      return;
    }
    startedAt = performance.now();
    recorderRef.current = recorder;
    setSeconds(0);
    setStatus("recording");
    tickRef.current = window.setInterval(() => setSeconds((performance.now() - startedAt) / 1000), TICK_MS);
    // Stopped like Stop at the limit, so the recording is kept and can be saved
    limitRef.current = window.setTimeout(() => {
      if (recorder.state !== "inactive") recorder.stop();
    }, MAX_DURATION_SECONDS * 1000);
  }, [source, fail]);

  const stop = useCallback(() => {
    const recorder = recorderRef.current;
    if (recorder && recorder.state !== "inactive") recorder.stop();
  }, []);

  const close = useCallback(() => {
    release();
    setStream(null);
    setError(null);
    setSeconds(0);
    setStatus("idle");
  }, [release]);

  return { status, stream, seconds, error, open, start, stop, close };
}
