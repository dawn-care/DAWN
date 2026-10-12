/**
 * SegmentedControl.tsx
 *
 * User: Caregiver
 *
 * A row of equal-width buttons for switching between a few modes
 * (e.g. Video / Photo). Each segment is a ToggleButton, so the active
 * one is announced with aria-pressed.
 */

import ToggleButton from "./ToggleButton";

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
    <div role="group" aria-label={label} className="flex gap-2">
      {segments.map((segment) => (
        <ToggleButton
          key={segment.value}
          shape="square"
          pressed={segment.value === value}
          onClick={() => onChange(segment.value)}
          className="flex-1"
        >
          {segment.label}
        </ToggleButton>
      ))}
    </div>
  );
}
