/**
 * SegmentedControl.tsx
 *
 * User: Caregiver
 *
 * A row of equal-width buttons for switching between a few modes
 * (e.g. Video / Photo). Exactly one segment is always chosen, so it is
 * built on Radix RadioGroup like RadioGroup and PhotoPicker.
 */

import * as RadixRadioGroup from "@radix-ui/react-radio-group";
import { FOCUS_RING } from "../styles";

export interface Segment<T extends string> {
  value: T;
  label: string;
}

interface SegmentedControlProps<T extends string> {
  segments: Segment<T>[];
  value: T;
  onChange: (value: T) => void;
  /** Names the group for screen readers */
  label: string;
}

export default function SegmentedControl<T extends string>({
  segments,
  value,
  onChange,
  label,
}: SegmentedControlProps<T>) {
  return (
    <RadixRadioGroup.Root
      aria-label={label}
      value={value}
      onValueChange={(next) => onChange(next as T)}
      className="flex gap-2"
    >
      {segments.map((segment) => (
        <RadixRadioGroup.Item
          key={segment.value}
          value={segment.value}
          className={`flex-1 h-11 px-3 rounded-[10px] border-2 text-base font-bold cursor-pointer bg-surface text-ink border-ink-muted data-[state=checked]:bg-accent data-[state=checked]:text-on-accent data-[state=checked]:border-accent ${FOCUS_RING}`}
        >
          {segment.label}
        </RadixRadioGroup.Item>
      ))}
    </RadixRadioGroup.Root>
  );
}
