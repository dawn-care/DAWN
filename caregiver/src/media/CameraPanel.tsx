/**
 * CameraPanel.tsx
 *
 * User: Caregiver
 *
 * The "Use your camera" panel: shows the computer's camera live, and takes a
 * photo or records a video (with sound) with it, switched with Photo / Video.
 * The camera is opened as soon as the panel opens, so the browser asks for it
 * straight away. Take photo, or Stop after recording, hands the result to the
 * save step (see SavePanel), where it can be played back, saved or retaken.
 *
 * The photo or video is exactly what the live picture shows: neither is
 * mirrored. If the camera can't be used (not allowed, none found, in use),
 * the reason is shown with Try again. A video over 10 minutes is refused
 * here, so it is never uploaded.
 */

import { useEffect, useRef, useState } from "react";
import { Camera, Circle, Square, Video, X } from "lucide-react";
import IconButton from "../components/ui/IconButton";
import SegmentedControl from "../components/ui/form/SegmentedControl";
import type { Segment } from "../components/ui/form/SegmentedControl";
import { SNAPSHOT_FAILED } from "./constants";
import { pendingFromRecording } from "./files";
import type { PendingFile } from "./files";
import { formatDuration } from "./labels";
import { RecordingError, snapshotFile } from "./recording";
import { PRIMARY_BUTTON } from "./styles";
import { useRecorder } from "./useRecorder";
import type { Recording } from "./useRecorder";

export type CameraMode = "photo" | "video";

interface CameraPanelProps {
  /** Photo or Video; held by the page, so it is the same after Retake or reopening the panel */
  mode: CameraMode;
  onModeChange: (mode: CameraMode) => void;
  onClose: () => void;
  /** A photo or video that is ready to be named and saved */
  onRecorded: (pending: PendingFile) => void;
}

const MODES: Segment<CameraMode>[] = [
  { value: "photo", label: "Photo" },
  { value: "video", label: "Video" },
];

const ACTION_BUTTON = `${PRIMARY_BUTTON} flex items-center gap-2`;

export default function CameraPanel({ mode, onModeChange, onClose, onRecorded }: CameraPanelProps) {
  // Why the last photo or video can't be used
  const [refused, setRefused] = useState<string | null>(null);
  const [taking, setTaking] = useState(false);

  const handleRecorded = ({ file, seconds }: Recording) => {
    const result = pendingFromRecording(file, "video", seconds, "camera");
    if (typeof result === "string") setRefused(result);
    else onRecorded(result);
  };

  const { status, stream, seconds, error, open, start, stop } = useRecorder("camera", handleRecorded);

  const preview = useRef<HTMLVideoElement>(null);
  // The main button for the current step, which gets the keyboard each time the step changes
  const action = useRef<HTMLButtonElement>(null);
  // False once the panel is closed, so a photo still being made can't be handed on after the caregiver left
  const mounted = useRef(true);

  useEffect(() => {
    mounted.current = true;
    return () => {
      mounted.current = false;
    };
  }, []);

  // The caregiver asked for the camera, so ask for it now
  useEffect(() => {
    void open();
  }, [open]);

  // Not on a mode change: the keyboard stays on the Photo / Video switch while it's used
  useEffect(() => {
    action.current?.focus();
  }, [status]);

  // Show the camera live while it's open
  useEffect(() => {
    if (preview.current) preview.current.srcObject = stream;
  }, [stream]);

  const takePhoto = async () => {
    if (!preview.current) return;
    setRefused(null);
    setTaking(true);
    let file: File;
    try {
      file = await snapshotFile(preview.current);
    } catch (err) {
      console.error("Could not take a photo", err);
      if (mounted.current) {
        setTaking(false);
        setRefused(err instanceof RecordingError ? err.message : SNAPSHOT_FAILED);
      }
      return;
    }
    if (!mounted.current) return;
    setTaking(false);
    const result = pendingFromRecording(file, "photo", null, "camera");
    if (typeof result === "string") setRefused(result);
    else onRecorded(result);
  };

  const begin = () => {
    setRefused(null);
    start();
  };

  const changeMode = (next: CameraMode) => {
    setRefused(null);
    onModeChange(next);
  };

  const live = status === "ready" || status === "recording";
  const recording = status === "recording";

  return (
    <div className="flex flex-col gap-3">
      <div className="flex items-center justify-between">
        <h2 className="m-0 text-xl font-extrabold">Use your camera</h2>
        <IconButton label="Close" icon={<X size={20} strokeWidth={2.4} aria-hidden="true" />} onClick={onClose} />
      </div>

      {live ? (
        <>
          <SegmentedControl
            segments={MODES}
            value={mode}
            onChange={changeMode}
            label="Take a photo or record a video"
            disabled={recording || taking}
          />

          <div className="relative">
            <video
              ref={preview}
              autoPlay
              muted
              playsInline
              aria-label="Your camera"
              className="block w-full aspect-video rounded-[14px] bg-border object-contain"
            />
            {recording && (
              <span className="absolute top-2.5 left-2.5 flex items-center gap-2 px-2.5 py-1 rounded-full bg-surface/90 text-base font-extrabold tabular-nums">
                {/* A dot like a recorder's, so it reads as "on" at a glance */}
                <span className="w-3 h-3 rounded-full bg-accent motion-safe:animate-pulse" aria-hidden="true" />
                {/* Whole seconds, so the timer never shows a second that hasn't passed */}
                {formatDuration(Math.floor(seconds))}
              </span>
            )}
          </div>

          {mode === "photo" ? (
            <button
              ref={action}
              type="button"
              onClick={() => void takePhoto()}
              disabled={taking}
              className={`self-center ${ACTION_BUTTON}`}
            >
              <Camera size={20} aria-hidden="true" />
              {/* Making the picture can take a moment */}
              {taking ? "Taking photo…" : "Take photo"}
            </button>
          ) : recording ? (
            <button ref={action} type="button" onClick={stop} className={`self-center ${ACTION_BUTTON}`}>
              <Square size={18} fill="currentColor" aria-hidden="true" />
              Stop
            </button>
          ) : (
            <button ref={action} type="button" onClick={begin} className={`self-center ${ACTION_BUTTON}`}>
              <Circle size={18} fill="currentColor" aria-hidden="true" />
              Start recording
            </button>
          )}

          {mode === "video" && !recording && (
            <div className="text-center text-[15px] text-ink-muted">Videos can be up to 10 minutes long.</div>
          )}
        </>
      ) : (
        <div className="h-[300px] flex flex-col items-center justify-center gap-3 p-4 text-center rounded-[14px] border-[3px] border-ink-muted">
          {status === "error" ? (
            <>
              <Video size={40} strokeWidth={2} className="text-ink-muted" aria-hidden="true" />
              <p role="alert" className="m-0 text-base font-bold text-link">
                {error}
              </p>
              <button ref={action} type="button" onClick={() => void open()} className={PRIMARY_BUTTON}>
                Try again
              </button>
            </>
          ) : (
            <>
              <Video size={40} strokeWidth={2} className="text-link" aria-hidden="true" />
              <div className="text-lg font-extrabold">Getting your camera ready…</div>
              <div className="text-[15px] text-ink-muted">Your browser may ask to use your camera and microphone.</div>
            </>
          )}
        </div>
      )}

      {/* The view changes without moving focus to text, so say what happened */}
      <p role="status" className="sr-only">
        {recording ? "Recording" : live ? "Camera ready" : ""}
      </p>

      {refused && (
        <p role="alert" className="m-0 text-base font-bold text-link">
          {refused}
        </p>
      )}
    </div>
  );
}
