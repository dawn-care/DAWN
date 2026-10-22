/**
 * ToggleButton.tsx
 *
 * User: Caregiver
 *
 * A 44px on/off button for filters and choices. The state is exposed with
 * aria-pressed, so use it for toggles; use RadioGroup when exactly one
 * option must be chosen from a set.
 */

import type { ButtonHTMLAttributes } from "react";
import { FOCUS_RING } from "../styles";

interface ToggleButtonProps extends Omit<ButtonHTMLAttributes<HTMLButtonElement>, "aria-pressed"> {
  pressed: boolean;
  /** "pill" is fully rounded; "square" matches the segmented control */
  shape?: "pill" | "square";
}

export default function ToggleButton({
  pressed,
  shape = "pill",
  className = "",
  type = "button",
  ...buttonProps
}: ToggleButtonProps) {
  const radius = shape === "pill" ? "rounded-full px-4" : "rounded-[10px] px-3";
  const tone = pressed
    ? "bg-accent text-on-accent border-accent"
    : "bg-surface text-ink border-ink-muted";

  return (
    <button
      type={type}
      aria-pressed={pressed}
      className={`h-11 border-2 text-base font-bold cursor-pointer ${radius} ${tone} ${FOCUS_RING} ${className}`}
      {...buttonProps}
    />
  );
}
