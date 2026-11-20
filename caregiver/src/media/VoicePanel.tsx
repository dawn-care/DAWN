/**
 * VoicePanel.tsx
 *
 * User: Caregiver
 *
 * The "Record your voice" panel: records the caregiver's voice (a message, a
 * story, a song) with the computer's microphone. The microphone is opened as
 * soon as the panel opens, so the browser asks for it straight away; Start
 * recording then starts at once, and Stop hands the recording to the save
 * step (see SavePanel), where it can be played back, saved or retaken.
 *
 * If the microphone can't be used (not allowed, none found, in use), the
 * reason is shown with Try again; if it stops working mid-recording, nothing
 * is kept and that is said. The page is told while recording, so it can ask
 * before leaving it. Recording stops by itself at 10 minutes, and one under a
 * second is refused here, so it is never uploaded.
 */

import { useEffect, useRef, useState } from "react";
import { Mic, Square, X } from "lucide-react";
import IconButton from "../components/ui/IconButton";
import { pendingFromRecording } from "./files";
import type { PendingFile } from "./files";
import { formatDuration } from "./labels";
import { PRIMARY_BUTTON } from "./styles";
import { useRecorder } from "./useRecorder";
import type { Recording } from "./useRecorder";

interface VoicePanelProps {
  onClose: () => void;
  /** A recording that is ready to be named and saved */
  onRecorded: (pending: PendingFile) => void;
  /** Told when a recording starts and ends, so the page can ask before leaving it */
  onRecordingChange: (recording: boolean) => void;
}

const ACTION_BUTTON = `${PRIMARY_BUTTON} flex items-center gap-2`;

export default function VoicePanel({ onClose, onRecorded, onRecordingChange }: VoicePanelProps) {
  // Why the last recording can't be used (too long or too large)
  const [refused, setRefused] = useState<string | null>(null);

  const handleRecorded = ({ file, seconds }: Recording) => {
    const result = pendingFromRecording(file, "audio", seconds, "voice");
    if (typeof result === "string") setRefused(result);
    else onRecorded(result);
  };

  const { status, seconds, error, open, start, stop } = useRecorder("voice", handleRecorded);

  // Let the page know while a recording would be lost by leaving, and that it no longer is once this panel goes
  useEffect(() => {
    const isRecording = status === "recording";
    onRecordingChange(isRecording);
    return () => {
      if (isRecording) onRecordingChange(false);
    };
  }, [status, onRecordingChange]);

  // The main button for the current step, which gets the keyboard each time the step changes
  const action = useRef<HTMLButtonElement>(null);

  // The caregiver asked to record, so ask for the microphone now
  useEffect(() => {
    void open();
  }, [open]);

  useEffect(() => {
    action.current?.focus();
  }, [status]);

  const begin = () => {
    setRefused(null);
    start();
  };

  return (
    <div className="flex flex-col gap-3">
      <div className="flex items-center justify-between">
        <h2 className="m-0 text-xl font-extrabold">Record your voice</h2>
        <IconButton label="Close" icon={<X size={20} strokeWidth={2.4} aria-hidden="true" />} onClick={onClose} />
      </div>

      <div className="h-[300px] flex flex-col items-center justify-center gap-3 p-4 text-center rounded-[14px] border-[3px] border-ink-muted">
        {status === "error" ? (
          <>
            <Mic size={40} strokeWidth={2} className="text-ink-muted" aria-hidden="true" />
            <p role="alert" className="m-0 text-base font-bold text-link">
              {error}
            </p>
            <button ref={action} type="button" onClick={() => void open()} className={PRIMARY_BUTTON}>
              Try again
            </button>
          </>
        ) : status === "recording" ? (
          <>
            <span className="flex items-center gap-2 text-lg font-extrabold">
              {/* A dot like a recorder's, so it reads as "on" at a glance */}
              <span className="w-3.5 h-3.5 rounded-full bg-accent motion-safe:animate-pulse" aria-hidden="true" />
              Recording
            </span>
            {/* Whole seconds, so the timer never shows a second that hasn't passed */}
            <span className="text-4xl font-extrabold tabular-nums">{formatDuration(Math.floor(seconds))}</span>
            <button ref={action} type="button" onClick={stop} className={ACTION_BUTTON}>
              <Square size={18} fill="currentColor" aria-hidden="true" />
              Stop
            </button>
          </>
        ) : (
          <>
            <Mic size={40} strokeWidth={2} className="text-link" aria-hidden="true" />
            <div className="text-lg font-extrabold">
              {status === "ready" ? "Ready when you are" : "Getting your microphone ready…"}
            </div>
            {status === "ready" ? (
              <button ref={action} type="button" onClick={begin} className={ACTION_BUTTON}>
                <Mic size={20} aria-hidden="true" />
                Start recording
              </button>
            ) : (
              <div className="text-[15px] text-ink-muted">Your browser may ask to use your microphone.</div>
            )}
            <div className="text-[15px] text-ink-muted">Record any audio, up to 10 minutes. Recording automatically stops after 10 minutes.</div>
          </>
        )}
      </div>

      {/* The view changes without moving focus to text, so say what happened */}
      <p role="status" className="sr-only">
        {status === "recording" ? "Recording" : status === "ready" ? "Microphone ready" : ""}
      </p>

      {refused && (
        <p role="alert" className="m-0 text-base font-bold text-link">
          {refused}
        </p>
      )}
    </div>
  );
}
