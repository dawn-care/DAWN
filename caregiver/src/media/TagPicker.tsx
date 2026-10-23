/**
 * TagPicker.tsx
 *
 * User: Caregiver
 *
 * Chooses how staff find a media item:
 *
 *  - When to use it: the situations staff pick from on their phone.
 *  - Who it's about: the carepartner's family members, so staff find it when
 *    the carepartner asks for that person.
 *
 * Nothing is chosen to start with. With neither chosen, staff never see the
 * item, which a note under the picker says.
 *
 * Someone with a guidance note can't be chosen for now: staff see the note
 * instead of media for them (the staff app's family screen).
 *
 * Both are rows of toggle buttons. "Add family member" opens a small inline
 * form (name and relationship); the new person is created right away and
 * chosen. Guidance notes are written elsewhere, later.
 *
 * Controlled: the caller keeps the chosen values and the family list.
 */

import { useId, useState } from "react";
import type { SyntheticEvent } from "react";
import { Plus } from "lucide-react";
import ToggleButton from "../components/ui/form/ToggleButton";
import TextField from "../components/ui/form/TextField";
import { createFamilyMember, friendlyError } from "./api";
import type { FamilyMember } from "./api";
import { SITUATIONS } from "./constants";
import type { Situation } from "./constants";
import { PRIMARY_BUTTON, SECONDARY_BUTTON } from "./styles";
import { FOCUS_RING } from "../components/ui/styles";

interface TagPickerProps {
  carepartnerId: number;
  situations: Situation[];
  familyMemberIds: number[];
  /** Null if the family couldn't be loaded */
  family: FamilyMember[] | null;
  onSituationsChange: (situations: Situation[]) => void;
  onFamilyMemberIdsChange: (ids: number[]) => void;
  /** A family member was added from here; the caller adds them to its list */
  onFamilyAdded: (member: FamilyMember) => void;
  disabled?: boolean;
}

const HEADING = "m-0 text-base font-bold text-ink";
const HINT = "m-0 text-[15px] text-ink-muted";

// Adds `value` to the list, or takes it out if it's there
function toggled<T>(list: T[], value: T): T[] {
  return list.includes(value) ? list.filter((item) => item !== value) : [...list, value];
}

export default function TagPicker({
  carepartnerId,
  situations,
  familyMemberIds,
  family,
  onSituationsChange,
  onFamilyMemberIdsChange,
  onFamilyAdded,
  disabled = false,
}: TagPickerProps) {
  const situationsId = useId();
  const familyId = useId();

  const [adding, setAdding] = useState(false);
  const [name, setName] = useState("");
  const [relationship, setRelationship] = useState("");
  const [saving, setSaving] = useState(false);
  const [addError, setAddError] = useState<string | null>(null);

  const openAdd = () => {
    setName("");
    setRelationship("");
    setAddError(null);
    setAdding(true);
  };

  const addMember = async (event: SyntheticEvent) => {
    // The picker sits inside the Save panel's form; Enter here must not save the file
    event.preventDefault();
    event.stopPropagation();
    if (name.trim() === "") {
      setAddError("Enter their name.");
      return;
    }
    setSaving(true);
    setAddError(null);
    let member: FamilyMember;
    try {
      member = await createFamilyMember(carepartnerId, name, relationship);
    } catch (err) {
      setAddError(friendlyError(err));
      setSaving(false);
      return;
    }
    onFamilyAdded(member);
    onFamilyMemberIdsChange([...familyMemberIds, member.id]);
    setSaving(false);
    setAdding(false);
  };

  return (
    <div className="flex flex-col gap-4">
      <div role="group" aria-labelledby={situationsId} className="flex flex-col gap-2">
        <div className="flex flex-col gap-0.5">
          <h3 id={situationsId} className={HEADING}>
            When should staff use it?
          </h3>
          <p className={HINT}>Staff see it when they pick one of these.</p>
        </div>
        <div className="flex flex-wrap gap-1.5">
          {SITUATIONS.map(({ value, label }) => (
            <ToggleButton
              key={value}
              pressed={situations.includes(value)}
              disabled={disabled}
              onClick={() => onSituationsChange(toggled(situations, value))}
            >
              {label}
            </ToggleButton>
          ))}
        </div>
      </div>

      <div role="group" aria-labelledby={familyId} className="flex flex-col gap-2">
        <div className="flex flex-col gap-0.5">
          <h3 id={familyId} className={HEADING}>
            Who is it about?
          </h3>
          <p className={HINT}>Staff see it when they're asked for this person.</p>
          {family?.some((member) => member.guidance_note !== null) && (
            <p className={HINT}>Greyed out: staff see your guidance note for them instead of media.</p>
          )}
        </div>

        {family === null ? (
          <p className={HINT}>Couldn't load the family. You can still save, and add them later.</p>
        ) : (
          <div className="flex flex-wrap gap-1.5">
            {family.map((member) => {
              const hasNote = member.guidance_note !== null;
              return (
                <ToggleButton
                  key={member.id}
                  pressed={familyMemberIds.includes(member.id)}
                  disabled={disabled || hasNote}
                  title={hasNote ? "Staff see your guidance note for them instead of media" : undefined}
                  onClick={() => onFamilyMemberIdsChange(toggled(familyMemberIds, member.id))}
                  className={hasNote ? "opacity-50 cursor-not-allowed" : ""}
                >
                  {member.name}
                  {member.relationship && <span className="font-normal"> · {member.relationship}</span>}
                </ToggleButton>
              );
            })}
            {!adding && (
              <button
                type="button"
                onClick={openAdd}
                disabled={disabled}
                className={`h-11 px-4 flex items-center gap-1.5 rounded-full border-2 border-dashed border-ink-muted bg-surface text-base font-bold text-ink cursor-pointer disabled:opacity-60 disabled:cursor-default ${FOCUS_RING}`}
              >
                <Plus size={16} strokeWidth={2.6} aria-hidden="true" />
                Add family member
              </button>
            )}
          </div>
        )}

        {adding && (
          // Not a <form>: this sits inside the Save panel's form, and forms can't nest
          <div
            className="flex flex-col gap-2 p-3 rounded-xl border-2 border-border"
            onKeyDown={(event) => {
              if (event.key === "Enter") void addMember(event);
              if (event.key === "Escape" && !saving) setAdding(false);
            }}
          >
            <TextField
              label="Name"
              value={name}
              onChange={(event) => setName(event.target.value)}
              disabled={saving}
              autoFocus
              aria-invalid={addError ? true : undefined}
            />
            <TextField
              label="Relationship"
              hint="Optional, e.g. Daughter"
              value={relationship}
              onChange={(event) => setRelationship(event.target.value)}
              disabled={saving}
            />
            {addError && (
              <p role="alert" className="m-0 text-base font-bold text-link">
                {addError}
              </p>
            )}
            <div className="flex gap-2">
              <button type="button" onClick={(event) => void addMember(event)} disabled={saving} className={`flex-1 ${PRIMARY_BUTTON}`}>
                {saving ? "Adding…" : "Add"}
              </button>
              <button type="button" onClick={() => setAdding(false)} disabled={saving} className={`flex-1 ${SECONDARY_BUTTON}`}>
                Cancel
              </button>
            </div>
          </div>
        )}
      </div>

      {situations.length === 0 && familyMemberIds.length === 0 && (
        <p role="status" className="m-0 px-2.5 py-2 rounded-lg bg-border text-[15px] font-bold text-ink">
          Nothing picked, so staff won't see this in their app.
        </p>
      )}
    </div>
  );
}
