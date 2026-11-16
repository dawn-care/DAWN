/**
 * CameraPanel.tsx
 *
 * User: Caregiver
 *
 * The "Use your camera" panel: shows the computer's camera live and takes a
 * photo with it. The camera is opened as soon as the panel opens, so the
 * browser asks for it straight away. Take photo hands the picture to the
 * save step (see SavePanel), where it can be saved or retaken.
 *
 * The photo is exactly what the live picture shows: neither is mirrored. If
 * the camera can't be used (not allowed, none found, in use), the reason is
 * shown with Try again.
 */

import { useEffect, useRef, useState } from "react";
import { Camera, Video, X } from "lucide-react";
import IconButton from "../components/ui/IconButton";
import { SNAPSHOT_FAILED } from "./constants";
import { pendingFromRecording } from "./files";
import type { PendingFile } from "./files";
import { RecordingError, snapshotFile } from "./recording";
import { PRIMARY_BUTTON } from "./styles";
import { useRecorder } from "./useRecorder";

interface CameraPanelProps {
  onClose: () => void;
  /** A photo that is ready to be named and saved */
  onRecorded: (pending: PendingFile) => void;
}

const ACTION_BUTTON = `${PRIMARY_BUTTON} flex items-center gap-2`;

export default function CameraPanel({ onClose, onRecorded }: CameraPanelProps) {
  // Why the last photo can't be used
  const [refused, setRefused] = useState<string | null>(null);
  const [taking, setTaking] = useState(false);

  // Video recording is added to this panel next; for now nothing is recorded
  const { status, stream, error, open } = useRecorder("camera", () => {});

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

  const live = status === "ready" || status === "recording";

  return (
    <div className="flex flex-col gap-3">
      <div className="flex items-center justify-between">
        <h2 className="m-0 text-xl font-extrabold">Use your camera</h2>
        <IconButton label="Close" icon={<X size={20} strokeWidth={2.4} aria-hidden="true" />} onClick={onClose} />
      </div>

      {live ? (
        <>
          <video
            ref={preview}
            autoPlay
            muted
            playsInline
            aria-label="Your camera"
            className="w-full aspect-video rounded-[14px] bg-border object-contain"
          />
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
        {live ? "Camera ready" : ""}
      </p>

      {refused && (
        <p role="alert" className="m-0 text-base font-bold text-link">
          {refused}
        </p>
      )}
    </div>
  );
}
