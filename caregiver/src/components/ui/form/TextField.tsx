/**
 * TextField.tsx
 *
 * User: Caregiver
 *
 * A labeled single-line text input. The label is linked to the input by
 * id. The optional hint sits outside the label, so it is read once, as
 * the input's description, and not as part of its name.
 */

import { forwardRef, useId } from "react";
import type { InputHTMLAttributes } from "react";
import { FIELD_CLASSES } from "../styles";

interface TextFieldProps extends InputHTMLAttributes<HTMLInputElement> {
  label: string;
  /** Short helper text shown under the label */
  hint?: string;
}

const TextField = forwardRef<HTMLInputElement, TextFieldProps>(function TextField(
  { label, hint, className = "", id, "aria-describedby": describedBy, ...inputProps },
  ref
) {
  const generatedId = useId();
  const inputId = id ?? generatedId;
  const hintId = `${inputId}-hint`;
  const description = [hint ? hintId : undefined, describedBy].filter(Boolean).join(" ") || undefined;

  return (
    <div className="flex flex-col gap-1.5">
      <label htmlFor={inputId} className="text-base font-bold text-ink">
        {label}
      </label>
      {hint && (
        <span id={hintId} className="text-[15px] font-normal text-ink-muted">
          {hint}
        </span>
      )}
      <input
        ref={ref}
        id={inputId}
        type="text"
        aria-describedby={description}
        className={`${FIELD_CLASSES} font-normal ${className}`}
        {...inputProps}
      />
    </div>
  );
});

export default TextField;
