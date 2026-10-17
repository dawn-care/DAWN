/**
 * DetailPanel.tsx
 *
 * User: Caregiver
 *
 * The right-hand panel for one selected photo, video or audio file:
 *
 *  - A preview (the photo itself, or an icon) and, for video and audio, a
 *    Play button that opens an inline player
 *  - Its name, which can be renamed in place
 *  - When it plays: a video or audio file shows its schedule ("Every day at
 *    8:00 am", linking to the Calendar); a photo shows whether it is on
 *    the screen, plus its schedule if it has timed showings
 *  - Remove, which asks first (see RemoveMediaDialog); the page does the
 *    removing and offers Undo
 *
 * Mount it with `key={media.id}` so the player and any rename in progress
 * reset when another item is selected.
 */

import { useRef, useState } from "react";
import type { FormEvent, KeyboardEvent, ReactNode } from "react";
import { Link } from "react-router-dom";
import { Image, Music, Play, Video } from "lucide-react";
import TextField from "../components/ui/form/TextField";
import { fileUrl, friendlyError, renameMedia } from "./api";
import type { Media, ScheduledItem } from "./api";
import { scheduleLine } from "./labels";
import RemoveMediaDialog from "./RemoveMediaDialog";
import { ACTION_LINK, PRIMARY_BUTTON, SECONDARY_BUTTON } from "./styles";
import { FOCUS_RING } from "../components/ui/styles";

interface DetailPanelProps {
  media: Media;
  /** All scheduled items, or null if they couldn't be loaded */
  scheduled: ScheduledItem[] | null;
  /** The server accepted a new name; `updated` is the item as it is now */
  onRenamed: (updated: Media) => void;
  /** The caregiver confirmed removing the item */
  onRemoved: () => void;
}

/* ============================
   Pieces
============================ */

// A small caps heading with a row of text and an optional action link beside it
function InfoBlock({ heading, children }: { heading: string; children: ReactNode }) {
  return (
    <div className="flex flex-col gap-1">
      <div className="text-[13px] font-bold uppercase tracking-[0.06em] text-dawn-muted">{heading}</div>
      <div className="flex items-center justify-between gap-2 text-base leading-snug">{children}</div>
    </div>
  );
}

// The schedule line is itself a link to the Calendar, with "See on calendar" beside it
function ScheduleBlock({ line }: { line: string | null }) {
  return (
    <InfoBlock heading="Schedule">
      {line ? (
        <>
          <Link to="/calendar" className={`text-dawn-ink underline underline-offset-2 hover:text-dawn-accent ${FOCUS_RING}`}>
            {line}
          </Link>
          <Link to="/calendar" className={ACTION_LINK}>
            See on calendar
          </Link>
        </>
      ) : (
        <>
          <span>Not scheduled</span>
          <Link to="/calendar" className={ACTION_LINK}>
            Schedule it
          </Link>
        </>
      )}
    </InfoBlock>
  );
}

const PREVIEW_BOX = "h-[196px] flex items-center justify-center overflow-hidden rounded-xl bg-dawn-avatar text-dawn-avatar-ink";

/* ============================
   Component
============================ */

export default function DetailPanel({ media, scheduled, onRenamed, onRemoved }: DetailPanelProps) {
  const [playing, setPlaying] = useState(false);
  const [playFailed, setPlayFailed] = useState(false);
  // Browsers can't draw every photo type (HEIC outside Safari); show the icon instead
  const [photoBroken, setPhotoBroken] = useState(false);

  const [renaming, setRenaming] = useState(false);
  const [draft, setDraft] = useState(media.display_name);
  const [saving, setSaving] = useState(false);
  const [renameError, setRenameError] = useState<string | null>(null);
  const renameButton = useRef<HTMLButtonElement>(null);
  const nameField = useRef<HTMLInputElement>(null);

  const [confirmingRemove, setConfirmingRemove] = useState(false);

  const isPhoto = media.kind === "photo";
  const line = scheduleLine(media, scheduled);
  const onScreen = media.usage.includes("on_screen");
  const PlaceholderIcon = { photo: Image, video: Video, audio: Music }[media.kind];

  /* ============================
     Rename
  ============================ */

  const startRename = () => {
    setDraft(media.display_name);
    setRenameError(null);
    setRenaming(true);
    // The field mounts on the next render
    requestAnimationFrame(() => {
      nameField.current?.focus();
      nameField.current?.select();
    });
  };

  const stopRename = () => {
    setRenaming(false);
    setRenameError(null);
    renameButton.current?.focus();
  };

  const submitRename = async (event: FormEvent) => {
    event.preventDefault();
    if (draft.trim() === media.display_name) {
      stopRename();
      return;
    }

    setSaving(true);
    setRenameError(null);
    try {
      onRenamed(await renameMedia(media.id, draft));
    } catch (err) {
      setRenameError(friendlyError(err));
      setSaving(false);
      return;
    }
    setSaving(false);
    stopRename();
  };

  const handleRenameKeyDown = (event: KeyboardEvent) => {
    if (event.key === "Escape" && !saving) stopRename();
  };

  /* ============================
     Render
  ============================ */

  return (
    <div className="flex flex-col gap-3">
      {/* Preview, or the player once Play is pressed */}
      {playing && media.kind === "video" ? (
        <video
          src={fileUrl(media)}
          controls
          autoPlay
          onError={() => setPlayFailed(true)}
          className="w-full h-[196px] rounded-xl bg-black"
        />
      ) : (
        <div className={PREVIEW_BOX}>
          {isPhoto && !photoBroken ? (
            <img
              src={fileUrl(media)}
              alt={media.display_name}
              onError={() => setPhotoBroken(true)}
              className="w-full h-full object-contain"
            />
          ) : (
            <PlaceholderIcon size={48} aria-hidden="true" />
          )}
        </div>
      )}

      {playing && media.kind === "audio" && (
        <audio src={fileUrl(media)} controls autoPlay onError={() => setPlayFailed(true)} className="w-full" />
      )}
      {playFailed && (
        <p role="alert" className="m-0 text-base font-bold text-dawn-accent">
          This file couldn&apos;t be played in your browser.
        </p>
      )}

      {!isPhoto && !playing && (
        <button
          type="button"
          onClick={() => setPlaying(true)}
          className={`self-start h-11 px-4 flex items-center gap-2 rounded-[10px] bg-dawn-ink text-base font-bold text-white cursor-pointer ${FOCUS_RING}`}
        >
          <Play size={16} fill="currentColor" aria-hidden="true" />
          Play
        </button>
      )}

      {/* Name, or the rename form */}
      {renaming ? (
        <form onSubmit={submitRename} onKeyDown={handleRenameKeyDown} className="flex flex-col gap-2">
          <TextField
            ref={nameField}
            label="Name"
            value={draft}
            onChange={(event) => setDraft(event.target.value)}
            aria-invalid={renameError ? true : undefined}
            aria-describedby={renameError ? "rename-error" : undefined}
          />
          {renameError && (
            <p id="rename-error" role="alert" className="m-0 text-base font-bold text-dawn-accent">
              {renameError}
            </p>
          )}
          <div className="flex gap-2">
            <button type="submit" disabled={saving} className={`flex-1 ${PRIMARY_BUTTON}`}>
              {saving ? "Saving…" : "Save"}
            </button>
            <button type="button" onClick={stopRename} disabled={saving} className={`flex-1 ${SECONDARY_BUTTON}`}>
              Cancel
            </button>
          </div>
        </form>
      ) : (
        <h2 className="m-0 text-xl font-extrabold break-words">{media.display_name}</h2>
      )}

      {/* When it plays */}
      {isPhoto ? (
        <>
          <InfoBlock heading="Screen">
            <span>{onScreen ? "On screen" : "Not on screen"}</span>
            <Link to="/" className={ACTION_LINK}>
              {onScreen ? "Remove from screen" : "Add to screen"}
            </Link>
          </InfoBlock>
          {line && <ScheduleBlock line={line} />}
        </>
      ) : (
        <ScheduleBlock line={line} />
      )}

      <div className="flex gap-2">
        <button ref={renameButton} type="button" onClick={startRename} className={`flex-1 ${SECONDARY_BUTTON}`}>
          Rename
        </button>
        <button type="button" onClick={() => setConfirmingRemove(true)} className={`flex-1 ${SECONDARY_BUTTON}`}>
          Remove
        </button>
      </div>

      <RemoveMediaDialog
        media={media}
        open={confirmingRemove}
        onOpenChange={setConfirmingRemove}
        onConfirm={onRemoved}
      />
    </div>
  );
}
