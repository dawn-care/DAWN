/**
 * SelectField.tsx
 *
 * User: Caregiver
 *
 * A labeled dropdown built on the native <select>, which keeps the
 * browser's keyboard and screen reader behavior. Styled to match TextField.
 */

import { forwardRef } from "react";
import type { SelectHTMLAttributes } from "react";
import { FIELD_CLASSES } from "../styles";

export interface SelectOption {
  value: string;
  label: string;
}

interface SelectFieldProps extends SelectHTMLAttributes<HTMLSelectElement> {
  label: string;
  options: SelectOption[];
  /**
   * Text for a disabled empty first option, shown while nothing is chosen
   * (value="" or defaultValue=""). Without it, a browser shows the first
   * real option as if it were chosen while the value is still empty.
   */
  placeholder?: string;
}

const SelectField = forwardRef<HTMLSelectElement, SelectFieldProps>(function SelectField(
  { label, options, placeholder, className = "", ...selectProps },
  ref
) {
  return (
    <label className="flex flex-col gap-1.5 text-base font-bold text-ink">
      {label}
      <select ref={ref} className={`${FIELD_CLASSES} font-normal ${className}`} {...selectProps}>
        {placeholder && (
          <option value="" disabled>
            {placeholder}
          </option>
        )}
        {options.map((option) => (
          <option key={option.value} value={option.value}>
            {option.label}
          </option>
        ))}
      </select>
    </label>
  );
});

export default SelectField;
