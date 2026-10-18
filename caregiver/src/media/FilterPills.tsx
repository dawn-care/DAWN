/**
 * FilterPills.tsx
 *
 * User: Caregiver
 *
 * The All / Photos / Videos / Audio filter above the media grid, each with
 * a count. Exactly one is always chosen, so it is a Radix RadioGroup (arrow
 * keys move between pills), like SegmentedControl but with rounded pills
 * that size to their label.
 */

import * as RadixRadioGroup from "@radix-ui/react-radio-group";
import { FOCUS_RING } from "../components/ui/styles";

export interface Pill<T extends string> {
  value: T;
  label: string;
  count: number;
}

interface FilterPillsProps<T extends string> {
  pills: Pill<T>[];
  value: T;
  onChange: (value: T) => void;
  /** Names the group for screen readers */
  label: string;
}

export default function FilterPills<T extends string>({ pills, value, onChange, label }: FilterPillsProps<T>) {
  return (
    <RadixRadioGroup.Root
      aria-label={label}
      value={value}
      onValueChange={(next) => onChange(next as T)}
      className="flex flex-wrap gap-1.5"
    >
      {pills.map((pill) => (
        <RadixRadioGroup.Item
          key={pill.value}
          value={pill.value}
          className={`h-11 px-4 rounded-full border-2 text-base font-bold cursor-pointer bg-white text-dawn-ink border-dawn-control data-[state=checked]:bg-dawn-ink data-[state=checked]:text-white data-[state=checked]:border-dawn-ink ${FOCUS_RING}`}
        >
          {pill.label} ({pill.count})
        </RadixRadioGroup.Item>
      ))}
    </RadixRadioGroup.Root>
  );
}
