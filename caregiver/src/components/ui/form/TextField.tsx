/**
 * TextField.tsx
 *
 * User: Caregiver
 *
 * A labeled single-line text input. The label wraps the input so the two
 * are associated without needing ids; the optional hint is read as the
 * input's description.
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
  { label, hint, className = "", ...inputProps },
  ref
) {
  const hintId = useId();

  return (
    <label className="flex flex-col gap-1.5 text-base font-bold text-dawn-ink">
      {label}
      {hint && (
        <span id={hintId} className="text-[15px] font-normal text-dawn-muted">
          {hint}
        </span>
      )}
      <input
        ref={ref}
        type="text"
        aria-describedby={hint ? hintId : undefined}
        className={`${FIELD_CLASSES} font-normal ${className}`}
        {...inputProps}
      />
    </label>
  );
});

export default TextField;
