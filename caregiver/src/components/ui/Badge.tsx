/**
 * Badge.tsx
 *
 * User: Caregiver
 *
 * A small label pill, e.g. the content type on a library card or a "New"
 * tag. "dark" is the neutral default; "accent" draws attention.
 */

import type { ReactNode } from "react";

interface BadgeProps {
  tone?: "dark" | "accent";
  children: ReactNode;
}

export default function Badge({ tone = "dark", children }: BadgeProps) {
  const toneClasses = tone === "accent" ? "bg-dawn-accent font-extrabold" : "bg-dawn-avatar-ink font-bold";

  return (
    <span className={`inline-block px-[7px] py-0.5 rounded-md text-xs text-white ${toneClasses}`}>
      {children}
    </span>
  );
}
