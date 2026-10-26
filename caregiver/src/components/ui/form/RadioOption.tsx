/**
 * RadioOption.tsx
 *
 * User: Caregiver
 *
 * One selectable row inside a RadioGroup: a 20px dot and a label, at least
 * 48px tall. The checked row gets a darker border and an accent dot.
 */

import * as RadixRadioGroup from "@radix-ui/react-radio-group";
import type { ReactNode } from "react";
import { FOCUS_RING } from "../styles";

interface RadioOptionProps {
  value: string;
  children: ReactNode;
  disabled?: boolean;
}

export default function RadioOption({ value, children, disabled }: RadioOptionProps) {
  return (
    <RadixRadioGroup.Item
      value={value}
      disabled={disabled}
      className={`group flex items-center gap-2.5 min-h-12 px-3 py-1.5 rounded-[10px] border-2 text-left text-base font-bold text-ink cursor-pointer bg-surface border-border data-[state=checked]:bg-page data-[state=checked]:border-ink disabled:bg-page disabled:border-border disabled:text-ink-muted/60 disabled:cursor-not-allowed ${FOCUS_RING}`}
    >
      <span
        aria-hidden="true"
        className="w-5 h-5 shrink-0 rounded-full bg-surface border-2 border-ink-muted group-data-[state=checked]:border-[6px] group-data-[state=checked]:border-accent group-disabled:bg-page group-disabled:border-border"
      />
      {children}
    </RadixRadioGroup.Item>
  );
}
