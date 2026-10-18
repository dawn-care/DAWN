/**
 * Preview.tsx
 *
 * User: Caregiver
 *
 * Full-screen preview of the carepartner's screen at `/preview`. Renders
 * outside AppShell, so there is no nav bar.
 *
 * FUTURE IMPLEMENTATION: the preview content and the button that opens this
 * page are added in a later ticket. Until then it is only reachable by
 * typing the URL.
 */

import { Link } from "react-router-dom";
import { X } from "lucide-react";

export default function Preview() {
  return (
    <div className="relative min-h-screen p-10 bg-dawn-preview text-dawn-ink font-dawn">
      <Link
        to="/"
        className="absolute right-6 top-6 inline-flex items-center gap-2 h-12 px-[18px] rounded-3xl bg-dawn-ink text-white text-base font-bold"
      >
        <X size={18} aria-hidden="true" />
        Exit preview
      </Link>
    </div>
  );
}
