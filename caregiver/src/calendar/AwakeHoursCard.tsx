/**
 * AwakeHoursCard.tsx
 *
 * User: Caregiver
 *
 * The "Awake hours" card in the calendar's left column: when the carepartner
 * wakes up and goes to bed, from the screen settings. Videos and audio only
 * play between them. "Change" opens the editor in place: Wakes up and Bedtime
 * selects, a warning naming every scheduled item still to play that would be
 * outside the chosen hours (including any already outside the saved ones),
 * and Done to save (Cancel throws the changes away).
 */

import { formatClock, parseClock, timeLabel } from "@shared";
import { useEffect, useRef, useState } from "react";
import SelectField from "../components/ui/form/SelectField";
import { FOCUS_RING } from "../components/ui/styles";
import { friendlyError, listScheduledItems, saveScreenSettings } from "./api";
import type { ScheduledItem, ScreenSettings } from "./api";
import { BED_RANGE, WAKE_RANGE } from "./constants";
import { awakeHoursOf, awakeOptions, dateText, itemsThatWontPlay, nameList } from "./time";

interface AwakeHoursCardProps {
  carepartnerId: number;
  /** Null until the settings have loaded (or if they couldn't be) */
  settings: ScreenSettings | null;
  editing: boolean;
  onEdit: () => void;
  onClose: () => void;
  /** The hours were saved; `saved` is what the server now has */
  onSaved: (saved: ScreenSettings) => void;
}

const TEXT_BUTTON = `h-9 px-1 bg-transparent text-[15px] font-bold cursor-pointer disabled:opacity-60 disabled:cursor-default ${FOCUS_RING}`;

// The options for one select, with the saved "HH:MM" always among them
function optionsFor(range: { earliest: number; latest: number }, saved: string) {
  const savedMinute = parseClock(saved) ?? range.earliest;
  return awakeOptions(range, savedMinute).map((minute) => {
    const value = formatClock(minute) ?? "";
    return { value, label: timeLabel(value) };
  });
}

export default function AwakeHoursCard({ carepartnerId, settings, editing, onEdit, onClose, onSaved }: AwakeHoursCardProps) {
  const changeButton = useRef<HTMLButtonElement>(null);
  const wasEditing = useRef(false);

  // Focus goes back to Change once the editor closes
  useEffect(() => {
    if (wasEditing.current && !editing) changeButton.current?.focus();
    wasEditing.current = editing;
  }, [editing]);

  const isEditing = editing && settings !== null;

  return (
    <section
      aria-label="Awake hours"
      className={`flex flex-col gap-1.5 p-3 bg-surface border-[1.5px] rounded-xl ${isEditing ? "border-accent" : "border-border"}`}
    >
      <div className="flex items-center justify-between">
        <h2 className="m-0 text-[15px] font-extrabold">Awake hours</h2>
        {!isEditing && settings && (
          <button ref={changeButton} type="button" onClick={onEdit} className={`${TEXT_BUTTON} text-link`}>
            Change
          </button>
        )}
      </div>
      {isEditing ? (
        <AwakeHoursEditor carepartnerId={carepartnerId} settings={settings} onCancel={onClose} onSaved={onSaved} />
      ) : (
        <>
          <p className="m-0 text-[17px] font-bold">
            {settings ? `${timeLabel(settings.wake_time)} – ${timeLabel(settings.bed_time)}` : "—"}
          </p>
          <p className="m-0 text-sm leading-snug text-ink-muted">Videos and audio only play while the carepartner is awake.</p>
        </>
      )}
    </section>
  );
}

interface AwakeHoursEditorProps {
  carepartnerId: number;
  settings: ScreenSettings;
  onCancel: () => void;
  onSaved: (saved: ScreenSettings) => void;
}

// Mounted only while editing, so each opening starts from the saved hours
function AwakeHoursEditor({ carepartnerId, settings, onCancel, onSaved }: AwakeHoursEditorProps) {
  const [wake, setWake] = useState(settings.wake_time);
  const [bed, setBed] = useState(settings.bed_time);
  const [items, setItems] = useState<ScheduledItem[]>([]);
  // The schedule couldn't be read, so the warning can't say what the new hours affect
  const [checkFailed, setCheckFailed] = useState(false);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const wakeSelect = useRef<HTMLSelectElement>(null);

  useEffect(() => {
    wakeSelect.current?.focus();
  }, []);

  // The warning is advice only, so if the schedule can't be read the hours can still be saved
  useEffect(() => {
    let cancelled = false;
    listScheduledItems(carepartnerId)
      .then((list) => {
        if (!cancelled) setItems(list);
      })
      .catch((err) => {
        console.warn("Could not load the schedule for the awake-hours warning", err);
        if (!cancelled) setCheckFailed(true);
      });
    return () => {
      cancelled = true;
    };
  }, [carepartnerId]);

  const draft = awakeHoursOf({ ...settings, wake_time: wake, bed_time: bed });
  // One name per file, so "is" and "are" agree with the names shown
  const names = draft
    ? [...new Set(itemsThatWontPlay(items, draft, dateText(new Date())).map((item) => item.media.display_name))]
    : [];

  const save = async () => {
    setSaving(true);
    setError(null);
    try {
      onSaved(await saveScreenSettings(carepartnerId, wake, bed));
    } catch (err) {
      setError(friendlyError(err));
      setSaving(false);
    }
  };

  return (
    <div className="flex flex-col gap-2">
      <SelectField
        ref={wakeSelect}
        label="Wakes up"
        value={wake}
        onChange={(event) => setWake(event.target.value)}
        options={optionsFor(WAKE_RANGE, settings.wake_time)}
      />
      <SelectField
        label="Bedtime"
        value={bed}
        onChange={(event) => setBed(event.target.value)}
        options={optionsFor(BED_RANGE, settings.bed_time)}
      />
      {names.length > 0 && (
        <p role="status" className="m-0 px-2 py-1.5 rounded-lg bg-status-orange-bg text-[13px] leading-snug text-status-orange-text">
          {nameList(names)} {names.length > 1 ? "are" : "is"} outside these hours and won't play until you move{" "}
          {names.length > 1 ? "them" : "it"}.
        </p>
      )}
      {checkFailed && (
        <p role="status" className="m-0 text-[13px] leading-snug text-ink-muted">
          Couldn't check which videos and audio these hours affect.
        </p>
      )}
      {error && (
        <p role="alert" className="m-0 text-sm text-link">
          {error}
        </p>
      )}
      <div className="flex justify-end gap-2">
        <button type="button" onClick={onCancel} disabled={saving} className={`${TEXT_BUTTON} text-ink`}>
          Cancel
        </button>
        <button
          type="button"
          onClick={save}
          disabled={saving}
          className={`h-10 px-4 rounded-[10px] border-0 bg-accent hover:brightness-95 text-[15px] font-bold text-on-accent cursor-pointer disabled:opacity-60 disabled:cursor-default ${FOCUS_RING}`}
        >
          Done
        </button>
      </div>
    </div>
  );
}
