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
 * Saving first checks the day for another item at the same time (the server
 * leaves overlap to this page), then creates the item or replaces the one
 * being edited. Why a time can't be saved comes back in words, next to the
 * form.
 */

import { formatClock, overlappingItems } from "@shared";
import type { AwakeHours, PlayableItem } from "@shared";
import { X } from "lucide-react";
import { useEffect, useRef, useState } from "react";
import { Link } from "react-router-dom";
import IconButton from "../components/ui/IconButton";
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
import type { ScheduledItem } from "./api";
import { lengthOf, minuteOptionLabel, minuteOptions, monthlyHint, repeatOptions, toInput, whyWontPlay } from "./formRules";
import type { FormKind, FormValues, ScreenChoice } from "./formRules";
import { minuteLabel, toPlayable } from "./time";

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
  /** The item being edited, and the date of the occurrence it was opened from; null for a new one */
  editing: { item: ScheduledItem; occurrenceDate: string } | null;
  onDraftChange: (draft: DraftSlot | null) => void;
  onClose: () => void;
  onSaved: (saved: ScheduledItem, wasEdit: boolean) => void;
}

const KIND_SEGMENTS: { value: FormKind; label: string }[] = [
  { value: "video", label: "Video" },
  { value: "audio", label: "Audio" },
];

const SMALL_LINK =
  "text-sm font-bold text-dawn-accent hover:text-dawn-accent-hover focus-visible:outline focus-visible:outline-[3px] focus-visible:outline-offset-2 focus-visible:outline-dawn-accent";

const NOTE = "m-0 px-2.5 py-2 rounded-lg bg-[#FBEDE3] text-sm leading-snug text-dawn-accent-hover";

const DATE_PATTERN = /^\d{4}-\d{2}-\d{2}$/;

export default function ScheduleForm({ carepartnerId, hours, library, initial, editing, onDraftChange, onClose, onSaved }: ScheduleFormProps) {
  const [values, setValues] = useState<FormValues>(initial);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const heading = useRef<HTMLHeadingElement>(null);

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
      const chosen: FormValues = { ...values, mediaId: media.id, screen, photoId: photo?.id ?? null };
      const input = toInput(chosen, carepartnerId, editing);

      // Nothing on the server stops two items starting together, so look at that day first
      const day = await getCalendar(carepartnerId, values.date, values.date);
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

  /* ============================
     Render
  ============================ */

  return (
    <section
      aria-label={editing ? "Edit a scheduled video or audio" : "Schedule a video or audio"}
      onKeyDown={(event) => {
        if (event.key === "Escape") onClose();
      }}
      className="w-full lg:w-[290px] shrink-0 lg:overflow-y-auto flex flex-col gap-3 p-3.5 bg-white border-[1.5px] border-dawn-line rounded-[14px]"
    >
      <div className="flex items-center gap-0.5">
        <h2 ref={heading} tabIndex={-1} className="m-0 grow text-lg font-extrabold outline-none">
          {editing ? "Edit" : "Schedule"}
        </h2>
        <IconButton label="Close" icon={<X size={18} strokeWidth={2.4} aria-hidden="true" />} onClick={onClose} />
      </div>

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
        <button type="button" onClick={onClose} className={SECONDARY_BUTTON}>
          Cancel
        </button>
        <button type="button" onClick={save} disabled={saving || !media || !dateIsValid} className={PRIMARY_BUTTON}>
          {saving ? "Saving…" : "Save"}
        </button>
      </div>
    </section>
  );
}
