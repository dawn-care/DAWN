/**
 * Switch.tsx
 *
 * User: Caregiver
 *
 * An on/off setting: a label with a sliding switch at the end of the row. The
 * whole row is the button, at least 44px tall, and screen readers announce it
 * as a switch that is on or off. Use ToggleButton for filters and choices that
 * read as buttons, and a Switch for a setting that is simply on or off.
 */

import type { ButtonHTMLAttributes } from "react";
import { FOCUS_RING } from "../styles";

interface SwitchProps extends Omit<ButtonHTMLAttributes<HTMLButtonElement>, "role" | "aria-checked" | "onClick" | "children"> {
  /** What the switch controls, as its name */
  label: string;
  checked: boolean;
  onCheckedChange: (checked: boolean) => void;
}

export default function Switch({ label, checked, onCheckedChange, className = "", type = "button", ...buttonProps }: SwitchProps) {
  const track = checked ? "bg-accent border-accent" : "bg-surface border-ink-muted";
  const knob = checked ? "bg-on-accent translate-x-5" : "bg-ink-muted translate-x-0";

  return (
    <button
      {...buttonProps}
      type={type}
      role="switch"
      aria-checked={checked}
      onClick={() => onCheckedChange(!checked)}
      className={`flex items-center justify-between gap-3 min-h-11 w-full rounded-[10px] text-left text-base font-bold text-ink cursor-pointer disabled:opacity-50 disabled:cursor-not-allowed ${FOCUS_RING} ${className}`}
    >
      {label}
      <span aria-hidden="true" className={`relative shrink-0 w-12 h-7 rounded-full border-2 ${track}`}>
        <span
          className={`absolute top-0.5 left-0.5 w-5 h-5 rounded-full transition-transform motion-reduce:transition-none ${knob}`}
        />
      </span>
    </button>
  );
}
