/**
 * SaveIndicator.tsx
 *
 * User: Caregiver
 *
 * The "Changes save automatically" note shown in page headers. It is
 * static reassurance text; it does not track save state.
 */

import { Check } from "lucide-react";

export default function SaveIndicator() {
  return (
    <div className="flex items-center gap-1.5 text-base font-bold text-dawn-success">
      <Check size={20} strokeWidth={3} aria-hidden="true" />
      Changes save automatically
    </div>
  );
}
