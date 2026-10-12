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
}

const SelectField = forwardRef<HTMLSelectElement, SelectFieldProps>(function SelectField(
  { label, options, className = "", ...selectProps },
  ref
) {
  return (
    <label className="flex flex-col gap-1.5 text-base font-bold text-dawn-ink">
      {label}
      <select ref={ref} className={`${FIELD_CLASSES} font-normal ${className}`} {...selectProps}>
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
