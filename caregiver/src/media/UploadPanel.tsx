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

import { useRef, useState } from "react";
import type { ChangeEvent, DragEvent } from "react";
import { Upload, X } from "lucide-react";
import IconButton from "../components/ui/IconButton";
import { FILE_INPUT_ACCEPT } from "./constants";
import { checkFiles } from "./files";
import type { PendingFile } from "./files";
import { PRIMARY_BUTTON } from "./styles";

interface UploadPanelProps {
  onClose: () => void;
  /** At least one file passed; `rejected` holds a message for each that didn't */
  onChosen: (accepted: PendingFile[], rejected: string[]) => void;
}

export default function UploadPanel({ onClose, onChosen }: UploadPanelProps) {
  const input = useRef<HTMLInputElement>(null);
  const [checking, setChecking] = useState(false);
  const [dragging, setDragging] = useState(false);
  // Messages for files that failed when none passed, so there is nothing else to show
  const [rejected, setRejected] = useState<string[]>([]);

  const handleFiles = async (picked: File[]) => {
    if (picked.length === 0) return;

    setChecking(true);
    setRejected([]);
    const result = await checkFiles(picked);
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
          dragging ? "border-dawn-accent bg-dawn-accent/10" : "border-dawn-control"
        }`}
      >
        <Upload size={40} strokeWidth={2} className="text-dawn-accent" aria-hidden="true" />
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
          type="button"
          onClick={() => input.current?.click()}
          disabled={checking}
          className={PRIMARY_BUTTON}
        >
          {checking ? "Checking files…" : "Choose files"}
        </button>
        <div className="text-[15px] text-dawn-muted">
          Photos, videos up to 10 minutes, or audio and songs. You can pick several.
        </div>
      </div>

      {rejected.length > 0 && (
        <ul role="alert" className="m-0 pl-5 flex flex-col gap-1 text-base font-bold text-dawn-accent">
          {rejected.map((message) => (
            <li key={message}>{message}</li>
          ))}
        </ul>
      )}
    </div>
  );
}
