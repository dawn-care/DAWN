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
 *    8 AM", linking to the Calendar); a photo shows whether it is on
 *    the screen, plus its schedule if it has timed showings
 *  - How staff find it: its situations and family members, with Edit, which
 *    opens the same picker as the Save panel (see TagPicker)
 *  - Remove, which asks first (see RemoveMediaDialog); the page does the
 *    removing and offers Undo
 *
 * Mount it with `key={media.id}` so the player and any rename in progress
 * reset when another item is selected.
 */

import { useRef, useState } from "react";
import type { FormEvent, KeyboardEvent, ReactNode, SyntheticEvent } from "react";
import { Link } from "react-router-dom";
import { Check, Image, Music, Play, Video } from "lucide-react";
import TextField from "../components/ui/form/TextField";
import { fileUrl, friendlyError, renameMedia, setMediaTags } from "./api";
import type { FamilyMember, Media, ScheduledItem } from "./api";
import { familyLine, scheduleLine, situationsLine } from "./labels";
import RemoveMediaDialog from "./RemoveMediaDialog";
import { ACTION_LINK, PRIMARY_BUTTON, SECONDARY_BUTTON } from "./styles";
import TagPicker from "./TagPicker";
import { FOCUS_RING } from "../components/ui/styles";

interface DetailPanelProps {
  media: Media;
  /** All scheduled items, or null if they couldn't be loaded */
  scheduled: ScheduledItem[] | null;
  /** The carepartner's family, or null if it couldn't be loaded */
  family: FamilyMember[] | null;
  onFamilyAdded: (member: FamilyMember) => void;
  /** The caregiver just saved this item; shows "Saved. You can use it any time." */
  justSaved?: boolean;
  /** The server accepted a new name; `updated` is the item as it is now */
  onRenamed: (updated: Media) => void;
  /** The server accepted new situations or family members */
  onTagsChanged: (updated: Media) => void;
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
          <Link to="/calendar" className={`inline-flex items-center min-h-11 text-dawn-ink underline underline-offset-2 hover:text-dawn-accent ${FOCUS_RING}`}>
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

export default function DetailPanel({
  media,
  scheduled,
  family,
  onFamilyAdded,
  justSaved = false,
  onRenamed,
  onTagsChanged,
  onRemoved,
}: DetailPanelProps) {
  const [playing, setPlaying] = useState(false);
  // Why playing failed, worded for the caregiver; null while it hasn't
  const [playError, setPlayError] = useState<string | null>(null);
  // The photo didn't load (a type the browser can't draw, such as HEIC outside Safari, or the file is unreachable); show the icon instead
  const [photoBroken, setPhotoBroken] = useState(false);

  const [renaming, setRenaming] = useState(false);
  const [draft, setDraft] = useState(media.display_name);
  const [saving, setSaving] = useState(false);
  const [renameError, setRenameError] = useState<string | null>(null);
  const renameButton = useRef<HTMLButtonElement>(null);
  const nameField = useRef<HTMLInputElement>(null);

  const [confirmingRemove, setConfirmingRemove] = useState(false);

  const [editingTags, setEditingTags] = useState(false);
  const [draftSituations, setDraftSituations] = useState(media.situations);
  const [draftFamily, setDraftFamily] = useState(media.family_member_ids);
  const [savingTags, setSavingTags] = useState(false);
  const [tagsError, setTagsError] = useState<string | null>(null);
  const editTagsButton = useRef<HTMLButtonElement>(null);

  const isPhoto = media.kind === "photo";
  const line = scheduleLine(media, scheduled);
  const onScreen = media.usage.includes("on_screen");
  const PlaceholderIcon = { photo: Image, video: Video, audio: Music }[media.kind];

  const handlePlayError = (event: SyntheticEvent<HTMLMediaElement>) => {
    const error = event.currentTarget.error;
    console.error(`Could not play media ${media.id}`, error);
    setPlayError(
      error?.code === MediaError.MEDIA_ERR_SRC_NOT_SUPPORTED
        ? "This file couldn't be played in your browser."
        : "This file couldn't be loaded. Check your connection and try again."
    );
  };

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
    let updated: Media;
    try {
      updated = await renameMedia(media.id, draft);
    } catch (err) {
      setRenameError(friendlyError(err));
      setSaving(false);
      return;
    }
    // Outside the try: a problem handling the new name must not read as a failed rename
    onRenamed(updated);
    setSaving(false);
    stopRename();
  };

  const handleRenameKeyDown = (event: KeyboardEvent) => {
    if (event.key === "Escape" && !saving) stopRename();
  };

  /* ============================
     How staff find it
  ============================ */

  const startEditingTags = () => {
    setDraftSituations(media.situations);
    setDraftFamily(media.family_member_ids);
    setTagsError(null);
    setEditingTags(true);
  };

  const stopEditingTags = () => {
    setEditingTags(false);
    setTagsError(null);
    requestAnimationFrame(() => editTagsButton.current?.focus());
  };

  const saveTags = async () => {
    setSavingTags(true);
    setTagsError(null);
    let updated: Media;
    try {
      updated = await setMediaTags(media.id, { situations: draftSituations, familyMemberIds: draftFamily });
    } catch (err) {
      setTagsError(friendlyError(err));
      setSavingTags(false);
      return;
    }
    // Outside the try: a problem showing the change must not read as a failed save
    onTagsChanged(updated);
    setSavingTags(false);
    stopEditingTags();
  };

  /* ============================
     Render
  ============================ */

  return (
    <div className="flex flex-col gap-3">
      {justSaved && (
        <p
          role="status"
          className="m-0 flex items-center gap-2 px-2.5 py-2 rounded-lg bg-dawn-success/10 text-[15px] font-bold text-dawn-success"
        >
          <Check size={18} strokeWidth={2.6} aria-hidden="true" />
          Saved. You can use it any time.
        </p>
      )}

      {/* Preview, or the player once Play is pressed */}
      {playing && media.kind === "video" ? (
        <video
          src={fileUrl(media)}
          controls
          autoPlay
          onError={handlePlayError}
          className="w-full h-[196px] rounded-xl bg-black"
        />
      ) : (
        <div className={PREVIEW_BOX}>
          {isPhoto && !photoBroken ? (
            <img
              src={fileUrl(media)}
              alt={media.display_name}
              onError={() => {
                console.warn(`Photo ${media.id} did not load`);
                setPhotoBroken(true);
              }}
              className="w-full h-full object-contain"
            />
          ) : (
            <PlaceholderIcon size={48} aria-hidden="true" />
          )}
        </div>
      )}

      {playing && media.kind === "audio" && (
        <audio src={fileUrl(media)} controls autoPlay onError={handlePlayError} className="w-full" />
      )}
      {playError && (
        <p role="alert" className="m-0 text-base font-bold text-dawn-accent">
          {playError}
        </p>
      )}

      {!isPhoto && !playing && (
        <button
          type="button"
          onClick={() => {
            setPlayError(null);
            setPlaying(true);
          }}
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

      {/* How staff find it on their phones */}
      {editingTags ? (
        <div className="flex flex-col gap-3 p-3 rounded-xl border-2 border-dawn-line">
          <TagPicker
            carepartnerId={media.carepartner_id}
            situations={draftSituations}
            familyMemberIds={draftFamily}
            family={family}
            onSituationsChange={setDraftSituations}
            onFamilyMemberIdsChange={setDraftFamily}
            onFamilyAdded={onFamilyAdded}
            disabled={savingTags}
          />
          {tagsError && (
            <p role="alert" className="m-0 text-base font-bold text-dawn-accent">
              {tagsError}
            </p>
          )}
          <div className="flex gap-2">
            <button type="button" onClick={() => void saveTags()} disabled={savingTags} className={`flex-1 ${PRIMARY_BUTTON}`}>
              {savingTags ? "Saving…" : "Save"}
            </button>
            <button type="button" onClick={stopEditingTags} disabled={savingTags} className={`flex-1 ${SECONDARY_BUTTON}`}>
              Cancel
            </button>
          </div>
        </div>
      ) : (
        <div className="flex flex-col gap-1">
          <div className="flex items-center justify-between gap-2">
            <div className="text-[13px] font-bold uppercase tracking-[0.06em] text-dawn-muted">Staff find it</div>
            <button ref={editTagsButton} type="button" onClick={startEditingTags} className={`${ACTION_LINK} cursor-pointer bg-transparent border-0 p-0`}>
              Edit
            </button>
          </div>
          <p className="m-0 text-base leading-snug">
            <span className="font-bold">When: </span>
            {situationsLine(media)}
          </p>
          <p className="m-0 text-base leading-snug">
            <span className="font-bold">About: </span>
            {familyLine(media, family)}
          </p>
        </div>
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
