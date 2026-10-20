/**
 * SavePanel.tsx
 *
 * User: Caregiver
 *
 * Names and saves one picked file. The caregiver sees a preview (the photo
 * itself, or an icon for video and audio), a name field (starting as the
 * file name), how staff find it (see TagPicker: nothing is chosen to start
 * with, so staff don't see it until the caregiver picks) and Save; there is no scheduling or
 * placement here, that happens from the detail panel afterwards.
 *
 * When several files were picked, the panel is shown for each in turn with
 * "File 2 of 5". Skip is always there. Close stops going through the rest,
 * and cancels the upload if one is running. If saving fails the file is
 * kept, so Save becomes "Try again" and nothing has to be picked again.
 */

import { useEffect, useRef, useState } from "react";
import type { FormEvent } from "react";
import { Image, Music, Video, X } from "lucide-react";
import IconButton from "../components/ui/IconButton";
import TextField from "../components/ui/form/TextField";
import { friendlyError, isAbort, uploadMedia } from "./api";
import type { FamilyMember, Media } from "./api";
import { ERROR_MESSAGES } from "./constants";
import type { Situation } from "./constants";
import type { PendingFile } from "./files";
import { PRIMARY_BUTTON, SECONDARY_BUTTON } from "./styles";
import TagPicker from "./TagPicker";

interface SavePanelProps {
  carepartnerId: number;
  pending: PendingFile;
  /** The carepartner's family, or null if it couldn't be loaded */
  family: FamilyMember[] | null;
  onFamilyAdded: (member: FamilyMember) => void;
  /** 1-based place of this file among the ones picked */
  position: number;
  total: number;
  /** Messages for files that were turned away before this step; shown on the first file only */
  notices: string[];
  onSaved: (media: Media) => void;
  onSkip: () => void;
  /** Stop going through the rest */
  onClose: () => void;
}

const KIND_ICONS = { photo: Image, video: Video, audio: Music };

// "Beach day.mp4" -> "Beach day"
function nameFromFile(fileName: string): string {
  return fileName.replace(/\.[^.]+$/, "");
}

/**
 * Mount with `key={pending.id}` so the name and any error start fresh for
 * each file.
 */
export default function SavePanel({
  carepartnerId,
  pending,
  family,
  onFamilyAdded,
  position,
  total,
  notices,
  onSaved,
  onSkip,
  onClose,
}: SavePanelProps) {
  const { file, kind } = pending;
  const [name, setName] = useState(() => nameFromFile(file.name));
  // Nothing to start with: the caregiver opts each item into the staff app
  const [situations, setSituations] = useState<Situation[]>([]);
  const [familyMemberIds, setFamilyMemberIds] = useState<number[]>([]);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);
  // The last attempt reached the server and failed, as opposed to an empty name caught here
  const [saveFailed, setSaveFailed] = useState(false);
  const [previewUrl, setPreviewUrl] = useState<string | null>(null);
  const [previewBroken, setPreviewBroken] = useState(false);
  const nameField = useRef<HTMLInputElement>(null);
  // The upload in progress, so closing the panel can cancel it
  const upload = useRef<AbortController | null>(null);

  // Each file's panel opens ready to type its name; the suggested name is selected so typing replaces it
  useEffect(() => {
    nameField.current?.focus();
    nameField.current?.select();
  }, []);

  // Leaving the panel any other way (e.g. picking a card) also cancels a running upload
  useEffect(() => () => upload.current?.abort(), []);

  // A photo is shown as itself; the object URL is released when the file changes or the panel goes
  useEffect(() => {
    if (kind !== "photo") return;
    const url = URL.createObjectURL(file);
    setPreviewUrl(url);
    return () => URL.revokeObjectURL(url);
  }, [file, kind]);

  const Icon = KIND_ICONS[kind];

  const save = async (event: FormEvent) => {
    event.preventDefault();
    if (name.trim() === "") {
      setError(ERROR_MESSAGES.INVALID_NAME);
      return;
    }

    setSaving(true);
    setError(null);
    setSaveFailed(false);
    const controller = new AbortController();
    upload.current = controller;
    let saved: Media;
    try {
      saved = await uploadMedia({
        carepartnerId,
        file,
        kind,
        displayName: name,
        tags: { situations, familyMemberIds },
        signal: controller.signal,
      });
    } catch (err) {
      // Closing the panel cancelled it; there is nothing left to update
      if (isAbort(err)) return;
      setError(friendlyError(err));
      setSaveFailed(true);
      setSaving(false);
      return;
    }
    // Outside the try: a problem handling the saved item must not read as a failed upload (retrying would add a duplicate)
    onSaved(saved);
  };

  // Close while an upload is running cancels it
  const close = () => {
    upload.current?.abort();
    onClose();
  };

  return (
    <form onSubmit={save} className="flex flex-col gap-3">
      <div className="flex items-center justify-between">
        <h2 className="m-0 text-xl font-extrabold">Name and Save</h2>
        <IconButton
          label="Close"
          icon={<X size={20} strokeWidth={2.4} aria-hidden="true" />}
          onClick={close}
        />
      </div>

      {total > 1 && (
        <p className="m-0 text-base font-bold text-dawn-muted">
          File {position} of {total}
        </p>
      )}

      {position === 1 && notices.length > 0 && (
        <ul role="alert" className="m-0 pl-5 flex flex-col gap-1 text-base font-bold text-dawn-accent">
          {notices.map((message) => (
            <li key={message}>{message}</li>
          ))}
        </ul>
      )}

      <div className="flex items-center gap-3">
        <div className="w-[140px] h-20 shrink-0 flex items-center justify-center overflow-hidden rounded-[10px] bg-dawn-avatar text-dawn-avatar-ink">
          {previewUrl && !previewBroken ? (
            <img
              src={previewUrl}
              alt=""
              onError={() => setPreviewBroken(true)}
              className="w-full h-full object-cover"
            />
          ) : (
            <Icon size={32} aria-hidden="true" />
          )}
        </div>
        <div className="grow min-w-0">
          <TextField
            ref={nameField}
            label="Name"
            value={name}
            onChange={(event) => setName(event.target.value)}
            disabled={saving}
            aria-invalid={error ? true : undefined}
            aria-describedby={error ? "save-error" : undefined}
          />
        </div>
      </div>

      <p className="m-0 text-[15px] text-dawn-muted break-words">{file.name}</p>

      <TagPicker
        carepartnerId={carepartnerId}
        situations={situations}
        familyMemberIds={familyMemberIds}
        family={family}
        onSituationsChange={setSituations}
        onFamilyMemberIdsChange={setFamilyMemberIds}
        onFamilyAdded={onFamilyAdded}
        disabled={saving}
      />

      {error && (
        <p id="save-error" role="alert" className="m-0 text-base font-bold text-dawn-accent">
          {error}
        </p>
      )}

      <div className="flex justify-between gap-2">
        <button type="button" onClick={onSkip} disabled={saving} className={SECONDARY_BUTTON}>
          Skip
        </button>
        <button type="submit" disabled={saving} className={PRIMARY_BUTTON}>
          {saving ? "Saving…" : saveFailed ? "Try again" : "Save"}
        </button>
      </div>
    </form>
  );
}
