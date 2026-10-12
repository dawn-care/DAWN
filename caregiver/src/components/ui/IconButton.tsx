/**
 * IconButton.tsx
 *
 * User: Caregiver
 *
 * A 40px icon-only button. The label is required because there is no
 * visible text; it becomes the button's accessible name. Forwards its ref
 * so Radix triggers (asChild) can use it.
 */

import { forwardRef } from "react";
import type { ButtonHTMLAttributes, ReactNode } from "react";
import { FOCUS_RING } from "./styles";

interface IconButtonProps extends Omit<ButtonHTMLAttributes<HTMLButtonElement>, "aria-label"> {
  /** Accessible name, e.g. "Edit Garden photo" */
  label: string;
  /** The icon, usually a lucide-react icon with aria-hidden */
  icon: ReactNode;
}

const IconButton = forwardRef<HTMLButtonElement, IconButtonProps>(function IconButton(
  { label, icon, className = "", type = "button", ...buttonProps },
  ref
) {
  return (
    <button
      ref={ref}
      type={type}
      aria-label={label}
      className={`w-10 h-10 shrink-0 flex items-center justify-center rounded-full bg-transparent text-dawn-ink cursor-pointer hover:bg-dawn-line/60 ${FOCUS_RING} ${className}`}
      {...buttonProps}
    >
      {icon}
    </button>
  );
});

export default IconButton;
