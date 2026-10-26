/**
 * UploadPanel.tsx
 *
 * User: Caregiver
 *
 * The "Upload a file" panel: a drop zone and a "Choose files" button that
 * take several photos, videos and audio files at once. The picked files are
 * checked first (see files.ts). Files that pass are handed to `onChosen`;
 * files that fail are listed here with the reason, or passed along with the
 * good ones so the next step can show them.
 */

import { useEffect, useRef, useState } from "react";
import type { ChangeEvent, DragEvent } from "react";
import { Upload, X } from "lucide-react";
import IconButton from "../components/ui/IconButton";
import { FILE_INPUT_ACCEPT } from "./constants";
import { checkFiles } from "./files";
import type { CheckedFiles, PendingFile } from "./files";
import { PRIMARY_BUTTON } from "./styles";

interface UploadPanelProps {
  onClose: () => void;
  /** At least one file passed; `rejected` holds a message for each that didn't */
  onChosen: (accepted: PendingFile[], rejected: string[]) => void;
}

export default function UploadPanel({ onClose, onChosen }: UploadPanelProps) {
  const input = useRef<HTMLInputElement>(null);
  const chooseButton = useRef<HTMLButtonElement>(null);
  const [checking, setChecking] = useState(false);
  const [dragging, setDragging] = useState(false);
  // Messages for files that failed when none passed, so there is nothing else to show
  const [rejected, setRejected] = useState<string[]>([]);

  // False once the panel is closed, so a check still running can't queue files the caregiver walked away from
  const mounted = useRef(true);

  // The panel opens because the caregiver asked for it, so put the keyboard in it
  useEffect(() => {
    chooseButton.current?.focus();
    mounted.current = true;
    return () => {
      mounted.current = false;
    };
  }, []);

  const handleFiles = async (picked: File[]) => {
    if (picked.length === 0) return;

    setChecking(true);
    setRejected([]);
    let result: CheckedFiles;
    try {
      result = await checkFiles(picked);
    } catch (err) {
      console.error("Checking the picked files failed", err);
      if (mounted.current) {
        setChecking(false);
        setRejected(["Those files couldn't be checked. Please try again."]);
      }
      return;
    }
    if (!mounted.current) return;
    setChecking(false);

    if (result.accepted.length > 0) onChosen(result.accepted, result.rejected);
    else setRejected(result.rejected);
  };

  const handleChange = (event: ChangeEvent<HTMLInputElement>) => {
    const picked = Array.from(event.target.files ?? []);
    // Clear it so choosing the same file again still counts as a change
    event.target.value = "";
    void handleFiles(picked);
  };

  const handleDrop = (event: DragEvent) => {
    event.preventDefault();
    setDragging(false);
    void handleFiles(Array.from(event.dataTransfer.files));
  };

  return (
    <div className="flex flex-col gap-3">
      <div className="flex items-center justify-between">
        <h2 className="m-0 text-xl font-extrabold">Upload a file</h2>
        <IconButton label="Close" icon={<X size={20} strokeWidth={2.4} aria-hidden="true" />} onClick={onClose} />
      </div>

      <div
        onDragOver={(event) => {
          event.preventDefault();
          setDragging(true);
        }}
        onDragLeave={() => setDragging(false)}
        onDrop={handleDrop}
        className={`h-[300px] flex flex-col items-center justify-center gap-3 p-4 text-center rounded-[14px] border-[3px] border-dashed ${
          dragging ? "border-accent bg-accent/10" : "border-ink-muted"
        }`}
      >
        <Upload size={40} strokeWidth={2} className="text-link" aria-hidden="true" />
        <div className="text-lg font-extrabold">Drag photos, videos or audio here</div>
        <input
          ref={input}
          type="file"
          multiple
          accept={FILE_INPUT_ACCEPT}
          onChange={handleChange}
          className="sr-only"
          tabIndex={-1}
          aria-hidden="true"
        />
        <button
          ref={chooseButton}
          type="button"
          onClick={() => input.current?.click()}
          disabled={checking}
          className={PRIMARY_BUTTON}
        >
          {checking ? "Checking files…" : "Choose files"}
        </button>
        <div className="text-[15px] text-ink-muted">
          Photos, videos up to 10 minutes, or audio and songs. You can pick several.
        </div>
      </div>

      {/* The button label changes while checking, which is not announced by itself */}
      <p role="status" className="sr-only">
        {checking ? "Checking files" : ""}
      </p>

      {rejected.length > 0 && (
        <ul role="alert" className="m-0 pl-5 flex flex-col gap-1 text-base font-bold text-link">
          {rejected.map((message) => (
            <li key={message}>{message}</li>
          ))}
        </ul>
      )}
    </div>
  );
}
