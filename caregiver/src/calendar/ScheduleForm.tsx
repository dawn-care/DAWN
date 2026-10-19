/**
 * ScheduleForm.tsx
 *
 * User: Caregiver
 *
 * The form that replaces the calendar's left column to schedule a video or
 * audio ("Schedule") or change one ("Edit"). It holds the choices (video or
 * audio, date, time, which file, repeats and, for audio, what shows while it
 * plays), tells the page where a new item would land so the calendar can draw
 * its "New" placeholder, and saves.
 *
 * Saving first checks the day for another item that would still be playing
 * (the server leaves overlap to this page), then creates the item or replaces the one
 * being edited. Why a time can't be saved comes back in words, next to the
 * form.
 */

import { formatClock, overlappingItems } from "@shared";
import type { AwakeHours, PlayableItem } from "@shared";
import { Trash2, X } from "lucide-react";
import { useEffect, useRef, useState } from "react";
import { Link } from "react-router-dom";
import IconButton from "../components/ui/IconButton";
import { FOCUS_RING } from "../components/ui/styles";
import PhotoPicker from "../components/ui/PhotoPicker";
import RadioGroup from "../components/ui/form/RadioGroup";
import RadioOption from "../components/ui/form/RadioOption";
import SegmentedControl from "../components/ui/form/SegmentedControl";
import SelectField from "../components/ui/form/SelectField";
import TextField from "../components/ui/form/TextField";
import { fileUrl } from "../media/api";
import type { Media } from "../media/api";
import { formatDuration } from "../media/labels";
import { PRIMARY_BUTTON, SECONDARY_BUTTON } from "../media/styles";
import { createScheduledItem, friendlyError, getCalendar, updateScheduledItem } from "./api";
import type { DeleteScope, Occurrence, ScheduledItem } from "./api";
import { lengthOf, minuteOptionLabel, minuteOptions, monthlyHint, repeatOptions, toInput, whyWontPlay } from "./formRules";
import type { EditTarget, FormKind, FormValues, ScreenChoice } from "./formRules";
import { minuteLabel, toPlayable } from "./time";

/** What the caregiver chose to delete: one day of a repeating item, or the item with all its days */
export interface Removal {
  item: ScheduledItem;
  target: DeleteScope;
}

/** Where a new item would land, for the calendar's "New" placeholder */
export interface DraftSlot {
  date: string;
  minute: number;
}

interface ScheduleFormProps {
  carepartnerId: number;
  hours: AwakeHours;
  /** Photos, videos and audio; null while they load */
  library: Media[] | null;
  initial: FormValues;
  /** The item being edited; null for a new one */
  editing: EditTarget | null;
  onDraftChange: (draft: DraftSlot | null) => void;
  onClose: () => void;
  onSaved: (saved: ScheduledItem, wasEdit: boolean) => void;
  /** The page hides it and deletes it for real once its Undo toast closes */
  onRemove: (removal: Removal) => void;
  /** Occurrences deleted but still waiting on their Undo toast; they don't count as overlapping */
  isHidden: (occurrence: Occurrence) => boolean;
}

const KIND_SEGMENTS: { value: FormKind; label: string }[] = [
  { value: "video", label: "Video" },
  { value: "audio", label: "Audio" },
];

const SMALL_LINK =
  "text-sm font-bold text-dawn-accent hover:text-dawn-accent-hover focus-visible:outline focus-visible:outline-[3px] focus-visible:outline-offset-2 focus-visible:outline-dawn-accent";

const NOTE = "m-0 px-2.5 py-2 rounded-lg bg-[#FBEDE3] text-sm leading-snug text-dawn-accent-hover";

const DATE_PATTERN = /^\d{4}-\d{2}-\d{2}$/;

// Close and Delete are off while a save is under way, so a late answer can't close a form opened since
const ICON_DISABLED = "disabled:opacity-60 disabled:cursor-default disabled:hover:bg-transparent";

export default function ScheduleForm({
  carepartnerId,
  hours,
  library,
  initial,
  editing,
  onDraftChange,
  onClose,
  onSaved,
  onRemove,
  isHidden,
}: ScheduleFormProps) {
  const [values, setValues] = useState<FormValues>(initial);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const heading = useRef<HTMLHeadingElement>(null);
  // Asking "Delete which?" before anything is deleted
  const [confirmingDelete, setConfirmingDelete] = useState(false);
  const deleteButton = useRef<HTMLButtonElement>(null);
  const firstChoice = useRef<HTMLButtonElement>(null);

  useEffect(() => {
    heading.current?.focus();
  }, []);

  // Where a new item would land; an existing one is already on the calendar
  const dateIsValid = DATE_PATTERN.test(values.date);
  useEffect(() => {
    onDraftChange(editing === null && dateIsValid ? { date: values.date, minute: values.minute } : null);
  }, [editing, dateIsValid, values.date, values.minute, onDraftChange]);

  const patch = (change: Partial<FormValues>) => {
    setValues((current) => ({ ...current, ...change }));
    setError(null);
  };

  /* ============================
     Derived values
  ============================ */

  const ofKind = (library ?? []).filter((media) => media.kind === values.kind);
  const photos = (library ?? []).filter((media) => media.kind === "photo");
  // What was picked, or the first listed while nothing (valid) is picked yet
  const media = ofKind.find((item) => item.id === values.mediaId) ?? ofKind[0];
  const photo = photos.find((item) => item.id === values.photoId) ?? photos[0];
  const screen: ScreenChoice = photos.length === 0 ? "usual" : values.screen;
  const duration = lengthOf(media);

  // Labels follow the date last typed that was a whole date
  const labelDate = dateIsValid ? values.date : initial.date;
  const options = minuteOptions(hours, duration, values.minute);
  const wontPlay = whyWontPlay(values.minute, duration, hours);
  const monthly = monthlyHint(labelDate);
  const noun = values.kind === "audio" ? "audio" : "video";

  /* ============================
     Save
  ============================ */

  const save = async () => {
    if (!media || !dateIsValid) return;
    setSaving(true);
    setError(null);
    try {
      const chosen = { ...values, mediaId: media.id, screen, photoId: photo?.id ?? null };
      const input = toInput(chosen, carepartnerId, editing);

      // Nothing on the server stops two items overlapping, so look at that day first
      const day = (await getCalendar(carepartnerId, values.date, values.date)).filter((occurrence) => !isHidden(occurrence));
      const candidate: PlayableItem = {
        kind: values.kind === "video" ? "video" : input.background_photo_id === null ? "audio" : "audioWithPhoto",
        date: values.date,
        startMinute: values.minute,
        durationSeconds: duration,
      };
      const schedule = day.flatMap((occurrence) => {
        const playable = toPlayable(occurrence);
        return playable ? [{ ...playable, name: occurrence.media.display_name }] : [];
      });
      const clash = overlappingItems(candidate, schedule, editing?.item.id, hours)[0];
      if (clash) {
        setError(`That overlaps "${clash.name}" at ${minuteLabel(clash.startMinute)}. Pick another time.`);
        setSaving(false);
        return;
      }

      const saved = editing
        ? await updateScheduledItem(editing.item.id, input)
        : await createScheduledItem(input);
      onSaved(saved, editing !== null);
    } catch (err) {
      setError(friendlyError(err));
      setSaving(false);
    }
  };

  const repeats = editing !== null && editing.item.repeat !== "once";

  const remove = (scope: DeleteScope["scope"]) => {
    if (!editing) return;
    const target: DeleteScope = scope === "series" ? { scope } : { scope, date: editing.occurrenceDate };
    onRemove({ item: editing.item, target });
  };

  // A one-time item has nothing to choose between, so it goes straight to the Undo toast
  const askDelete = () => {
    setError(null);
    // For a repeating item the bin opens the question and pressing it again closes it
    if (repeats) setConfirmingDelete((open) => !open);
    else remove("series");
  };

  const cancelDelete = () => {
    setConfirmingDelete(false);
    setError(null);
    deleteButton.current?.focus();
  };

  // Focus moves into the question when it appears
  useEffect(() => {
    if (confirmingDelete) firstChoice.current?.focus();
  }, [confirmingDelete]);

  /* ============================
     Render
  ============================ */

  return (
    <section
      aria-label={editing ? "Edit a scheduled video or audio" : "Schedule a video or audio"}
      onKeyDown={(event) => {
        // A save under way finishes before the form can close
        if (event.key !== "Escape" || saving) return;
        // Escape backs out of the delete question first
        if (confirmingDelete) cancelDelete();
        else onClose();
      }}
      className="w-full lg:w-[290px] shrink-0 lg:overflow-y-auto flex flex-col gap-3 p-3.5 bg-white border-[1.5px] border-dawn-line rounded-[14px]"
    >
      <div className="flex items-center gap-0.5">
        <h2 ref={heading} tabIndex={-1} className="m-0 grow text-lg font-extrabold outline-none">
          {editing ? "Edit" : "Schedule"}
        </h2>
        {editing && (
          <IconButton
            ref={deleteButton}
            label="Delete"
            icon={<Trash2 size={18} aria-hidden="true" />}
            onClick={askDelete}
            disabled={saving}
            className={ICON_DISABLED}
            aria-expanded={repeats ? confirmingDelete : undefined}
          />
        )}
        <IconButton
          label="Close"
          icon={<X size={18} strokeWidth={2.4} aria-hidden="true" />}
          onClick={onClose}
          disabled={saving}
          className={ICON_DISABLED}
        />
      </div>

      {confirmingDelete && (
        <div role="group" aria-labelledby="schedule-delete-question" className="flex flex-col gap-2 p-2.5 rounded-[10px] bg-[#FBEDE3] border-[1.5px] border-[#E7C3AA]">
          <span id="schedule-delete-question" className="text-[15px] font-bold">
            Delete which?
          </span>
          <div className="flex flex-wrap gap-1.5">
            <button
              ref={firstChoice}
              type="button"
              onClick={() => remove("occurrence")}
              className={`h-10 px-3 rounded-[10px] border-2 border-dawn-control bg-white text-sm font-bold text-dawn-ink cursor-pointer ${FOCUS_RING}`}
            >
              Just this day
            </button>
            <button
              type="button"
              onClick={() => remove("series")}
              className={`h-10 px-3 rounded-[10px] border-0 bg-dawn-accent hover:bg-dawn-accent-hover text-sm font-bold text-white cursor-pointer ${FOCUS_RING}`}
            >
              Every time
            </button>
          </div>
        </div>
      )}

      {wontPlay && <p className={NOTE}>{wontPlay}</p>}

      <SegmentedControl
        label="Video or audio"
        segments={KIND_SEGMENTS}
        value={values.kind}
        onChange={(kind) => patch({ kind, mediaId: null })}
      />

      <TextField
        label="Date"
        type="date"
        value={values.date}
        onChange={(event) => patch({ date: event.target.value })}
      />

      <SelectField
        label="Time"
        value={formatClock(values.minute) ?? ""}
        onChange={(event) => {
          const [hour, minute] = event.target.value.split(":").map(Number);
          patch({ minute: hour * 60 + minute });
        }}
        options={options.map((minute) => ({
          value: formatClock(minute) ?? "",
          label: minuteOptionLabel(minute, hours, duration),
        }))}
      />

      <div className="flex flex-col gap-1.5">
        <div className="flex items-center justify-between">
          <span id="schedule-which" className="text-base font-bold">
            Which {noun}
          </span>
          <Link to="/media" className={SMALL_LINK}>
            Add new
          </Link>
        </div>
        {library === null ? (
          <p role="status" className="m-0 text-sm text-dawn-muted">
            Loading…
          </p>
        ) : ofKind.length === 0 ? (
          <p className="m-0 text-sm text-dawn-muted">No {noun} in the library yet. Add one, then come back.</p>
        ) : (
          <RadioGroup
            aria-labelledby="schedule-which"
            value={media ? String(media.id) : ""}
            onValueChange={(id) => patch({ mediaId: Number(id) })}
          >
            {ofKind.map((item) => (
              <RadioOption key={item.id} value={String(item.id)}>
                <span className="grow min-w-0 truncate text-[15px] font-normal">{item.display_name}</span>
                <span className="text-[13px] font-normal text-dawn-muted">
                  {item.duration_seconds === null ? "" : formatDuration(item.duration_seconds)}
                </span>
              </RadioOption>
            ))}
          </RadioGroup>
        )}
      </div>

      <SelectField
        label="Repeats"
        value={values.repeat}
        onChange={(event) => patch({ repeat: event.target.value as FormValues["repeat"] })}
        options={repeatOptions(labelDate)}
      />
      {monthly && values.repeat === "monthly" && <p className="m-0 -mt-1.5 text-sm text-dawn-muted">{monthly}</p>}

      {values.kind === "audio" && (
        <>
          <div className="flex flex-col gap-1.5">
            <span id="schedule-show" className="text-base font-bold">
              Display a photo?
            </span>
            <RadioGroup
              aria-labelledby="schedule-show"
              value={screen}
              onValueChange={(next) => patch({ screen: next as ScreenChoice })}
            >
              <RadioOption value="photo" disabled={photos.length === 0}>
                <span className="text-[15px] font-normal">Yes</span>
              </RadioOption>
              <RadioOption value="usual">
                <span className="text-[15px] font-normal">No</span>
              </RadioOption>
            </RadioGroup>
            {photos.length === 0 && (
              <p className="m-0 text-sm text-dawn-muted">
                Add a photo to the library to show one.{" "}
                <Link to="/media" className={SMALL_LINK}>
                  Add new
                </Link>
              </p>
            )}
          </div>

          {screen === "photo" && photo && (
            <div className="flex flex-col gap-1.5">
              <div className="flex items-center justify-between">
                <span className="text-base font-bold">Which photo</span>
                <Link to="/media" className={SMALL_LINK}>
                  Add new
                </Link>
              </div>
              <PhotoPicker
                label="Which photo"
                photos={photos.map((item) => ({ id: String(item.id), label: item.display_name, src: fileUrl(item) }))}
                value={String(photo.id)}
                onChange={(id) => patch({ photoId: Number(id) })}
              />
            </div>
          )}
        </>
      )}

      <p className="m-0 text-[13px] leading-snug text-dawn-muted">
        {values.kind === "audio"
          ? "Plays once at this time."
          : "Plays full screen once at this time, then the carepartner's usual screen comes back."}{" "}
        Times are in 15-minute steps. It has to start after {minuteLabel(hours.wakeMinute)} and finish by{" "}
        {minuteLabel(hours.bedMinute)}.
      </p>

      {error && (
        <p role="alert" className="m-0 text-sm font-bold text-dawn-accent">
          {error}
        </p>
      )}

      <div className="flex justify-end gap-2">
        <button type="button" onClick={onClose} disabled={saving} className={SECONDARY_BUTTON}>
          Cancel
        </button>
        <button type="button" onClick={save} disabled={saving || !media || !dateIsValid} className={PRIMARY_BUTTON}>
          {saving ? "Saving…" : "Save"}
        </button>
      </div>
    </section>
  );
}
