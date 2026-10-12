/**
 * NavBar.tsx
 *
 * User: Caregiver
 *
 * Left navigation bar shown on every page inside AppShell.
 *
 * Responsibilities:
 *  - Show the D.A.W.N. branding
 *  - Show the active carepartner's initials avatar and name
 *  - Link to Carepartner Screen, Calendar, and Photos, videos & audio
 *  - Show the signed-in caregiver's name
 *
 * /family exists as a stub route but is intentionally not linked here yet.
 */

import type { ReactNode } from "react";
import { NavLink } from "react-router-dom";
import { useAuth } from "../context/AuthContext";

/* ============================
   Static Data
============================ */

// TODO: hardcoded until the carepartner's facility is stored.
const FACILITY = "St. Ann's";

/**
 * Shared SVG attributes for the nav item icons
 */
const iconProps = {
  width: 22,
  height: 22,
  viewBox: "0 0 24 24",
  fill: "none",
  stroke: "currentColor",
  strokeWidth: 2,
  strokeLinecap: "round" as const,
  strokeLinejoin: "round" as const,
  "aria-hidden": true,
};

/**
 * Nav items, in display order
 */
const NAV_ITEMS: { to: string; label: string; end?: boolean; icon: ReactNode }[] = [
  {
    to: "/",
    label: "Carepartner Screen",
    end: true, // otherwise "/" would match every route
    icon: (
      <svg {...iconProps}>
        <rect x="3" y="4" width="18" height="12" rx="2" />
        <path d="M8 20h8M12 16v4" />
      </svg>
    ),
  },
  {
    to: "/calendar",
    label: "Calendar",
    icon: (
      <svg {...iconProps}>
        <rect x="3" y="5" width="18" height="16" rx="2" />
        <path d="M3 10h18M8 3v4M16 3v4" />
      </svg>
    ),
  },
  {
    to: "/media",
    label: "Photos, videos & audio",
    icon: (
      <svg {...iconProps}>
        <rect x="3" y="5" width="18" height="14" rx="2" />
        <circle cx="9" cy="10" r="1.5" />
        <path d="m21 16-5-5-9 8" />
      </svg>
    ),
  },
];

/* ============================
   Helpers
============================ */

/**
 * Builds avatar initials from a name: the first letter of up to
 * the first two words ("Jane Doe" → "JD")
 */
function getInitials(name: string): string {
  return name
    .split(/\s+/)
    .filter(Boolean)
    .slice(0, 2)
    .map((part) => part[0].toUpperCase())
    .join("");
}

/* ============================
   Component
============================ */

/**
 * NavBar Component
 *
 * Displays the branding, active carepartner, page links,
 * and signed-in caregiver for the app shell.
 */
export default function NavBar() {

  /* ============================
     Context
  ============================ */

  const { user, activeCarePartnerId } = useAuth();

  /* ============================
     Derived Values
  ============================ */

  const carepartner = user?.carepartners?.find((cp) => cp.id === activeCarePartnerId);
  const carepartnerName = carepartner?.name ?? "";

  /* ============================
     Render
  ============================ */

  return (
    <nav
      aria-label="Main"
      className="w-1/5 min-w-[220px] max-w-[320px] shrink-0 flex flex-col gap-2.5 px-4 py-7 bg-dawn-sidebar border-r border-dawn-line"
    >
      <div>
        <div className="text-2xl font-extrabold">D.A.W.N.</div>
        <div className="text-base text-dawn-muted">Caregiver</div>
      </div>

      <div className="flex items-center gap-3 mt-3.5 mb-2.5">
        <div
          aria-hidden="true"
          className="w-12 h-12 shrink-0 rounded-full bg-dawn-avatar text-dawn-avatar-ink flex items-center justify-center text-lg font-bold"
        >
          {getInitials(carepartnerName)}
        </div>
        <div className="flex flex-col min-w-0">
          <div className="text-[17px] font-bold truncate">{carepartnerName}</div>
          <div className="text-base text-dawn-muted">{FACILITY}</div>
        </div>
      </div>

      {NAV_ITEMS.map((item) => (
        <NavLink
          key={item.to}
          to={item.to}
          end={item.end}
          className={({ isActive }) =>
            `flex items-center gap-2.5 h-12 px-2.5 rounded-[10px] whitespace-nowrap text-base border-[1.5px] ${
              isActive
                ? "bg-white border-dawn-line-strong font-bold"
                : "border-transparent hover:bg-white/60"
            }`
          }
        >
          {item.icon}
          {item.label}
        </NavLink>
      ))}

      <div className="grow" />
      <div className="text-base text-dawn-muted">
        {user?.name ? `Signed in as ${user.name}` : "Signed in"}
      </div>
    </nav>
  );
}
